package server

import (
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
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
		Use:   "mp0-s",
		Short: "mp0-s",
		Long:  "centralized logger start by listening on a port, specified on a command line, and allow nodes to connect to it and start sending it events. It should then print out the events, along with the name of the node sending the events, to standard out. diagnostic messages are sent to stderr",
		Args: func(cmd *cobra.Command, args []string) error {
			n := 1
			if len(args) != n {
				return fmt.Errorf("accepts %d arg(s), received %d", n, len(args))
			}
			portRawStr := args[0]
			_, err := ParsePort(portRawStr)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			portRawStr := args[0]
			port, err := ParsePort(portRawStr)
			if err != nil {
				return err
			}
			log.Infof("try start server in port %d", port)
			return nil
		},
	}

	return cmd
}
