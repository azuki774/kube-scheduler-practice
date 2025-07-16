package client

import (
	"context"
	"log"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type K8sClient struct {
	Clientset *kubernetes.Clientset
}

func (k *K8sClient) GetNodes() {
	nodes, err := k.Clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error getting nodes: %s", err.Error())
	}

	for _, node := range nodes.Items {
		slog.Info("node", "name", node.Name, "label.tier", node.Labels["tier"])
	}
}
