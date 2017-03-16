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
    AllocatedResources string   `json:"allocatedresources,omitempty"`
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
func (p *EnergyPlacementStrategy) RankAndSort(config *cluster.ContainerConfig, nodes []*node.Node) ([]*node.Node, error) {

	affinities, err := filter.ParseExprs(config.Affinities())
	fmt.Println(affinities)	
	requestClass := ""

	for _, affinity := range affinities {
		if affinity.Key == "requestclass" {
			requestClass = affinity.Value
		}
	}	

	if err != nil {
		return nil, err
	}

	//+1 is the list type go get +2 is the other list type //see hostregistry.go for +info.. //endereço do manager e port do host registry
	listHostsLEE_DEE := GetHosts("http://192.168.1.154:12345/host/list/"+requestClass+"&1")
	
	fmt.Println(listHostsLEE_DEE)
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
					return output, nil
				}
			}
		}
			
	}*/
	//obtemos a nova listHostsLEE_DEE, desta vez ordenada de forma diferente	
	listHostsLEE_DEE = GetHosts("http://192.168.1.154:12345/host/list/"+requestClass+"&2")
	
	fmt.Println("Nova Lista")
	fmt.Println(listHostsLEE_DEE)

	host, allocable :=  cut(listHostsLEE_DEE, requestClass, config)
	if allocable == true { //if true then it means that we have a host that it can be scheduled	
		
		return findNode(host, nodes), nil
	}
/*
	output[0] = kill()
	if len(output) > 0 
		return output, nil 
*/
	return nodes, nil //can't be scheduled
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

func cut(listHostsLEE_DEE []*Host, requestClass string, config *cluster.ContainerConfig) (*Host, bool) {
		
	for i := 0; i < len(listHostsLEE_DEE); i++ {
		cutList := make([]Task,0)
		listTasks := make([]Task,0)

		host := listHostsLEE_DEE[i]		

		if host.HostClass >= requestClass && requestClass != "4" {
			//listTasks = append(listTasks, GetTasks("http://" + host.HostIP + ":1234/task/higher/" + requestClass)
			listTasks = append(listTasks, GetTasks("http://192.168.1.154:1234/task/higher/" + requestClass)...)
		} else if requestClass != "1" {
			listTasks = append(listTasks, GetTasks("http://192.168.1.154:1234/task/equalhigher/" + requestClass)...)
		}
		/* else if requestClass != "1" && afterCutRequestFits() {
			newRequest = cutRequest(request)
			return host, request*
		}*/

		fmt.Println("list tasks")
		fmt.Println(listTasks)
		
		for _, task := range listTasks {
			if task.TaskClass == "1" {
				break
			}
			cutList = append(cutList, task)
			/*
			if fitsAfterCut() {
				cutRequests()
				return host, nil
			}*/
		}
		fmt.Println("cut list")
		fmt.Println(cutList)
		return host, true
	}
	return listHostsLEE_DEE[0], false
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



