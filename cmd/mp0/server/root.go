package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/bamboovir/cs425/lib/mp0/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_TYPE = "tcp"
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
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			portRawStr := args[0]
			log.Infof("try start server in port %s", portRawStr)
			err = TCPToMath(portRawStr)
			return err
		},
	}

	return cmd
}

func TCPToMath(portStr string) (err error) {
	// Listen for incoming connections.
	addr := net.JoinHostPort(CONN_HOST, portStr)
	socket, err := net.Listen(CONN_TYPE, addr)
	if err != nil {
		log.Errorf("Error listening: %v", err.Error())
		return err
	}
	// Close the listener when the application closes.
	defer socket.Close()
	log.Infof("Listening on: %s", addr)
	for {
		// Listen for an incoming connection.
		conn, err := socket.Accept()
		connectionEstablishedTimestamp := timeNow()
		if err != nil {
			log.Errorf("Error accepting: %v", err.Error())
			continue
		}

		// Handle connections in a new goroutine.
		go handleRequest(conn, connectionEstablishedTimestamp)
	}

}

// Handles incoming requests.
func handleRequest(conn net.Conn, connectionEstablishedTimestamp float64) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	var msg types.Msg

	isFirst := true
	var currTimestamp float64
	for {
		err := decoder.Decode(&msg)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("%f - %s disconnected\n", timeNow(), msg.From)
				log.Errorf("socket closed reach EOF: %v", err)
				break
			}
			log.Errorf("decode message failed: %v", err)
			break
		}

		currTimestamp = timeNow()
		if isFirst {
			fmt.Printf("%f - %s connected\n", connectionEstablishedTimestamp, msg.From)
			isFirst = false
		}

		fmt.Printf("%f %s %s\n", currTimestamp, msg.From, msg.Payload)
		timeSendFloat, err := strconv.ParseFloat(msg.TimeStamp, 64)
		if err != nil {
			log.Infof("parse float timestamp failed : %v, skip", err)
			continue
		}
		payloadSizeInBytes := len(msg.Payload)
		res := types.Params{
			Delay:     strconv.FormatFloat(currTimestamp-timeSendFloat, 'E', -1, 64),
			TimeStamp: strconv.FormatFloat(currTimestamp, 'E', -1, 64),
			Size:      strconv.Itoa(payloadSizeInBytes),
		}

		resbytes, err := json.Marshal(res)
		if err != nil {
			log.Errorf("marshal message failed %v, skip", err)
			continue
		}
		log.Infof("%s", string(resbytes))
	}
}

func timeNow() float64 {
	return float64(time.Now().UnixNano()) / float64((time.Second))
}
