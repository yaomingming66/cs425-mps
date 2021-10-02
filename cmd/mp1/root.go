package mp1

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/bamboovir/cs425/lib/mp1/config"
	"github.com/bamboovir/cs425/lib/mp1/multicast"
	"github.com/bamboovir/cs425/lib/mp1/transaction"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logger = log.WithField("src", "main")
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
		logger.Errorf("err: %v", err)
		os.Exit(1)
	}
}

func NewRootCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mp1",
		Short: "mp1",
		Long:  "receive transactions events from the standard input (as sent by the generator) and send them to the decentralized node",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			nodeID := args[0]
			port := args[1]
			configPath := args[2]

			nodesConfig, err := config.ConfigParser(configPath)
			ExitWrapper(err)

			nodesConfig.ConfigItems = append(nodesConfig.ConfigItems, config.ConfigItem{
				NodeID:   nodeID,
				NodeHost: multicast.CONN_HOST,
				NodePort: port,
			})

			group, err := multicast.NewGroup(context.Background(), nodeID, port, *nodesConfig)
			ExitWrapper(err)

			transactionEventEmitter := transaction.TransactionEventListenerPipeline(os.Stdin)

			group.RegisterRDeliver()

			go func() {
				for msg := range transactionEventEmitter {
					group.RMulticast(msg)
				}
			}()

			rDeliverChannel := group.RDeliver()
			go func() {
				for msg := range rDeliverChannel {
					logger.Infof("%s -> %s", msg.SrcID, string(msg.Body))
				}
			}()

			err = group.Wait()
			ExitWrapper(err)

		},
	}

	return cmd
}
