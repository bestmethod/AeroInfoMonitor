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

// so we can just do logger.this and logger.that
var logger Logger.Logger

// golang entrypoint
func main() {
	// we like structs and objects, lets have a struct-main :)
	m := mainStruct{}
	m.main()
}

// this is the real entry point
func (m *mainStruct) main() {
	m.setLogger()
	m.osArgs()
	m.mainLoop()
}

// the main loop which never ends - the actual monitoring loop
func (m *mainStruct) mainLoop() {
	for {
		m.connect()                  // will only connect if not yet connected, logic is there
		nodes := m.client.GetNodes() // get a list of nodes

		// execute for each node
		for _, node := range nodes {

			// we will need them later, cluster size, tx pending and rx pending
			// yes, we want to declare them this way, because if something wrong goes on, we should still have them and print what we got
			cs := ""
			tx := ""
			rx := ""

			// some basic stuff we can log, because we can
			nodeName := node.GetName()
			nodeHost := node.GetHost().String()
			nodeActive := node.IsActive()

			// lets issue actual Info()
			info, err := node.RequestInfo(fmt.Sprintf("namespace/%s", m.ns))
			if err != nil {
				// wops, log that Info() failed, reason and anything else we gathered
				logger.Error("InfoGetError,node=%s,host=%s,nodeActive=%t,clientNodesConnected=%d,err=%s", nodeName, nodeHost, nodeActive, len(nodes), err)
			} else {

				// now some play with strings, separate results
				ndts := strings.Split(info[fmt.Sprintf("namespace/%s",m.ns)], ";")

				// go through each separate result and find the 3 we care about (cs, tx, rx)
				// if we found all 3 already, break out of this loop, no point wasting precious milliseconds
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

				// lets log success and all the data we got
				logger.Info("InfoGetSuccess,node=%s,host=%s,nodeActive=%t,clientNodesConnected=%d,%s,%s,%s", nodeName, nodeHost, nodeActive, len(nodes), cs, tx, rx)
			}
		}

		// short snooze ;)
		time.Sleep(50 * time.Millisecond)
	}
}

// process the os.Args[] arguments into the struct with what the user filled in. Provide usage if needed.
func (m *mainStruct) osArgs() {
	var err error
	if len(os.Args) < 4 || len(os.Args) == 5 || len(os.Args) > 6 {
		logger.Fatalf(1, "Incorrect usage.\nUsage: %s NodeIP NodePORT NamespaceName [username] [password]", os.Args[0])
	}

	m.nodeIp = os.Args[1]
	m.nodePort, err = strconv.Atoi(os.Args[2])
	m.ns = os.Args[3]
	if err != nil {
		logger.Fatalf(2, "Port number incorrect: %s", os.Args[2])
	}
	if len(os.Args) == 6 {
		m.user = os.Args[4]
		m.pass = os.Args[5]
		m.policy = as.NewClientPolicy()
		m.policy.User = m.user
		m.policy.Password = m.pass
		m.policy.Timeout = 10 * time.Second
	}
}

// set the logger and configure it
func (m *mainStruct) setLogger() {
	logger = Logger.Logger{}
	err := logger.Init("", "AeroInfoMonitor", Logger.LEVEL_INFO|Logger.LEVEL_DEBUG, Logger.LEVEL_WARN|Logger.LEVEL_ERROR|Logger.LEVEL_CRITICAL, Logger.LEVEL_NONE)
	logger.TimeFormat("Jan 02 15:04:05.000000-0700")
	if err != nil {
		log.Fatalf("Logger init failed: %s", err)
	}
}

// connect to aerospike, will do up to 5 attempts at 100ms interval and die if no success. only do it IF not connected
func (m *mainStruct) connect() {
	var err error
	if m.client == nil || m.client.IsConnected() == false {
		for range []int{1, 2, 3, 4, 5} {

			logger.Info("ConnectingToCluster")
			if m.policy == nil {
				m.client, err = as.NewClient(m.nodeIp, m.nodePort)
			} else {
				m.client, err = as.NewClientWithPolicy(m.policy, m.nodeIp, m.nodePort)
			}
			if err != nil {
				logger.Error("ClientConnectFailed: %s", err)
				time.Sleep(100 * time.Millisecond)
			} else {
				logger.Info("ClientConnected")
				return
			}
		}
		logger.Fatalf(3, "Client connect failed after 5 attempts: %s", err)
	}
}

// the jewel ;) Yes, the struct for real main()
type mainStruct struct {
	nodeIp   string
	nodePort int
	client   *as.Client
	ns       string
	user     string
	pass     string
	policy   *as.ClientPolicy
}
