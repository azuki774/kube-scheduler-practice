package client

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type K8sClient struct {
	Clientset kubernetes.Interface
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
	return K8sClient{Clientset: clientset}, nil
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
	return K8sClient{Clientset: clientset}, nil
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
