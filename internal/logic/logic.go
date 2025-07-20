package logic

import (
	"math/rand"

	v1 "k8s.io/api/core/v1"
)

type ScheduleLogic struct{}

// unscheduled pod が、配置して良いnodesを返す
func (s *ScheduleLogic) ChooseAvailableNodes(unschedulePod *v1.Pod, vs *v1.NodeList) (*v1.NodeList, error) {
	retv := v1.NodeList{
		TypeMeta: vs.TypeMeta,
		ListMeta: vs.ListMeta,
		Items:    []v1.Node{},
	}
	for _, vi := range vs.Items {
		if vi.Labels["tier"] == "control" {
			continue
		}

		if unschedulePod.Spec.NodeSelector["tier"] != "cronjob" && vi.Labels["tier"] == "cronjob" {
			continue
		} else if unschedulePod.Spec.NodeSelector["tier"] == "cronjob" && vi.Labels["tier"] != "cronjob" {
			continue
		}

		// ここまで問題なければ配置してOK
		retv.Items = append(retv.Items, vi)
	}
	return &retv, nil
}

// unscheduled podと配置していいnodesを与えると、配置するのに最適なnodeを返す
func (s *ScheduleLogic) ChooseSuitableNode(unschedulePod *v1.Pod, vs *v1.NodeList) (v1.Node, error) {
	if len(vs.Items) == 0 {
		return v1.Node{}, nil
	}

	// vs.Items の要素からランダムで選択する
	idx := rand.Intn(len(vs.Items))
	return vs.Items[idx], nil
}
