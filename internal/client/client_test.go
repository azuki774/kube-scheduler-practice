package client

import (
	"errors"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	coretesting "k8s.io/client-go/testing"
)

func TestK8sClient_GetNodes(t *testing.T) {
	type fields struct {
		K8sClient K8sClient
	}

	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
	}
	node2 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
	}

	tests := []struct {
		name    string
		fields  fields
		want    *v1.NodeList
		wantErr bool
	}{
		{
			name: "success: get multiple nodes",
			fields: fields{
				// node1とnode2が存在するFakeClientSetを準備
				K8sClient: K8sClient{Clientset: fake.NewSimpleClientset(node1, node2)},
			},
			want: &v1.NodeList{
				Items: []v1.Node{*node1, *node2},
			},
			wantErr: false,
		},
		{
			name: "success: get one node",
			fields: fields{
				// node1が存在するFakeClientSetを準備
				K8sClient: K8sClient{Clientset: fake.NewSimpleClientset(node1)},
			},
			want: &v1.NodeList{
				Items: []v1.Node{*node1},
			},
			wantErr: false,
		},
		{
			name: "success: get no nodes",
			fields: fields{
				// Nodeが一つも存在しない空のFakeClientSetを準備
				K8sClient: K8sClient{Clientset: fake.NewSimpleClientset()},
			},
			want: &v1.NodeList{
				Items: []v1.Node{},
			},
			wantErr: false,
		},
	}

	// --- テストの実行ループ ---
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &K8sClient{
				Clientset: tt.fields.K8sClient.Clientset,
			}
			got, err := k.GetNodes()
			if (err != nil) != tt.wantErr {
				t.Errorf("K8sClient.GetNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// --- 結果の検証 ---
			// 1. 取得したNodeの数が期待通りか
			if len(got.Items) != len(tt.want.Items) {
				t.Fatalf("Expected to get %d node(s), but got %d", len(tt.want.Items), len(got.Items))
			}

			// 2. 取得したNodeの名前が期待通りか (順不同でチェック)
			expectedNodeNames := make(map[string]struct{})
			for _, node := range tt.want.Items {
				expectedNodeNames[node.Name] = struct{}{}
			}

			for _, gotNode := range got.Items {
				if _, ok := expectedNodeNames[gotNode.Name]; !ok {
					t.Errorf("Unexpected node found in results: %s", gotNode.Name)
				}
			}
		})
	}
}

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

func TestK8sClient_AssignPodToNode(t *testing.T) {
	type fields struct {
		Clientset kubernetes.Interface
	}
	type args struct {
		pod  *v1.Pod
		node *v1.Node
	}

	// --- テストデータとシナリオの準備 ---
	podToBind := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"}}
	nodeToBindTo := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}}

	// テストケースを定義
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success: correctly bind pod to node",
			fields: fields{
				Clientset: func() *fake.Clientset {
					clientset := fake.NewSimpleClientset()
					clientset.PrependReactor("create", "pods", func(action coretesting.Action) (handled bool, ret runtime.Object, err error) {
						createAction := action.(coretesting.CreateAction)
						if createAction.GetSubresource() != "binding" {
							return false, nil, nil // bindingでなければこのリアクターは処理しない
						}

						binding := createAction.GetObject().(*v1.Binding)
						if binding.Name != podToBind.Name || binding.Target.Name != nodeToBindTo.Name {
							t.Errorf("mismatched binding: got pod %s on node %s", binding.Name, binding.Target.Name)
						}

						// 成功したことを示す
						return true, binding, nil
					})
					return clientset
				}(),
			},
			args: args{
				pod:  podToBind,
				node: nodeToBindTo,
			},
			wantErr: false,
		},
		{
			name: "failure: API server returns error",
			fields: fields{
				Clientset: func() *fake.Clientset {
					clientset := fake.NewSimpleClientset()
					// 常にエラーを返すリアクターを設定
					clientset.PrependReactor("create", "pods", func(action coretesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New("simulated API error")
					})
					return clientset
				}(),
			},
			args: args{
				pod:  podToBind,
				node: nodeToBindTo,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &K8sClient{
				Clientset: tt.fields.Clientset,
			}
			if err := k.AssignPodToNode(tt.args.pod, tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("K8sClient.AssignPodToNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
