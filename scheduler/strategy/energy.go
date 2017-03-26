package strategy

import (
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
	"github.com/docker/swarm/scheduler/filter"

	"time"
	"math/rand"	
	"net/http"	
	"encoding/json"
	"strconv"
	"fmt"
)

type Host struct {
        HostID      string              	`json:"hostid,omitempty"`
		HostIP		string					`json:"hostip,omitempty"`
        WorkerNodesID []string          	`json:"workernodesid,omitempty"`
		WorkesNodes []*node.Node			`json:"workernodes,omitempty"`
        HostClass   string              	`json:"hostclass,omitempty"`
        Region      string              	`json:"region,omitempty"`
        TotalResourcesUtilization string   	`json:"totalresouces,omitempty"`
        CPU_Utilization int             	`json:"cpu,omitempty"`
        MemoryUtilization int           	`json:"memory,omitempty"`
		AvailableCPUs int64					`json:"availablecpus,omitempty"`
        AvailableMemory int64				`json:"availablememory,omitempty"`
		AllocatedResources int          	`json:"resoucesallocated,omitempty"`
        TotalHostResources int          	`json:"totalresources,omitempty"`
        OverbookingFactor int           	`json:"overbookingfactor,omitempty"`
}

type Task struct {
    TaskID string               `json:"taskid,omitempty"`
	TaskClass string			`json:"taskclass,omitempty"`
	Image string				`json:"image,omitempty"`
	CPU string					`json:"cpu,omitempty"`
	Memory string				`json:"memory,omitempty"`
    TaskType string             `json:"tasktype,omitempty"`
    CutReceived string          `json:"cutreceived,omitempty"`
}

var ipAddress = "192.168.1.154"

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

func findNode(host *Host, nodes []*node.Node) ([]*node.Node) {
	output := make([]*node.Node,0)

	//going to choose a worker randomly from the host
	numWorkers := len(host.WorkerNodesID)
	
	seed := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(seed)
	randomNumber := r1.Intn(numWorkers)	

	fmt.Println("random number")
	fmt.Println(randomNumber)

	for j := 0; j < len(nodes); j++ {
		if nodes[j].ID == host.WorkerNodesID[randomNumber] && nodes[j].Name != "manager1" {
			output = append(output, nodes[j])
			return output
		}
	}
	return output
}

// RankAndSort randomly sorts the list of nodes.
func (p *EnergyPlacementStrategy) RankAndSort(config *cluster.ContainerConfig, nodes []*node.Node) ([]*node.Node, error, string, string, bool) {

	affinities, err := filter.ParseExprs(config.Affinities())
	requestClass := ""
	requestType := ""

	for _, affinity := range affinities {
		if affinity.Key == "requestclass" {
			requestClass = affinity.Value
			continue
		}
		if affinity.Key == "requesttype" {
			requestType = affinity.Value
		}
	}	

	if err != nil {
		return nil, err, "0", requestType, false
	}

	//+1 is the list type go get +2 is the other list type //see hostregistry.go for +info.. //endereço do manager e port do host registry
	listHostsLEE_DEE := GetHosts("http://"+ipAddress+":12345/host/list/"+requestClass+"&1")
		
	
//------------ Scheduling without cuts and kills----------------------------------------------------------

	for i := 0; i < len(listHostsLEE_DEE); i++ {
		//check if host has enough resources to accomodate the request, if it does, return it
		host := listHostsLEE_DEE[i]
		if host.AvailableMemory > config.HostConfig.Memory && host.AvailableCPUs > config.HostConfig.CPUShares {
			return findNode(host,nodes), nil, requestClass, requestType, false
		}
			
	}
//------------------------ Cut algorithm ------------------------

	//obtemos a nova listHostsLEE_DEE, desta vez ordenada de forma diferente	
	listHostsLEE_DEE = GetHosts("http://192.168.1.154:12345/host/list/"+requestClass+"&2")
	
	host, allocable, cut := cut(listHostsLEE_DEE, requestClass, config)
	if allocable { //if true then it means that we have a host that it can be scheduled	
		return findNode(host, nodes), nil, cut, requestType, true
	}

	return nodes, nil, requestClass, requestType, false

//-------------------------Kill algorithm ----------------------------

	listHostsEED_DEE := GetHosts("http://"+ipAddress+":12345/host/listkill/"+requestClass)

	host, allocable = kill(listHostsEED_DEE, requestClass, requestType, config)
	
	if allocable {
		return findNode(host,nodes), nil, requestClass, requestType,  false
	}

//	return nodes, nil, requestClass, requestType, false
	return nil, nil, requestClass, requestType, false //can't be scheduled É PARA FICAR ESTE QUANDO FOR DEFINITIVO
}


func kill(listHostsEED_DEE []*Host, requestClass string, requestType string, config *cluster.ContainerConfig) (*Host, bool) {
	for i := 0 ; i < len(listHostsEED_DEE); i++ {
		possibleKillList := make([]Task, 0)
		host := listHostsEED_DEE[i] 

		if requestClass == "4" && requestType == "job"{
			possibleKillList = append(possibleKillList, GetTasks("http://"+ipAddress+":1234/task/class4")...)

		} else if requestClass == "4" {
			return nil, false
		} else {
			possibleKillList = append(possibleKillList, GetTasks("http://"+ipAddress+":1234/task/higher/" + requestClass)...)
		}

		killList := make([]Task, 0)
		for _, task := range possibleKillList {
			
			if task.TaskClass == "4" && task.TaskType == "service" {
				killList = append(killList, task)
			} else if requestClass != "4" {
				killList = append(killList, task)
			}

			if requestFitsAfterKills(killList, host, config) {
				go killTasks(killList)
				go reschedule(killList)
				return host, true
			}
		}
	}
	return nil, false //em definitivo sera nil,false
}

func reschedule(killList []Task) {
	for _, task := range killList {
		go UpdateTask("http://"+ipAddress+":12345/host/reschedule/"+task.CPU+"&"+task.Memory+"&"+task.TaskClass+"&"+task.Image)
	}
}

func killTasks(killList []Task) {
	for _, task := range killList {
		go UpdateTask("http://"+ipAddress+":12345/host/killtask/"+task.TaskID)
		go UpdateTask("http://"+ipAddress+":1234/task/remove/"+task.TaskID)
	}
}

func requestFitsAfterKills(killList []Task, host *Host, config *cluster.ContainerConfig) (bool) {
	cpuReduction := 0.0
	memoryReduction := 0.0
	
	for _, task := range killList {
		switch task.TaskClass {
			case "2":
				taskCPU, _ := strconv.ParseFloat(task.CPU, 64)
				taskMemory, _ := strconv.ParseFloat(task.Memory, 64)
				cpuReduction += taskCPU 
				memoryReduction +=  taskMemory
				break
			case "3":
				taskCPU, _ := strconv.ParseFloat(task.CPU, 64)
				taskMemory, _ := strconv.ParseFloat(task.Memory, 64)
				cpuReduction += taskCPU 
				memoryReduction +=  taskMemory 
				break
			case "4":
				taskCPU, _ := strconv.ParseFloat(task.CPU, 64)
				taskMemory, _ := strconv.ParseFloat(task.Memory, 64)
				cpuReduction += taskCPU 
				memoryReduction +=  taskMemory
				break
		}
	}

	//after killing the tasks, the host will have the following memory and cpu
	hostMemory := float64(host.AvailableMemory) - memoryReduction
	hostCPU := float64(host.AvailableCPUs) - cpuReduction
	
	//lets see if after those kill the request will now fit
	if hostMemory > float64(config.HostConfig.Memory) && hostCPU > float64(config.HostConfig.CPUShares) {
		return true
	} else {
		return false
	}
	
	return true	

}

func cut(listHostsLEE_DEE []*Host, requestClass string, config *cluster.ContainerConfig) (*Host, bool, string) {
		
	for i := 0; i < len(listHostsLEE_DEE); i++ {
		cutList := make([]Task,0)
		listTasks := make([]Task,0)

		host := listHostsLEE_DEE[i]		

		
		if host.HostClass >= requestClass && requestClass != "4" {
			//listTasks = append(listTasks, GetTasks("http://" + host.HostIP + ":1234/task/higher/" + requestClass)
			listTasks = append(listTasks, GetTasks("http://"+ipAddress+":1234/task/higher/" + requestClass)...)
		//} //else if requestClass != "1" && afterCutRequestFits(requestClass, host, config){
		//	return host, true, requestClass	//requestClass indicates the cut received, performed at /cluster/swarm/cluster.go
		} else if requestClass != "1" {
			listTasks = append(listTasks, GetTasks("http://"+ipAddress+":1234/task/equalhigher/" + requestClass)...)
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
				cutRequests(cutList)
				fmt.Println("Cut done") 
				return host, true, requestClass
			}
		}
		//return host, false, requestClass 
	}
	return listHostsLEE_DEE[0], false, requestClass
}

func cutRequests(cutList []Task) {
	for _, task := range cutList {
		newCPU, err := strconv.ParseFloat(task.CPU, 64) 
		newMemory, err := strconv.ParseFloat(task.Memory, 64) 

		newCPU *= 0.2
		newMemory *= 0.2

		if err != nil {
			fmt.Println(err)
		}
		cpu := strconv.FormatFloat(newCPU, 'f', -1, 64)
		memory := strconv.FormatFloat(newMemory, 'f', -1, 64)
	
		cutReceived := "20"

		go UpdateTask("http://"+ipAddress+":1234/task/updatetask/"+task.TaskClass+"&"+cpu+"&"+memory+"&"+task.TaskID+"&"+cutReceived)
		go UpdateTask("http://"+ipAddress+":12345/host/updatetask/"+task.TaskID+"&"+cpu+"&"+memory)
	}
}

func UpdateTask (url string) {

	req, err := http.NewRequest("GET", url, nil)
    req.Header.Set("X-Custom-Header", "myvalue")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
}

//this function checks if after cutting the tasks and incoming request the request fits on this host
func fitAfterCuts(requestClass string, host *Host, memory float64, cpu float64, cutList []Task)(bool) {
	
	cpuReduction := 0.0
	memoryReduction := 0.0
	
	for _, task := range cutList {
		switch task.TaskClass {
			case "2":
				taskCPU, _ := strconv.ParseFloat(task.CPU, 64)
				taskMemory, _ := strconv.ParseFloat(task.Memory, 64)
				cpuReduction += taskCPU * 0.2
				memoryReduction +=  taskMemory * 0.2
				break
			case "3":
				taskCPU, _ := strconv.ParseFloat(task.CPU, 64)
				taskMemory, _ := strconv.ParseFloat(task.Memory, 64)
				cpuReduction += taskCPU * 0.2
				memoryReduction +=  taskMemory * 0.2
				break
			case "4":
				taskCPU, _ := strconv.ParseFloat(task.CPU, 64)
				taskMemory, _ := strconv.ParseFloat(task.Memory, 64)
				cpuReduction += taskCPU * 0.2
				memoryReduction +=  taskMemory * 0.2
				break
		}
	}

	//after cutting the tasks, the host will have the following memory and cpu
	hostMemory := float64(host.AvailableMemory) - memoryReduction
	hostCPU := float64(host.AvailableCPUs) - cpuReduction
	
	//lets see if after those cuts the request will now fit
	if hostMemory > memory && hostCPU > cpu {
		return true
	} else {
		return false
	}
	
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

