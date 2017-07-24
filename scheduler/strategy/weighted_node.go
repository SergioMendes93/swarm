package strategy

import (
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
	"fmt"
)


// WeightedNode represents a node in the cluster with a given weight, typically used for sorting
// purposes.
type weightedNode struct {
	Node *node.Node
	// Weight is the inherent value of this node.
	Weight int64
}

type weightedNodeList []*weightedNode

func (n weightedNodeList) Len() int {
	return len(n)
}

func (n weightedNodeList) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n weightedNodeList) Less(i, j int) bool {
	var (
		ip = n[i]
		jp = n[j]
	)

	// If the nodes have the same weight sort them out by number of containers.
	if ip.Weight == jp.Weight {
		return len(ip.Node.Containers) < len(jp.Node.Containers)
	}
	return ip.Weight < jp.Weight
}

func weighNodes(config *cluster.ContainerConfig, nodes []*node.Node, healthinessFactor int64, nodesMap map[string]*node.Node) (weightedNodeList, error) {
	weightedNodes := weightedNodeList{}
	
	hosts := GetHosts("http://"+ipHostRegistry+":12345/host/list")		
/*
	for _, node := range nodes {
		nodeMemory := node.TotalMemory
		nodeCpus := node.TotalCpus * 1024 

		// Skip nodes that are smaller than the requested resources.
		if nodeMemory < int64(config.HostConfig.Memory) || nodeCpus < config.HostConfig.CPUShares {
			continue
		}

		var (
			cpuScore    int64 = 100
			memoryScore int64 = 100
		)

		if config.HostConfig.CPUShares > 0 {
			fmt.Println("Allocated CPUs " + hosts[node.IP].AllocatedCPUs)
			cpuScore = (hosts[node.IP].AllocatedCPUs + config.HostConfig.CPUShares) * 100 / nodeCpus
		}
		if config.HostConfig.Memory > 0 {
			memoryScore = (node.UsedMemory + config.HostConfig.Memory) * 100 / nodeMemory
		}

		if cpuScore <= 100 && memoryScore <= 100 {
			weightedNodes = append(weightedNodes, &weightedNode{Node: node, Weight: cpuScore + memoryScore + healthinessFactor*node.HealthIndicator})
		}
	}*/
		for _, host := range hosts {
			nodeMemory := host.TotalMemory
			nodeCpus := host.TotalCPUs 

			// Skip nodes that are smaller than the requested resources.
			if nodeMemory < int64(config.HostConfig.Memory) || nodeCpus < config.HostConfig.CPUShares {
				continue
			}

			var (
				cpuScore    int64 = 100
				memoryScore int64 = 100
			)

			if config.HostConfig.CPUShares > 0 {
				fmt.Print("Allocated CPUs ") 
				fmt.Println(host.AllocatedCPUs)

				cpuScore = (host.AllocatedCPUs + config.HostConfig.CPUShares) * 100 / nodeCpus
			}
			if config.HostConfig.Memory > 0 {
				fmt.Print("Allocated Memorys ") 
				fmt.Println(host.AllocatedMemory)

				memoryScore = (host.AllocatedMemory + config.HostConfig.Memory) * 100 / nodeMemory
			}

			if cpuScore <= 100 && memoryScore <= 100 {
				output := make([]*node.Node,0)
				output = append(output, nodesMap[host.HostIP])
				weightedNodes = append(weightedNodes, &weightedNode{Node: output[0], Weight: cpuScore + memoryScore + healthinessFactor*output[0].HealthIndicator})
			}
		}

	if len(weightedNodes) == 0 {
		return nil, ErrNoResourcesAvailable
	}

	return weightedNodes, nil
}
