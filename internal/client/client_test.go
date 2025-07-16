package client

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestK8sClient_GetPodsNotScheduled(t *testing.T) {
	type fields struct {
		K8sClient K8sClient
	}
	unscheduledPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "unscheduled-pod",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			NodeName: "", // NodeNameが空
		},
	}

	scheduledPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "scheduled-pod",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			NodeName: "kind-worker", // NodeNameが設定済み
		},
	}

	fakeClientset := fake.NewSimpleClientset(unscheduledPod, scheduledPod)

	tests := []struct {
		name    string
		fields  fields
		want    *v1.PodList
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				K8sClient: K8sClient{Clientset: fakeClientset},
			},
			want: &v1.PodList{
				Items: []v1.Pod{*unscheduledPod},
			},
			wantErr: false,
		},
		{
			name: "none",
			fields: fields{
				K8sClient: K8sClient{Clientset: fake.NewSimpleClientset()},
			},
			want: &v1.PodList{
				Items: []v1.Pod{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &K8sClient{
				Clientset: tt.fields.K8sClient.Clientset,
			}
			got, err := k.GetUnscheduledPods()
			if (err != nil) != tt.wantErr {
				t.Errorf("K8sClient.GetUnscheduledPods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got.Items) != len(tt.want.Items) {
				t.Fatalf("Expected to get %d pod(s), but got %d", len(tt.want.Items), len(got.Items))
			}

			expectedPodNames := make(map[string]struct{})
			for _, pod := range tt.want.Items {
				expectedPodNames[pod.Name] = struct{}{}
			}

			// 実際に返ってきたPodが、すべて期待したものであることを確認する
			for _, gotPod := range got.Items {
				if _, ok := expectedPodNames[gotPod.Name]; !ok {
					t.Errorf("Unexpected pod found in results: %s", gotPod.Name)
				}
			}
		})
	}
}
