package server

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/bamboovir/cs425/lib/mp0/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	CONN_HOST = "127.0.0.1"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			portRawStr := args[0]
			log.Infof("try start server in port %s", portRawStr)
			TCPToMath(portRawStr)
			return nil
		},
	}

	return cmd
}

func TCPToMath(portStr string) {
	// Listen for incoming connections.
	addr := net.JoinHostPort(CONN_HOST, portStr)
	socket, err := net.Listen(CONN_TYPE, addr)
	if err != nil {
		log.Errorf("Error listening: %v", err.Error())
		return
	}
	// Close the listener when the application closes.
	defer socket.Close()
	log.Infof("Listening on: %s", addr)
	for {
		// Listen for an incoming connection.
		conn, err := socket.Accept()
		if err != nil {
			log.Errorf("Error accepting: %v", err.Error())
			return
		}

		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	d := json.NewDecoder(conn)
	var msg types.Msg

	fmt.Printf("%f - %s connected", timeNow(), msg.From)

	for {
		if d.Decode(&msg) != nil {
			log.Errorf("socket closed or decode message failed")
			break
		}
		fmt.Printf("%f %s\n%s\n", timeNow(), msg.From, msg.Payload)

		timeNow := timeNow()
		timeSendFloat, err := strconv.ParseFloat(msg.TimeStamp, 64)
		if err != nil {
			log.Info("Error reading:", err.Error())
			break
		}
		var sendPayloadBytes int = utf8.RuneCountInString(msg.Payload)
		res := types.Params{
			Delay:     strconv.FormatFloat(timeNow-timeSendFloat, 'E', -1, 64),
			TimeStamp: strconv.FormatFloat(timeNow, 'E', -1, 64),
			Size:      strconv.Itoa(sendPayloadBytes),
		}

		resbytes, err := json.Marshal(res)
		if err != nil {
			log.Errorf("marshal message failed %v", err)
			break
		}
		log.Infof("%s", string(resbytes))

		// Send a response back to person contacting us.
		conn.Write([]byte("message received."))
	}
	fmt.Printf("%f - %s disconnected", timeNow(), msg.From)

	// Close the connection when you're done with it.
	conn.Close()
}

func timeNow() float64 {
	return float64(time.Now().UnixNano()) / float64(time.Second)
}
