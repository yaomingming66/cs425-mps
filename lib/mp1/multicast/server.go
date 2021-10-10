package multicast

import (
	"net"

	sync "github.com/sasha-s/go-deadlock"

	"bufio"
	"encoding/json"

	"github.com/bamboovir/cs425/lib/mp1/metrics"
	"github.com/bamboovir/cs425/lib/mp1/router"
	"github.com/bamboovir/cs425/lib/mp1/types"
	log "github.com/sirupsen/logrus"
)

var (
	serverLogger = log.WithField("src", "multicast.server")
)

const (
	CONN_TYPE = "tcp"
)

const (
	_      = iota // ignore first value by assigning to blank identifier
	KB int = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
)

func startServer(nodeID string, addr string, router *router.Router) (socket net.Listener, err error) {
	socket, err = net.Listen(CONN_TYPE, addr)

	if err != nil {
		serverLogger.Errorf("node [%s] error listening: %v", nodeID, err)
		return nil, err
	}

	serverLogger.Infof("node [%s] success listening on: %s", nodeID, addr)
	return socket, nil
}

func runServer(startSyncWaitGroup *sync.WaitGroup, nodeID string, socket net.Listener, router *router.Router) {
	defer socket.Close()
	for {
		conn, err := socket.Accept()
		if err != nil {
			serverLogger.Errorf("node [%s] error accepting: %v", nodeID, err)
			continue
		}

		go handleConn(startSyncWaitGroup, nodeID, conn, router)
	}
}

func handleConn(startSyncWaitGroup *sync.WaitGroup, nodeID string, conn net.Conn, router *router.Router) {
	defer conn.Close()
	tcpConn := conn.(*net.TCPConn)

	tcpConn.SetReadBuffer(5 * MB)
	tcpConn.SetWriteBuffer(5 * MB)

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

	// wait for all client ready
	startSyncWaitGroup.Wait()

	for scanner.Scan() {
		line := scanner.Text()
		lineBytes := []byte(line)
		metrics.NewBandwidthLogEntry(nodeID, len(lineBytes)).Log()

		msg := &BMsg{}
		_, err := msg.Decode(lineBytes)
		if err != nil {
			serverLogger.Errorf("server decode msg failed: %v", err)
		}
		err = router.Run(BMulticastPath, msg)
		if err != nil {
			serverLogger.Errorf("server process msg failed: %v", err)
		}
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
