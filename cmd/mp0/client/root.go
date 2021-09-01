package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/bamboovir/cs425/lib/mp0/types"
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
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]
			host := args[1]
			port := args[2]
			addr := fmt.Sprintf("%s:%s", host, port)
			err = CopyStdinToTCP(name, addr)
			return err
		},
	}

	return cmd
}

func CopyStdinToTCP(name string, addr string) (err error) {
	log.Infof("node [%s] tries to connect to the server [%s]", name, addr)
	client, err := net.Dial("tcp", addr)
	if err != nil {
		log.Errorf("node [%s] failed to connect to the server [%s]", name, addr)
		return err
	}
	defer client.Close()

	encoder := json.NewEncoder(client)
	scanner := bufio.NewScanner(os.Stdin)

	// const maxCapacity = longLineLen
	// buf := make([]byte, maxCapacity)
	// scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		tmp := strings.Fields(line)
		if len(tmp) != 2 {
			log.Errorf("invalid input format, skip")
			continue
		}
		timestamp, payload := tmp[0], tmp[1]
		msg := types.Msg{
			TimeStamp: timestamp,
			Payload:   payload,
			From:      name,
		}
		err := encoder.Encode(msg)
		if err != nil {
			log.Errorf("encode message failed %v", err)
			return err
		}
	}
	return nil
}
