package scheduler

import (
	"errors"
	"strings"
	"sync"

	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/filter"
	"github.com/docker/swarm/scheduler/node"
	"github.com/docker/swarm/scheduler/strategy"
)

var (
	errNoNodeAvailable = errors.New("No nodes available in the cluster")
)

// Scheduler is exported
type Scheduler struct {
	sync.Mutex

	strategy strategy.PlacementStrategy
	filters  []filter.Filter
}

// New is exported
func New(strategy strategy.PlacementStrategy, filters []filter.Filter) *Scheduler {
	return &Scheduler{
		strategy: strategy,
		filters:  filters,
	}
}

// SelectNodesForContainer will return a list of nodes where the container can
// be scheduled, sorted by order or preference.
//The string refers to if request should receive cut or not, 0 == no cut, 2 == class 2 cut, etc
func (s *Scheduler) SelectNodesForContainer(nodes []*node.Node, config *cluster.ContainerConfig) ([]*node.Node, error, string) {
	candidates, err, cut := s.selectNodesForContainer(nodes, config, true)

	if err != nil {
		candidates, err, cut = s.selectNodesForContainer(nodes, config, false)
	}
	return candidates, err, cut
}

func (s *Scheduler) selectNodesForContainer(nodes []*node.Node, config *cluster.ContainerConfig, soft bool) ([]*node.Node, error, string) {

	if s.Strategy() == "energy" {
		return s.strategy.RankAndSort(config, nodes)
	}  

	accepted, err := filter.ApplyFilters(s.filters, config, nodes, soft)

	if err != nil {
		return nil, err,"0"
	}

	if len(accepted) == 0 {
		return nil, errNoNodeAvailable, "0"
	}

	return s.strategy.RankAndSort(config, accepted)
}

// Strategy returns the strategy name
func (s *Scheduler) Strategy() string {
	return s.strategy.Name()
}

// Filters returns the list of filter's name
func (s *Scheduler) Filters() string {
	filters := []string{}
	for _, f := range s.filters {
		filters = append(filters, f.Name())
	}

	return strings.Join(filters, ", ")
}
