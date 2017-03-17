package strategy

import (
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
	"github.com/docker/swarm/scheduler/filter"

	"net/http"	
	"encoding/json"
	"fmt"
)

type Host struct {
        HostID      string              `json:"hostid,omitempty"`
		HostIP		string				`json:"hostip,omitempty"`
        WorkerNodesID []string          `json:"workernodesid,omitempty"`
		WorkesNodes []*node.Node		`json:"workernodes,omitempty"`
        HostClass   string              `json:"hostclass,omitempty"`
        Region      string              `json:"region,omitempty"`
        TotalResourcesUtilization int   `json:"totalresouces,omitempty"`
        CPU_Utilization int             `json:"cpu,omitempty"`
        MemoryUtilization int           `json:"memory,omitempty"`
		AvailableCPUs int64				`json:"availablecpus,omitempty"`
        AvailableMemory int64			`json:"availablememory,omitempty"`
		AllocatedResources int          `json:"resoucesallocated,omitempty"`
        TotalHostResources int          `json:"totalresources,omitempty"`
        OverbookingFactor int           `json:"overbookingfactor,omitempty"`
}

type Task struct {
    TaskID string               `json:"taskid,omitempty"`
	TaskClass string			`json:"taskclass,omitempty"`
	CPU float64					`json:"cpu,omitempty"`
	Memory float64				`json:"memory,omitempty"`
    TaskType string             `json:"tasktype,omitempty"`
    CutReceived string          `json:"cutreceived,omitempty"`
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
func (p *EnergyPlacementStrategy) RankAndSort(config *cluster.ContainerConfig, nodes []*node.Node) ([]*node.Node, error, string, bool) {

	affinities, err := filter.ParseExprs(config.Affinities())
	fmt.Println(affinities)	
	requestClass := ""

	for _, affinity := range affinities {
		if affinity.Key == "requestclass" {
			requestClass = affinity.Value
		}
	}	

	if err != nil {
		return nil, err, "0", false
	}

	//+1 is the list type go get +2 is the other list type //see hostregistry.go for +info.. //endereço do manager e port do host registry
	listHostsLEE_DEE := GetHosts("http://192.168.1.154:12345/host/list/"+requestClass+"&1")
	
	/*
	output := make([]*node.Node,0)
	
	for i := 0; i < len(listHostsLEE_DEE); i++ {
		//check if host has enough resources to accomodate the request, if it does, return it
		host := listHostsLEE_DEE[i]
		if host.AvailableMemory > config.HostConfig.Memory && host.AvailableCPUs > config.HostConfig.CPUShares {
			//obter lista de workers do host e escolher um, randomly?
			for j := 0; j < len(nodes); j++ {			
				if nodes[j].ID == host.WorkerNodesID[0] && nodes[j].Name != "manager1" {
					output = append(output, nodes[j])
					return output, nil, requestClass, false
				}
			}
		}
			
	}*/
	//obtemos a nova listHostsLEE_DEE, desta vez ordenada de forma diferente	
	listHostsLEE_DEE = GetHosts("http://192.168.1.154:12345/host/list/"+requestClass+"&2")
	
	host, allocable, cut := cut(listHostsLEE_DEE, requestClass, config)
	if allocable == true { //if true then it means that we have a host that it can be scheduled	
		return findNode(host, nodes), nil, cut, true
	}
/*
	output[0] = kill()
	if len(output) > 0 
		return output, nil 
*/
	return nodes, nil, "0", false //can't be scheduled
}

func findNode(host *Host, nodes []*node.Node) ([]*node.Node) {
	output := make([]*node.Node,0)
	
	for j := 0; j < len(nodes); j++ {
		if nodes[j].ID == host.WorkerNodesID[0] && nodes[j].Name != "manager1" {
			output = append(output, nodes[j])
			return output
		}
	}
	return output
}

func cut(listHostsLEE_DEE []*Host, requestClass string, config *cluster.ContainerConfig) (*Host, bool, string) {
		
	for i := 0; i < len(listHostsLEE_DEE); i++ {
		cutList := make([]Task,0)
		listTasks := make([]Task,0)

		host := listHostsLEE_DEE[i]		

		if host.HostClass >= requestClass && requestClass != "4" {
			//listTasks = append(listTasks, GetTasks("http://" + host.HostIP + ":1234/task/higher/" + requestClass)
			listTasks = append(listTasks, GetTasks("http://192.168.1.154:1234/task/higher/" + requestClass)...)
		} else if requestClass != "1" && afterCutRequestFits(requestClass, host, config){
			return host, true, requestClass	//requestClass indicates the cut received, performed at /cluster/swarm/cluster.go
		} else if requestClass != "1" {
			listTasks = append(listTasks, GetTasks("http://192.168.1.154:1234/task/equalhigher/" + requestClass)...)
		}
		
		newCPU := 0.0
		newMemory := 0.0
		if requestClass != "1" {
			newMemory,newCPU = applyCut(requestClass, config)
		} 
	
		for _, task := range listTasks {
			if task.TaskClass == "1" {
				break
			}
			cutList = append(cutList, task)
			
			if fitAfterCuts(requestClass, host, newMemory, newCPU, cutList) {
			//	cutRequests() //inclui o incoming request
				fmt.Println("TRUE TRUE")
				return host, true, requestClass
			}
		}
		return host, true, requestClass 
	}
	return listHostsLEE_DEE[0], false, requestClass
}

//this function checks if after cutting the tasks and incoming request the request fits on this host
func fitAfterCuts(requestClass string, host *Host, memory float64, cpu float64, cutList []Task)(bool) {
	return true	
}

//por enquanto esta float mas depois mudar para int
func applyCut(requestClass string, config *cluster.ContainerConfig) (float64, float64){
	
	switch requestClass {
		case "2":
			newMemory := float64(config.HostConfig.Memory) * 0.2
			newCPU := float64(config.HostConfig.CPUShares) * 0.2
			return newMemory, newCPU
		case "3":
			newMemory := float64(config.HostConfig.Memory) * 0.3
			newCPU := float64(config.HostConfig.CPUShares) * 0.3
			return newMemory, newCPU
		case "4":
			newMemory := float64(config.HostConfig.Memory) * 0.4
			newCPU := float64(config.HostConfig.CPUShares) * 0.4
			return newMemory, newCPU
	}
	return 0.0,0.0
}

func afterCutRequestFits(requestClass string, host *Host, config *cluster.ContainerConfig) (bool) {

	newMemory, newCPU := applyCut(requestClass, config)
	
	if float64(host.AvailableMemory) > newMemory && float64(host.AvailableCPUs) > newCPU {
		return true
	} else {
		return false
	}
	return false
}

func GetTasks(url string) ([]Task) {
	req, err := http.NewRequest("GET", url, nil)
  
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

	var listTasks []Task 
	_ = json.NewDecoder(resp.Body).Decode(&listTasks)	
	
	return listTasks
	
}

func GetHosts(url string) ([]*Host) {
	req, err := http.NewRequest("GET", url, nil)
  
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

	var listHosts []*Host 
	_ = json.NewDecoder(resp.Body).Decode(&listHosts)	
	
	return listHosts
}

/*
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



