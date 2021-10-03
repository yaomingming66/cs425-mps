package multicast

import (
	"net"

	"bufio"
	"encoding/json"

	"github.com/bamboovir/cs425/lib/mp1/broker"
	"github.com/bamboovir/cs425/lib/mp1/types"
	log "github.com/sirupsen/logrus"
)

// var (
// 	transactionProcessor = transaction.NewTransactionProcessor()
// )

var (
	serverLogger = log.WithField("src", "multicast.server")
)

const (
	CONN_TYPE = "tcp"
)

func RunServer(nodeID string, addr string, out *broker.Broker) (err error) {
	socket, err := net.Listen(CONN_TYPE, addr)

	if err != nil {
		serverLogger.Errorf("node [%s] error listening: %v", nodeID, err)
		return err
	}

	defer socket.Close()
	serverLogger.Infof("node [%s] success listening on: %s", nodeID, addr)
	for {
		conn, err := socket.Accept()
		if err != nil {
			serverLogger.Errorf("node [%s] error accepting: %v", nodeID, err)
			continue
		}

		go handleConn(conn, out)
	}
}

func handleConn(conn net.Conn, out *broker.Broker) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	hi := &types.Hi{}
	if scanner.Scan() {
		firstLine := scanner.Text()
		err := json.Unmarshal([]byte(firstLine), hi)
		if err != nil || hi.From == "" {
			serverLogger.Errorf("unrecognized event message, except hi")
			return
		}
		serverLogger.Infof("node [%s] connected", hi.From)
	}

	for scanner.Scan() {
		line := scanner.Text()
		msg, err := DecodeMsg([]byte(line))
		if err != nil {
			serverLogger.Errorf("decode message failed: %v", err)
			continue
		}

		out.Publish(*msg)
	}

	err := scanner.Err()

	if err != nil {
		if hi.From != "" {
			serverLogger.Errorf("node [%s] connection err: %v", hi.From, err)
		}
	} else {
		if hi.From != "" {
			serverLogger.Infof("node [%s] connection reach EOF", hi.From)
		}
	}
}
