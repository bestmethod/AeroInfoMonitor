package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)
import "github.com/bestmethod/go-logger"
import as "github.com/aerospike/aerospike-client-go"

var logger Logger.Logger

func main() {
	m := mainStruct{}
	m.main()
}

func (m *mainStruct) main() {
	logger = Logger.Logger{}
	err := logger.Init("","AeroInfoMonitor",Logger.LEVEL_INFO|Logger.LEVEL_DEBUG,Logger.LEVEL_WARN|Logger.LEVEL_ERROR|Logger.LEVEL_CRITICAL,Logger.LEVEL_NONE)
	logger.TimeFormat("Jan 02 15:04:05.000000-0700")
	if err != nil {
		log.Fatalf("Logger init failed: %s",err)
	}
	if len(os.Args) < 4 || len(os.Args) == 5 || len(os.Args) > 6 {
		logger.Fatalf(1,"Incorrect usage.\nUsage: %s NodeIP NodePORT NamespaceName [username] [password]",os.Args[0])
	}

	m.nodeIp = os.Args[1]
	m.nodePort, err = strconv.Atoi(os.Args[2])
	m.ns = os.Args[3]
	if err != nil {
		logger.Fatalf(2,"Port number incorrect: %s",os.Args[2])
	}
	if len(os.Args) == 6 {
		m.user = os.Args[4]
		m.pass = os.Args[5]
		m.policy = as.NewClientPolicy()
		m.policy.User = m.user
		m.policy.Password = m.pass
		m.policy.Timeout = 10*time.Second
	}

	m.connect()

	for {
		if m.client.IsConnected() == false {
			m.connect()
		}
		nodes := m.client.GetNodes()
		for _,node := range nodes {
			cs := ""
			tx := ""
			rx := ""
			nodeName := node.GetName()
			nodeHost := node.GetHost().String()
			nodeActive := node.IsActive()
			info, err := node.RequestInfo(fmt.Sprintf("namespace/%s", m.ns))
			if err != nil {
				logger.Error("InfoGetError,node=%s,host=%s,nodeActive=%t,clientNodesConnected=%d,err=%s", nodeName, nodeHost, nodeActive, len(nodes), err)
			} else {
				// get cs,tx,rx
				ndts := strings.Split(info["namespace/test"], ";")
				found := 0
				for _, ndt := range ndts {
					if strings.HasPrefix(ndt, "ns_cluster_size=") {
						cs = ndt
						found = found + 1
					} else if strings.HasPrefix(ndt, "migrate_tx_partitions_remaining=") {
						tx = ndt
						found = found + 1
					} else if strings.HasPrefix(ndt, "migrate_rx_partitions_remaining=") {
						rx = ndt
						found = found + 1
					}
					if found == 3 {
						break
					}
				}
				logger.Info("InfoGetSuccess,node=%s,host=%s,nodeActive=%t,clientNodesConnected=%d,%s,%s,%s", nodeName, nodeHost, nodeActive, len(nodes), cs, tx, rx)
			}
		}
		time.Sleep(50*time.Millisecond)
	}
}

func (m *mainStruct) connect() {
	var err error
	for range []int{1,2,3,4,5} {

		logger.Info("ConnectingToCluster")
		if m.policy == nil {
			m.client, err = as.NewClient(m.nodeIp, m.nodePort)
		} else {
			m.client, err = as.NewClientWithPolicy(m.policy, m.nodeIp, m.nodePort)
		}
		if err != nil {
			logger.Error("ClientConnectFailed: %s",err)
			time.Sleep(100*time.Millisecond)
		} else {
			logger.Info("ClientConnected")
			return
		}
	}
	logger.Fatalf(3, "Client connect failed after 5 attempts: %s",err)
}

type mainStruct struct {
	nodeIp string
	nodePort int
	client *as.Client
	ns string
	user string
	pass string
	policy *as.ClientPolicy
}
