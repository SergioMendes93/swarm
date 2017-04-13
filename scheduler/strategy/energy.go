package strategy

import (
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
	"github.com/docker/swarm/scheduler/filter"

	"math"
//	"os/exec"
//	"math/rand"	
	"net/http"	
	"encoding/json"
	"strconv"
	"fmt"
	"net"
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
var ipTaskRegistry = "146.193.41.143"
//var ipAddress = getIPAddress()

var MAX_OVERBOOKING_CLASS1 = 1.0
var MAX_OVERBOOKING_CLASS2 = 1.2
var MAX_OVERBOOKING_CLASS3 = 1.5
var MAX_OVERBOOKING_CLASS4 = 2.0

var MAX_CUT_CLASS2 = 0.0
var MAX_CUT_CLASS3 = 0.0
var MAX_CUT_CLASS4 = 0.0

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
/*
func findNode(host *Host, nodes []*node.Node) ([]*node.Node) {
	output := make([]*node.Node,0)

	//going to choose a worker randomly from the host
	numWorkers := len(host.WorkerNodes)
	
	seed := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(seed)
	randomNumber := r1.Intn(numWorkers)		
		
	if randomNumber != 0 {
		randomNumber = randomNumber - 1
	}

	output = append(output, host.WorkerNodes[randomNumber])
	
	return output
}
*/
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
func (p *EnergyPlacementStrategy) RankAndSort(config *cluster.ContainerConfig, nodes []*node.Node) ([]*node.Node, error, string, string, bool) {

//	ipHostRegistry = getIPAddress()
	/*cmd := "docker"
	args := []string{"run", "-itd", "-c", "30", "-m", "500m", "-e", "affinity:requestclass==4", "--name", "lala1", "ubuntu"}

	if err := exec.Command(cmd, args...).Run(); err != nil {
		fmt.Println("Error using docker run")
		fmt.Println(err)
	}*/

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
		return nil, err, "0", requestType, false
	}

	//+1 is the list type go get +2 is the other list type //see hostregistry.go for +info.. //endereço do manager e port do host registry
	listHostsLEE_DEE := GetHosts("http://"+ipHostRegistry+":12345/host/list/"+requestClass+"&1")
		
	
//------------ Scheduling without cuts and kills----------------------------------------------------------

	for i := 0; i < len(listHostsLEE_DEE); i++ {
		//we must check if after allocating this request the overbooking factor of the host doesnt go over the host maximum overbooking allowed
		//if it does no go over it, then the request can be scheduled to this host
		host := listHostsLEE_DEE[i]
		fmt.Println(host)
		if CheckOverbookingLimit(host, host.AllocatedCPUs, host.AllocatedMemory, float64(config.HostConfig.Memory), float64(config.HostConfig.CPUShares)) {
			//return findNode(host,nodes), nil, requestClass, requestType, false
			output = append(output, host.WorkerNodes[0])
			return output, nil, requestClass, requestType, false
		}
			
	}
//------------------------ Cut algorithm ------------------------

	//obtemos a nova listHostsLEE_DEE, desta vez ordenada de forma diferente	
	listHostsLEE_DEE = GetHosts("http://ipHostRegistry:12345/host/list/"+requestClass+"&2")
	
	host, allocable, cut := cut(listHostsLEE_DEE, requestClass, config)
	if allocable { //if true then it means that we have a host that it can be scheduled	
		//return findNode(host, nodes), nil, cut, requestType, true
		output = append(output, host.WorkerNodes[0])
		return output, nil, cut, requestType, true
	}

	return nodes, nil, requestClass, requestType, false

//-------------------------Kill algorithm ----------------------------

	listHostsEED_DEE := GetHosts("http://"+ipHostRegistry+":12345/host/listkill/"+requestClass)

	host, allocable = kill(listHostsEED_DEE, requestClass, requestType, config)
	
	if allocable {
		//return findNode(host,nodes), nil, requestClass, requestType,  false
		output = append(output, host.WorkerNodes[0])
		return output, nil, requestClass, requestType,  false
	}

//	return nodes, nil, requestClass, requestType, false
	return nil, nil, requestClass, requestType, false //can't be scheduled É PARA FICAR ESTE QUANDO FOR DEFINITIVO
}


func kill(listHostsEED_DEE []*Host, requestClass string, requestType string, config *cluster.ContainerConfig) (*Host, bool) {
	for i := 0 ; i < len(listHostsEED_DEE); i++ {
		possibleKillList := make([]Task, 0)
		host := listHostsEED_DEE[i] 

		if requestClass == "4" && requestType == "job"{
			possibleKillList = append(possibleKillList, GetTasks("http://"+ipTaskRegistry+":1234/task/class4")...)

		} else if requestClass == "4" {
			return nil, false
		} else {
			possibleKillList = append(possibleKillList, GetTasks("http://"+ipTaskRegistry+":1234/task/higher/" + requestClass)...)
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
		go UpdateTask("http://"+ipHostRegistry+":12345/host/reschedule/"+task.CPU+"&"+task.Memory+"&"+task.TaskClass+"&"+task.Image)
	}
}

func killTasks(killList []Task) {
	for _, task := range killList {
		go UpdateTask("http://"+ipHostRegistry+":12345/host/killtask/"+task.TaskID)
		go UpdateTask("http://"+ipTaskRegistry+":1234/task/remove/"+task.TaskID)
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

func cut(listHostsLEE_DEE []*Host, requestClass string, config *cluster.ContainerConfig) (*Host, bool, string) {
		
	for i := 0; i < len(listHostsLEE_DEE); i++ {
		cutList := make([]Task,0)
		listTasks := make([]Task,0)

		host := listHostsLEE_DEE[i]		

		if host.HostClass >= requestClass && requestClass != "4" {
			//listTasks = append(listTasks, GetTasks("http://" + host.HostIP + ":1234/task/highercut/" + requestClass)
			listTasks = append(listTasks, GetTasks("http://"+ipTaskRegistry+":1234/task/highercut/" + requestClass)...)
		} else if requestClass != "1" && afterCutRequestFits(requestClass, host, config){
			return host, true, requestClass	//requestClass indicates the cut received, performed at /cluster/swarm/cluster.go
		} else if requestClass != "1" {
			listTasks = append(listTasks, GetTasks("http://"+ipTaskRegistry+":1234/task/equalhigher/" + requestClass+"&"+host.HostClass)...)
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
				go cutRequests(cutList)
				fmt.Println("Cut done") 
				return host, true, requestClass
			}
		}
		//return host, false, requestClass 
	}
	//TODO: este return esta assim para efeitos de teste, depois repensar nisto
	return listHostsLEE_DEE[0], false, requestClass
}

func cutRequests(cutList []Task) {
	for _, task := range cutList {
		newCPU, err := strconv.ParseFloat(task.CPU, 64) 
		newMemory, err := strconv.ParseFloat(task.Memory, 64) 
		amountToCut, err := strconv.ParseFloat(task.CutToReceive,64)

		newCPU *= (1 - amountToCut)
		newMemory *= (1 - amountToCut)

		if err != nil {
			fmt.Println(err)
		}
		cpu := strconv.FormatFloat(newCPU, 'f', -1, 64)
		memory := strconv.FormatFloat(newMemory, 'f', -1, 64)
	
		//TODO: Para testes, isto depois é removido, no updatetask vai ser substituido pelo Up
		cut := strconv.FormatFloat(amountToCut, 'f', -1, 64)

		go UpdateTask("http://"+ipTaskRegistry+":1234/task/updatetask/"+task.TaskClass+"&"+cpu+"&"+memory+"&"+task.TaskID+"&"+cut)
		go UpdateTask("http://"+ipHostRegistry+":12345/host/updatetask/"+task.TaskID+"&"+cpu+"&"+memory)
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

func getIPAddress() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        fmt.Println(err.Error())
    }
    for _, a := range addrs {
        if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return ""
}
