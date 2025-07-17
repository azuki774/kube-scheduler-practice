package logic

import (
	v1 "k8s.io/api/core/v1"
)

type ScheduleLogic struct{}

// unscheduled pod が、配置して良いnodesを返す
func (s *ScheduleLogic) ChooseAvailableNodes(unschedulePod *v1.Pod, vs *v1.NodeList) (*v1.NodeList, error) {

	return vs, nil
}

// unscheduled podと配置していいnodesを与えると、配置するのに最適なnodeを返す
func (s *ScheduleLogic) ChooseSuitableNode(unschedulePod *v1.Pod, vs *v1.NodeList) (v1.Node, error) {

	return vs.Items[0], nil
}
