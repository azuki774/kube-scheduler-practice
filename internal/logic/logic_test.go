package logic

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestScheduleLogic_ChooseAvailableNodes(t *testing.T) {
	type args struct {
		unschedulePod *v1.Pod
		nodes         *v1.NodeList
	}
	tests := []struct {
		name    string
		args    args
		want    *v1.NodeList
		wantErr bool
	}{
		{
			name: "success: filter control plane node",
			args: args{
				unschedulePod: &v1.Pod{},
				nodes: &v1.NodeList{
					Items: []v1.Node{
						{ObjectMeta: metav1.ObjectMeta{Name: "node1", Labels: map[string]string{"tier": "control"}}},
						{ObjectMeta: metav1.ObjectMeta{Name: "node2", Labels: map[string]string{"tier": "frontend"}}},
					},
				},
			},
			want: &v1.NodeList{
				Items: []v1.Node{
					{ObjectMeta: metav1.ObjectMeta{Name: "node2", Labels: map[string]string{"tier": "frontend"}}},
				},
			},
			wantErr: false,
		},
		{
			name: "success: filter cronjob node for normal pod",
			args: args{
				unschedulePod: &v1.Pod{},
				nodes: &v1.NodeList{
					Items: []v1.Node{
						{ObjectMeta: metav1.ObjectMeta{Name: "node1", Labels: map[string]string{"tier": "cronjob"}}},
						{ObjectMeta: metav1.ObjectMeta{Name: "node2", Labels: map[string]string{"tier": "frontend"}}},
					},
				},
			},
			want: &v1.NodeList{
				Items: []v1.Node{
					{ObjectMeta: metav1.ObjectMeta{Name: "node2", Labels: map[string]string{"tier": "frontend"}}},
				},
			},
			wantErr: false,
		},
		{
			name: "success: filter normal node for cronjob pod",
			args: args{
				unschedulePod: &v1.Pod{
					TypeMeta: metav1.TypeMeta{Kind: "CronJob"},
				},
				nodes: &v1.NodeList{
					Items: []v1.Node{
						{ObjectMeta: metav1.ObjectMeta{Name: "node1", Labels: map[string]string{"tier": "cronjob"}}},
						{ObjectMeta: metav1.ObjectMeta{Name: "node2", Labels: map[string]string{"tier": "frontend"}}},
					},
				},
			},
			want: &v1.NodeList{
				Items: []v1.Node{
					{ObjectMeta: metav1.ObjectMeta{Name: "node1", Labels: map[string]string{"tier": "cronjob"}}},
				},
			},
			wantErr: false,
		},
		{
			name: "success: no available nodes",
			args: args{
				unschedulePod: &v1.Pod{},
				nodes: &v1.NodeList{
					Items: []v1.Node{
						{ObjectMeta: metav1.ObjectMeta{Name: "node1", Labels: map[string]string{"tier": "control"}}},
						{ObjectMeta: metav1.ObjectMeta{Name: "node2", Labels: map[string]string{"tier": "cronjob"}}},
					},
				},
			},
			want: &v1.NodeList{
				Items: []v1.Node{},
			},
			wantErr: false,
		},
		{
			name: "success: multiple available nodes",
			args: args{
				unschedulePod: &v1.Pod{},
				nodes: &v1.NodeList{
					Items: []v1.Node{
						{ObjectMeta: metav1.ObjectMeta{Name: "node1", Labels: map[string]string{"tier": "frontend"}}},
						{ObjectMeta: metav1.ObjectMeta{Name: "node2", Labels: map[string]string{"tier": "backend"}}},
					},
				},
			},
			want: &v1.NodeList{
				Items: []v1.Node{
					{ObjectMeta: metav1.ObjectMeta{Name: "node1", Labels: map[string]string{"tier": "frontend"}}},
					{ObjectMeta: metav1.ObjectMeta{Name: "node2", Labels: map[string]string{"tier": "backend"}}},
				},
			},
			wantErr: false,
		},
		{
			name: "success: only control node",
			args: args{
				unschedulePod: &v1.Pod{},
				nodes: &v1.NodeList{
					Items: []v1.Node{
						{ObjectMeta: metav1.ObjectMeta{Name: "node1", Labels: map[string]string{"tier": "control"}}},
					},
				},
			},
			want: &v1.NodeList{
				Items: []v1.Node{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScheduleLogic{}
			got, err := s.ChooseAvailableNodes(tt.args.unschedulePod, tt.args.nodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScheduleLogic.ChooseAvailableNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScheduleLogic.ChooseAvailableNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}
