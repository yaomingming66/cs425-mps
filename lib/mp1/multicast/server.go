package multicast

import (
	"net"

	"bufio"
	"encoding/json"

	"github.com/bamboovir/cs425/lib/mp1/metrics"
	"github.com/bamboovir/cs425/lib/mp1/types"
	log "github.com/sirupsen/logrus"
)

var (
	serverLogger = log.WithField("src", "multicast.server")
)

const (
	CONN_TYPE = "tcp"
)

func runServer(nodeID string, addr string, out chan []byte) (err error) {
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

		go handleConn(nodeID, conn, out)
	}
}

func handleConn(nodeID string, conn net.Conn, out chan []byte) {
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
		lineBytes := []byte(line)
		metrics.NewBandwidthLogEntry(nodeID, len(lineBytes)).Log()
		out <- lineBytes
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
