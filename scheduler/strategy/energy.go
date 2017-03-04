package strategy

import (
	"math/rand"
	"time"

	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
)

// EnergyPlacementStrategy randomly places the container into the cluster.
type EnergyPlacementStrategy struct {
}

// Initialize a RandomPlacementStrategy.
func (p *EnergyPlacementStrategy) Initialize() error {
	return nil
}

// Name returns the name of the strategy.
func (p *EnergyPlacementStrategy) Name() string {
	return "energy"
}

// RankAndSort randomly sorts the list of nodes.
func (p *EnergyPlacementStrategy) RankAndSort(config *cluster.ContainerConfig, nodes []*node.Node) ([]*node.Node, error) {
	for i := len(nodes) - 1; i > 0; i-- {
		j := p.r.Intn(i + 1)
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}
	return nodes, nil
}

func (argumentos, 1 ou 2 significando como devera estar ordenado) getHostsListsLEE_DEE (retorno) {
		
}

func  (argumentos) getHostsListsEED_DEE (retorno) {

}




