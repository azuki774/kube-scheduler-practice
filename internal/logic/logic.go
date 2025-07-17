package logic

import v1 "k8s.io/api/core/v1"

// unscheduled pod が、配置して良いnodesを返す
func ChooseAvailableNodes(unschedulePod *v1.Pod, vs *v1.NodeList) (*v1.NodeList, error) {

	return nil, nil
}

// unscheduled podと配置していいnodesを与えると、配置するのに最適なnodeを返す
func ChooseSuitableNode(unschedulePod *v1.Pod, vs *v1.NodeList) (v1.Node, error) {

	return v1.Node{}, nil
}
