package client

import (
	"context"
	"flag"
	"fmt"
	"kube-scheduler-practice/internal/logic"
	"log/slog"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type K8sClient struct {
	Clientset     kubernetes.Interface
	ScheduleLogic ScheduleLogic
}

type ScheduleLogic interface {
	ChooseAvailableNodes(unschedulePod *v1.Pod, vs *v1.NodeList) (*v1.NodeList, error)
	ChooseSuitableNode(unschedulePod *v1.Pod, vs *v1.NodeList) (v1.Node, error)
}

func NewLocalClient() (K8sClient, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return K8sClient{}, fmt.Errorf("error building kubeconfig: %s", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return K8sClient{}, fmt.Errorf("error creating clientset: %s", err.Error())
	}

	scheduleLogic := &logic.ScheduleLogic{}
	return K8sClient{Clientset: clientset, ScheduleLogic: scheduleLogic}, nil
}

func NewInClusterClient() (K8sClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return K8sClient{}, fmt.Errorf("error creating in-cluster config: %s", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return K8sClient{}, fmt.Errorf("error creating clientset: %s", err.Error())
	}

	scheduleLogic := &logic.ScheduleLogic{}
	return K8sClient{Clientset: clientset, ScheduleLogic: scheduleLogic}, nil
}

func (k *K8sClient) GetNodes() (*v1.NodeList, error) {
	nodes, err := k.Clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting nodes: %s", err.Error())
	}
	return nodes, nil
}

func (k *K8sClient) GetUnscheduledPods() (*v1.PodList, error) {
	// node にアサインされていない Pod の一覧を取得する

	// TODO: この実装はFieldSelectorを使うことで効率化される
	unscheduledPods := &v1.PodList{}
	pods, err := k.Clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting pods: %s", err.Error())
	}
	for _, pod := range pods.Items {
		if pod.Spec.NodeName == "" {
			unscheduledPods.Items = append(unscheduledPods.Items, pod)
			slog.Info("detect unscheduled pods", "name", pod.Name, "namespace", pod.Namespace)
		}
	}

	return unscheduledPods, nil
}

func (k *K8sClient) AssignPodToNode(pod *v1.Pod, node *v1.Node) error {
	binding := &v1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		Target: v1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       node.Name,
		},
	}

	slog.Info("attempting to bind pod to node", "pod", pod.Name, "node", node.Name)

	err := k.Clientset.CoreV1().Pods(pod.Namespace).Bind(context.TODO(), binding, metav1.CreateOptions{})
	if err != nil {
		slog.Error("failed to bind pod to node", "pod", pod.Name, "node", node.Name, "error", err)
		return fmt.Errorf("failed to bind pod %s/%s to node %s: %w", pod.Namespace, pod.Name, node.Name, err)
	}

	return nil
}

// スケジュールされていない pod 取得 → ノード情報取得 → 配置するpodを選択 → 配置指示
// 一連の処理の一巡を行う
func (k *K8sClient) ProcessOneLoop() error {
	unscheduledPods, err := k.GetUnscheduledPods()
	if err != nil {
		return err
	}

	for _, pod := range unscheduledPods.Items {
		nodes, err := k.GetNodes()
		if err != nil {
			return err
		}

		// 配置して良いノードを取得
		availableNodes, err := k.ScheduleLogic.ChooseAvailableNodes(&pod, nodes)
		if err != nil {
			return err
		}
		// 実際に配置するノードを取得
		selectNode, err := k.ScheduleLogic.ChooseSuitableNode(&pod, availableNodes)
		if err != nil {
			return err
		}

		// もし selectNode が空だったら、スケジューリングをスキップ
		if selectNode.Name == "" {
			slog.Info("no suitable node found for pod", "pod", pod.Name)
			continue
		}

		if err := k.AssignPodToNode(&pod, &selectNode); err != nil {
			return err
		}

		slog.Info("assign pod to node successfully", "pod", pod.Name, "node", selectNode.Name)
	}
	return nil
}

func (k *K8sClient) Run() {
	for {
		if err := k.ProcessOneLoop(); err != nil {
			slog.Error(err.Error())
		}
		time.Sleep(10 * time.Second)
	}
}
