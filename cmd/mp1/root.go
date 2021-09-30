package mp1

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/bamboovir/cs425/cmd/mp1/client"
	"github.com/bamboovir/cs425/cmd/mp1/server"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func ParsePort(portRawStr string) (port int, err error) {
	port, err = strconv.Atoi(portRawStr)
	if err != nil || (err == nil && port < 0) {
		return -1, fmt.Errorf("port should be positive number, received %s", portRawStr)
	}
	return port, nil
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
			errG, _ := errgroup.WithContext(context.Background())
			errG.Go(
				func() error {
					return server.RunServer(nodeID, port)
				},
			)

			go client.RunClients(nodeID, port, configPath)

			err := errG.Wait()
			if err != nil {
				os.Exit(1)
			}
		},
	}

	return cmd
}
