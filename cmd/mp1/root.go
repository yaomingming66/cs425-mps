package mp1

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
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
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			nodeIDStr := args[0]
			portStr := args[1]
			configPathStr := args[2]
			go RunServer(portStr)
			go RunClients(nodeIDStr, configPathStr)
			select {}
		},
	}

	return cmd
}
