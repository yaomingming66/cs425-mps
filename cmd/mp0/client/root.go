package client

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
		Use:   "mp0-c",
		Short: "mp0-c",
		Long:  "receive events from the standard input (as sent by the generator) and send them to the centralized logger.",
		Args: func(cmd *cobra.Command, args []string) error {
			n := 2
			if len(args) != n {
				return fmt.Errorf("accepts %d arg(s), received %d", n, len(args))
			}
			portRawStr := args[1]
			_, err := ParsePort(portRawStr)
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			host := args[0]
			portRawStr := args[1]
			port, err := ParsePort(portRawStr)
			if err != nil {
				return err
			}

			log.Infof("try connect server in %s:%d", host, port)
			return nil
		},
	}

	return cmd
}
