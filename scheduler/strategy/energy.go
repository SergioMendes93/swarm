package strategy

import (
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
	"github.com/docker/swarm/scheduler/filter"

	"math"
	"errors"
	"net/http"	
	"encoding/json"
	"strconv"
	"fmt"
)

type Host struct {
    	HostIP                    string       `json:"hostip, omitempty"`
    	WorkerNodes               []*node.Node    `json:"workernode,omitempty"`
    	HostClass                 string       `json:"hostclass,omitempty"`
    	Region                    string       `json:"region,omitempty"`
    	TotalResourcesUtilization string       `json:"totalresouces,omitempty"`
    	CPU_Utilization           string       `json:"cpu,omitempty"`
    	MemoryUtilization         string       `json:"memory,omitempty"`
    	AllocatedMemory           float64      `json:"allocatedmemory,omitempty"`
    	AllocatedCPUs             float64      `json:"allocatedcpus,omitempty"`
    	OverbookingFactor         float64      `json:"overbookingfactor,omitempty"`
    	TotalMemory               float64      `json:"totalmemory,omitempty"`
    	TotalCPUs                 float64      `json:"totalcpus, omitempty"`

}

type Task struct {
    	TaskID string           `json:"taskid,omitempty"`
	TaskClass string	`json:"taskclass,omitempty"`
	Image string		`json:"image,omitempty"`
	CPU string		`json:"cpu,omitempty"`
	TotalResources string	`json:"totalresources,omitempty"`
	Memory string		`json:"memory,omitempty"`
    	TaskType string         `json:"tasktype,omitempty"`
    	CutReceived string      `json:"cutreceived,omitempty"`
	CutToReceive string	`json:"cuttoreceive,omitempty"`
}

var ipHostRegistry = ""
//var ipAddress = getIPAddress()

var MAX_OVERBOOKING_CLASS1 = 1.0
var MAX_OVERBOOKING_CLASS2 = 1.16
var MAX_OVERBOOKING_CLASS3 = 1.33
var MAX_OVERBOOKING_CLASS4 = 1.5

var MAX_CUT_CLASS2 = 0.16
var MAX_CUT_CLASS3 = 0.33
var MAX_CUT_CLASS4 = 0.5

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
func CheckOverbookingLimit(host *Host, hostAllocatedCPUs float64, hostAllocatedMemory float64, requestCPUs float64, requestMemory float64) bool {
	//Lets get the new overbooking factor after the request is allocated  (worst case estimation because we assume that they are going to use all the resources)
	overbookingMemory := (hostAllocatedMemory + requestMemory)/host.TotalMemory
	overbookingCPUs := 	(hostAllocatedCPUs + requestCPUs)/host.TotalCPUs
	
	newOverbooking := math.Max(overbookingMemory,overbookingCPUs)
	switch host.HostClass {
		case "1":
			if newOverbooking < MAX_OVERBOOKING_CLASS1 {
				return true
			}
			break
		case "2":
			if newOverbooking < MAX_OVERBOOKING_CLASS2 {
				return true
			}
			break
		case "3":
			if newOverbooking < MAX_OVERBOOKING_CLASS3 {
				return true
			}

			break
		case "4":
			if newOverbooking < MAX_OVERBOOKING_CLASS4 {
				return true
			}
			break
	}
	return false
}

// RankAndSort randomly sorts the list of nodes.
func (p *EnergyPlacementStrategy) RankAndSort(config *cluster.ContainerConfig, nodes []*node.Node) ([]*node.Node, error, string, string, float64) {

//	ipHostRegistry = getIPAddress()

	ipHostRegistry = "146.193.41.142"

	output := make([]*node.Node,0)

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
		return nil, err, "0", requestType, 0.0
	}

	//+1 is the list type go get +2 is the other list type //see hostregistry.go for +info.. //endereço do manager e port do host registry
	listHostsLEE_DEE := GetHosts("http://"+ipHostRegistry+":12345/host/list/"+requestClass+"&1")
		
	
//------------ Scheduling without cuts and kills----------------------------------------------------------

	for i := 0; i < len(listHostsLEE_DEE); i++ {
		//we must check if after allocating this request the overbooking factor of the host doesnt go over the host maximum overbooking allowed
		//if it does no go over it, then the request can be scheduled to this host
		host := listHostsLEE_DEE[i]
		if CheckOverbookingLimit(host, host.AllocatedCPUs, host.AllocatedMemory, float64(config.HostConfig.CPUShares), float64(config.HostConfig.Memory)) {
			//return findNode(host,nodes), nil, requestClass, requestType, false
			output = append(output, host.WorkerNodes[0])
			return output, nil, requestClass, requestType, 0.0
		}
			
	}
//------------------------ Cut algorithm ------------------------

	//obtemos a nova listHostsLEE_DEE, desta vez ordenada de forma diferente	
	listHostsLEE_DEE = GetHosts("http://"+ipHostRegistry+":12345/host/list/"+requestClass+"&2")

	host, allocable, cut := cut(listHostsLEE_DEE, requestClass, config)
	if allocable { //if true then it means that we have a host that it can be scheduled	
		//return findNode(host, nodes), nil, cut, requestType, true
		output = append(output, host.WorkerNodes[0])
		return output, nil, requestClass, requestType, cut
	}

//-------------------------Kill algorithm ----------------------------

	listHostsEED_DEE := GetHosts("http://"+ipHostRegistry+":12345/host/listkill/"+requestClass)

	host, allocable = kill(listHostsEED_DEE, requestClass, requestType, config)
	
	if allocable {
		//return findNode(host,nodes), nil, requestClass, requestType,  false
		output = append(output, host.WorkerNodes[0])
		return output, nil, requestClass, requestType,  0.0
	}
	fmt.Println("Not allocable")
	return nil, errors.New("Does not fit on current hosts"), requestClass, requestType, 0.0 //can't be scheduled É PARA FICAR ESTE QUANDO FOR DEFINITIVO
}


func kill(listHostsEED_DEE []*Host, requestClass string, requestType string, config *cluster.ContainerConfig) (*Host, bool) {
	for i := 0 ; i < len(listHostsEED_DEE); i++ {
		possibleKillList := make([]Task, 0)
		host := listHostsEED_DEE[i] 

		if requestClass == "4" && requestType == "job"{
			possibleKillList = append(possibleKillList, GetTasks("http://"+host.HostIP+":1234/task/class4")...)

		} else if requestClass == "4" {
			return nil, false
		} else {
			possibleKillList = append(possibleKillList, GetTasks("http://"+host.HostIP+":1234/task/higher/" + requestClass)...)
		}

		killList := make([]Task, 0)
		for _, task := range possibleKillList {
			
			if task.TaskClass == "4" && task.TaskType == "service" {
				killList = append(killList, task)
			} else if requestClass != "4" {
				killList = append(killList, task)
			}

			if requestFitsAfterKills(killList, host, config) {
				go killTasks(killList, host.HostIP)
				go reschedule(killList)
				return host, true
			}
		}
	}
	return nil, false //em definitivo sera nil,false
}

func reschedule(killList []Task) {
	for _, task := range killList {
		go UpdateTask("http://"+ipHostRegistry+":12345/host/reschedule/"+task.CPU+"&"+task.Memory+"&"+task.TaskClass+"&"+task.Image)
	}
}

func killTasks(killList []Task, hostIP string) {
	for _, task := range killList {
		//this besides killing the task it will also update the allocated cpu/memory  
		go UpdateTask("http://"+ipHostRegistry+":12345/host/killtask/"+task.TaskID+"&"+task.CPU+"&"+task.Memory+"&"+hostIP)
		go UpdateTask("http://"+hostIP+":1234/task/remove/"+task.TaskID)
	}
}

func requestFitsAfterKills(killList []Task, host *Host, config *cluster.ContainerConfig) (bool) {
	cpuReduction := 0.0
	memoryReduction := 0.0
	
	for _, task := range killList {
		taskCPU, _ := strconv.ParseFloat(task.CPU, 64)
		taskMemory, _ := strconv.ParseFloat(task.Memory, 64)
		cpuReduction += taskCPU 
		memoryReduction +=  taskMemory
	}

	//after killing the tasks, the host will have the following memory and cpu
	hostMemory := host.AllocatedMemory - memoryReduction
	hostCPU := host.AllocatedCPUs - cpuReduction

	//lets see if after those kill the request will now fit
	return CheckOverbookingLimit(host, hostCPU, hostMemory, float64(config.HostConfig.CPUShares), float64(config.HostConfig.Memory))	
}

func cut(listHostsLEE_DEE []*Host, requestClass string, config *cluster.ContainerConfig) (*Host, bool, float64) {
		
	for i := 0; i < len(listHostsLEE_DEE); i++ {
		cutList := make([]Task,0)
		listTasks := make([]Task,0)

		host := listHostsLEE_DEE[i]		

		if host.HostClass >= requestClass && requestClass != "4" {
			listTasks = append(listTasks, GetTasks("http://"+host.HostIP+":1234/task/highercut/" + requestClass)...)
		} else if requestClass != "1" && afterCutRequestFits(requestClass, host, config){
			cutToReceive := amountToCut(requestClass, host.HostClass)
			return host, true, cutToReceive	//cutToReceived indicates the cut to be received by this request, performed at /cluster/swarm/cluster.go
		} else if requestClass != "1" {
			listTasks = append(listTasks, GetTasks("http://"+host.HostIP+":1234/task/equalhigher/" + requestClass+"&"+host.HostClass)...)
		}
		
		newCPU := 0.0
		newMemory := 0.0
		canCut := false
		
		if requestClass != "1" {
			newMemory,newCPU,canCut = applyCut(requestClass, config, host.HostClass)
			if !canCut { //if we cannot cut the request we use the original memory and cpu values
				newCPU = float64(config.HostConfig.CPUShares)
				newMemory = float64(config.HostConfig.Memory)				
			}
		} else { //if its a request class 1
			newCPU = float64(config.HostConfig.CPUShares)
			newMemory = float64(config.HostConfig.Memory)
		}
	
		for _, task := range listTasks {
			if task.TaskClass == "1" {
				break
			}
			cutList = append(cutList, task)
			
			if fitAfterCuts(requestClass, host, newMemory, newCPU, cutList) {
				go cutRequests(cutList, host.HostIP)
				cutToReceive := amountToCut(requestClass, host.HostClass)
				return host, true, cutToReceive
			}
		}
	}
	return nil, false, 0.0
}

func cutRequests(cutList []Task, hostIP string) {
	for _, task := range cutList {
		currentCPU, _ := strconv.ParseFloat(task.CPU, 64) 
		currentMemory, _ := strconv.ParseFloat(task.Memory, 64) 
		amountToCut, _ := strconv.ParseFloat(task.CutToReceive,64)

		newCPU := currentCPU * (1 - amountToCut)
		newMemory := currentMemory * (1 - amountToCut)

		//these two will be used to update the amount of allocated cpu and memory for this host at the host registry
		cpuCutPerform := currentCPU - newCPU 
		memoryCutPerform := currentMemory - newMemory

		cpu := strconv.FormatInt(int64(newCPU),10)
		memory := strconv.FormatInt(int64(newMemory),10)
	
		//TODO: Para testes, isto depois é removido, no updatetask vai ser substituido pelo Up
		cut := strconv.FormatFloat(amountToCut, 'f', -1, 64)
		cpuCutPerformed := strconv.FormatFloat(cpuCutPerform, 'f',-1,64)
		memoryCutPerformed := strconv.FormatFloat(memoryCutPerform, 'f',-1,64)

		go UpdateTask("http://"+hostIP+":1234/task/updatetask/"+task.TaskClass+"&"+cpu+"&"+memory+"&"+task.TaskID+"&"+cut)
		go UpdateTask("http://"+ipHostRegistry+":12345/host/updatetask/"+task.TaskID+"&"+cpu+"&"+memory+"&"+hostIP+"&"+cpuCutPerformed+"&"+memoryCutPerformed)
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
		cut,_ := strconv.ParseFloat(task.CutToReceive, 64)
		taskCPU, _ := strconv.ParseFloat(task.CPU, 64)
		taskMemory, _ := strconv.ParseFloat(task.Memory, 64)
		cpuReduction += taskCPU * cut
		memoryReduction +=  taskMemory * cut
	}

	//after cutting the tasks, the host will have the following memory and cpu
	hostMemory := host.AllocatedMemory - memoryReduction
	hostCPU := host.AllocatedCPUs - cpuReduction
	
	//lets see if after those cuts the request will now fit
	return CheckOverbookingLimit(host, hostCPU, hostMemory, cpu, memory)
}

func amountToCut(requestClass string, hostClass string) (float64) {
	switch requestClass {
		case "2":
                	//Applying cut restrictions
                        if hostClass == "1" {
                                return (1 - MAX_CUT_CLASS2)
                        }else {
                                return 0.0
                        }
			break
                case "3":
                        if hostClass == "1" {
                                return (1 - MAX_CUT_CLASS3)
                        } else if hostClass == "2" {
                                cutItCanReceive := MAX_CUT_CLASS3 - MAX_CUT_CLASS2
                                return (1 - cutItCanReceive)
                        } else {
                                return 0.0
                        }
			break
                case "4":
                        if hostClass == "1" {
                                return (1 - MAX_CUT_CLASS4)
                        } else if hostClass == "2" {
                                cutItCanReceive := MAX_CUT_CLASS4 - MAX_CUT_CLASS2
                                return (1 - cutItCanReceive)
                        } else if hostClass == "3" {
                                cutItCanReceive := MAX_CUT_CLASS4 - MAX_CUT_CLASS3
                                return (1 - cutItCanReceive)
                        } else {
                                return 0.0
                        }
			break
	}
	return 0.0
}

//por enquanto esta float mas depois mudar para int
func applyCut(requestClass string, config *cluster.ContainerConfig, hostClass string) (float64, float64, bool){
	switch requestClass {
		case "2":
			 //Applying cut restrictions
			if hostClass == "1" {
				newMemory := float64(config.HostConfig.Memory) * (1 - MAX_CUT_CLASS2)
				newCPU := float64(config.HostConfig.CPUShares) * (1 - MAX_CUT_CLASS2)
				return newMemory, newCPU, true
			}else {
				return 0.0,0.0, false
			}
		case "3":
			if hostClass == "1" {
				newMemory := float64(config.HostConfig.Memory) * (1 - MAX_CUT_CLASS3)
				newCPU := float64(config.HostConfig.CPUShares) * (1 - MAX_CUT_CLASS3)
				return newMemory, newCPU, true
			} else if hostClass == "2" {
				cutItCanReceive := MAX_CUT_CLASS3 - MAX_CUT_CLASS2
				newMemory := float64(config.HostConfig.Memory) * (1 - cutItCanReceive)
				newCPU := float64(config.HostConfig.CPUShares) * (1 - cutItCanReceive)
				return newMemory, newCPU, true
			} else {
				return 0.0,0.0,false
			}
		case "4":
			if hostClass == "1" {
				newMemory := float64(config.HostConfig.Memory) * (1 - MAX_CUT_CLASS4)
				newCPU := float64(config.HostConfig.CPUShares) * (1 - MAX_CUT_CLASS4)
				return newMemory, newCPU, true
			} else if hostClass == "2" {
				cutItCanReceive := MAX_CUT_CLASS4 - MAX_CUT_CLASS2
				newMemory := float64(config.HostConfig.Memory) * (1 - cutItCanReceive)
				newCPU := float64(config.HostConfig.CPUShares) * (1 - cutItCanReceive)
				return newMemory, newCPU, true
			} else if hostClass == "3" {
				cutItCanReceive := MAX_CUT_CLASS4 - MAX_CUT_CLASS3
				newMemory := float64(config.HostConfig.Memory) * (1 - cutItCanReceive)
				newCPU := float64(config.HostConfig.CPUShares) * (1 - cutItCanReceive)
				return newMemory, newCPU, true
			} else {
				return 0.0, 0.0, false
			}
	}
	return 0.0,0.0, false
}

func afterCutRequestFits(requestClass string, host *Host, config *cluster.ContainerConfig) (bool) {

	newMemory, newCPU, canCut := applyCut(requestClass, config, host.HostClass)
	
	if !canCut {
		return false
	}
	return CheckOverbookingLimit(host, host.AllocatedCPUs, host.AllocatedMemory, newCPU, newMemory)	
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

