package node

import (
	"errors"
	"net/http"
	"bytes"
	"encoding/json"
	"net"	
	"fmt"

	"github.com/docker/swarm/cluster"
)

// Node is an abstract type used by the scheduler.
type Node struct {
	ID         string
	IP         string
	Addr       string
	Name       string
	Labels     map[string]string
	Containers cluster.Containers
	Images     []*cluster.Image

	UsedMemory  int64
	UsedCpus    int64
	TotalMemory int64
	TotalCpus   int64

	HealthIndicator int64
}


// NewNode creates a node from an engine.
func NewNode(e *cluster.Engine) *Node {
	//TODO: Para identificar o host à qual este worker pertence usar as labels
	//TODO: isto vai ser mudado de manager1 para manager
	//TODO: Isto só pode ser enviado uma vez, atualmente envia varias vezes
/*	if e.Name != "manager1" {
		url := "http://192.168.1.154:12345/host/addworker/1"
	

		req, err := http.NewRequest("GET", url, nil)
		req.Header.Set("X-Custom-Header", "myvalue")
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()
	}*/
	fmt.Println("Sending")
	fmt.Println(e.Name)
	if e.Name != "manager1" {
		//url := "http://192.168.1.168:12345/host/addworker/4&"+e.ID
		url := "http://146.193.41.142:12345/host/addworker/4&"+e.IP

/*		var jsonStr = []byte(`{	"ID:" `e.ID`,
		"IP":             `e.IP`,
		"Addr":           `e.Addr`,
		"Name":           `e.Name`,
		"Labels":          `e.Labels`,
		"Containers":      `e.Containers()`,
		"Images":          `e.Images()`,
		"UsedMemory":      `e.UsedMemory()`,
		"UsedCpus":        `e.UsedCpus()`,
		"TotalMemory":     `e.TotalMemory()`,
		"TotalCpus":       `e.TotalCpus()`,
		"HealthIndicator": `e.HealthIndicator()`}`)
*/		
		marshalled, _ := json.Marshal(Node{
		ID:              e.ID,
		IP:              e.IP,
		Addr:            e.Addr,
		Name:            e.Name,
		Labels:          e.Labels,
		Containers:      e.Containers(),
		Images:          e.Images(),
		UsedMemory:      e.UsedMemory(),
		UsedCpus:        e.UsedCpus(),
		TotalMemory:     e.TotalMemory(),
		TotalCpus:       e.TotalCpus(),
		HealthIndicator: e.HealthIndicator()})

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalled))
		req.Header.Set("X-Custom-Header", "myvalue")
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()
	}
	return &Node{
		ID:              e.ID,
		IP:              e.IP,
		Addr:            e.Addr,
		Name:            e.Name,
		Labels:          e.Labels,
		Containers:      e.Containers(),
		Images:          e.Images(),
		UsedMemory:      e.UsedMemory(),
		UsedCpus:        e.UsedCpus(),
		TotalMemory:     e.TotalMemory(),
		TotalCpus:       e.TotalCpus(),
		HealthIndicator: e.HealthIndicator(),
	}
}

// IsHealthy responses if node is in healthy state
func (n *Node) IsHealthy() bool {
	return n.HealthIndicator > 0
}

// Container returns the container with IDOrName in the engine.
func (n *Node) Container(IDOrName string) *cluster.Container {
	return n.Containers.Get(IDOrName)
}

// AddContainer injects a container into the internal state.
func (n *Node) AddContainer(container *cluster.Container) error {
	if container.Config != nil {
		memory := container.Config.HostConfig.Memory
		cpus := container.Config.HostConfig.CPUShares
		if n.TotalMemory-memory < 0 || n.TotalCpus-cpus < 0 {
			return errors.New("not enough resources")
		}
		n.UsedMemory = n.UsedMemory + memory
		n.UsedCpus = n.UsedCpus + cpus
	}
	n.Containers = append(n.Containers, container)
	return nil
}

func getIPAddress() string {
	addrs, err := net.InterfaceAddrs()
    	if err != nil {
        	fmt.Println(err.Error())
    	}
    	for _, a := range addrs {
        	if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            		if ipnet.IP.To4() != nil {
	    			//return ipnet.IP.String()
		//  return "192.168.1.4"
           		 }
		fmt.Println("IP")
		fmt.Println(ipnet.IP.String())
       		 }
    	}
    return "146.193.41.143"
}
