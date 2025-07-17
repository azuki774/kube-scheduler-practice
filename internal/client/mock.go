package client

import v1 "k8s.io/api/core/v1"

type mockScheduleLogic struct {
	funcChooseAvailableNodes func(unschedulePod *v1.Pod, vs *v1.NodeList) (*v1.NodeList, error)
	funcChooseSuitableNode   func(unschedulePod *v1.Pod, vs *v1.NodeList) (v1.Node, error)
}

func (m *mockScheduleLogic) ChooseAvailableNodes(unschedulePod *v1.Pod, vs *v1.NodeList) (*v1.NodeList, error) {
	return m.funcChooseAvailableNodes(unschedulePod, vs)
}

func (m *mockScheduleLogic) ChooseSuitableNode(unschedulePod *v1.Pod, vs *v1.NodeList) (v1.Node, error) {
	return m.funcChooseSuitableNode(unschedulePod, vs)
}
