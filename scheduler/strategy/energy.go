package strategy

import (
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
)


type hostInfo struct {

	HostID string
	TotalResourcesUtilization int
	CPU_Utilization int
	MemoryUtilization int
	HostClass int
	AllocatedResources int
	TotalHostResources int
	OverbookingFactor int
}


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
	
	output := make([]*node.Node,1)
	//listHostsLEE_DEE := getHostsListsLEE_DEE()
	
	//listHostsLEE_DEE é suposto ser um slice para o for abaixo funcionar
	for _, host := range listHostsLEE_DEE {
		//ver se o host tem recursos que chegue para acomodar o request, se fit entao faço return
		if fits {
			output[0] = host
			return output, nil
		}
			
	}

	//obtemos a nova listHostsLEE_DEE, desta vez ordenada de forma diferente
	output[0], request  =  cut()
	if len(output) > 0 { //if > 0 then it means that we have a host that it can be scheduled	
		return output, nil
	}

	output[0] = kill()
	if len(output) > 0 
		return output, nil

	return nil, nil //can't be scheduled
}

func cut() {

	for _, host := range listHostsLEE_DEE {
		cutList = 
		listTasks =
		
		if host.HostClass >= request.Class {
			listTasks = host.GetListsTasksHigherThanRequestClass()
		} else if request.Class != 1 && afterCutRequestFits() {
			newRequest = cutRequest(request)
			return host, request
		} else if request.Class != 1 {
			listTasks = host.getListTasksEqualHigherThanRequestClass()
		}
		
		for _, task := range listTasks {
			if task.Class == 1 {
				break
			}
			cutList += task
			
			if fitsAfterCut() {
				cutRequests()
				return host, nil
			}
		}

	}
}

func kill() {
	for _, host := range listHostsEED_DEE {
		possibleKillList =
		if request.Class == 4 && request.Type == Job {
			possibleKillList = host.GetListTasksClass4()
		} else if request.Class == 4 {
			return nil
		} else {
			possibleKillList = host.GetListTasksHigherThanRequestClass()
		}

		killList =
		for _, task := range possibleKillList {
			if request.Class == 4 && request.Type == Service {
				killList += task
			} else if request.Class != 4 {
				killList += task
			}

			if requestFits(request,killList) {
				kill(killList)
				reschedule(killList)
				return host
			}
		}

	}
}


/*
func (p *EnergyPlacementStrategy) getHostsListsLEE_DEE (argumentos, 1 ou 2 para identificar o tipo de sort a ser feito) {
		
}

func  (p *EnergyPlacementStrategy) getHostsListsEED_DEE (argumentos) {

}
*/



