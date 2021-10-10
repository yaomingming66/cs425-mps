package mp1

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/bamboovir/cs425/lib/mp1/config"
	"github.com/bamboovir/cs425/lib/mp1/metrics"
	"github.com/bamboovir/cs425/lib/mp1/multicast"
	"github.com/bamboovir/cs425/lib/mp1/transaction"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logger = log.WithField("src", "main")
)

const (
	CONN_HOST = "0.0.0.0"
)

func ParsePort(portRawStr string) (port int, err error) {
	port, err = strconv.Atoi(portRawStr)
	if err != nil || (err == nil && port < 0) {
		return -1, fmt.Errorf("port should be positive number, received %s", portRawStr)
	}
	return port, nil
}

func ExitWrapper(err error) {
	if err != nil {
		logger.Errorf("command err: %v", err)
		os.Exit(1)
	}
}

func ConstructGroup(nodeID string, nodePort string, configPath string) (group *multicast.Group, err error) {
	nodesConfig, err := config.ConfigParser(configPath)
	if err != nil {
		return nil, err
	}

	members := make([]multicast.Node, 0)
	for _, configItem := range nodesConfig.ConfigItems {
		addr := net.JoinHostPort(configItem.NodeHost, configItem.NodePort)
		members = append(members, multicast.Node{
			ID:   configItem.NodeID,
			Addr: addr,
		})
	}

	addr := net.JoinHostPort(CONN_HOST, nodePort)
	group = multicast.NewGroupBuilder().
		WithSelfNodeID(nodeID).
		WithSelfNodeAddr(addr).
		WithMembers(members).
		AddMember(nodeID, addr).
		Build()
	return group, nil
}

func RootCMDMain(nodeID string, nodePort string, configPath string) (err error) {
	metrics.SetupMetrics()
	group, err := ConstructGroup(nodeID, nodePort, configPath)
	if err != nil {
		return err
	}
	router := group.TO()

	transactionProcessor := transaction.NewProcessor()
	transactionProcessor.RegisteTransactionHandler(router)

	err = group.Start(context.Background())

	if err != nil {
		return errors.Wrap(err, "group start failed")
	}
	transactionEventEmitter := transaction.TransactionEventListenerPipeline(os.Stdin)

	go func() {
		for msg := range transactionEventEmitter {
			err = group.TO().Multicast(msg.Path, msg.Body)
			if err != nil {
				logger.Errorf("%v", err)
				continue
			}
		}
	}()
	select {}
}

func NewRootCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mp1",
		Short: "mp1",
		Long:  "receive transactions events from the standard input (as sent by the generator) and send them to the decentralized node",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			nodeID := args[0]
			nodePort := args[1]
			configPath := args[2]

			err := RootCMDMain(nodeID, nodePort, configPath)
			ExitWrapper(err)
		},
	}

	return cmd
}
