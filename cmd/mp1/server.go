package mp1

import (
	"bufio"
	"encoding/json"
	"net"

	"github.com/bamboovir/cs425/lib/mp1/types"
	log "github.com/sirupsen/logrus"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_TYPE = "tcp"
)

func RunServer(portStr string) (err error) {
	addr := net.JoinHostPort(CONN_HOST, portStr)
	socket, err := net.Listen(CONN_TYPE, addr)

	if err != nil {
		log.Errorf("error listening: %v", err)
		return err
	}

	defer socket.Close()
	log.Infof("listening on: %s", addr)
	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Errorf("error accepting: %v", err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	hi := &types.Hi{}
	if scanner.Scan() {
		firstLine := scanner.Text()
		err := json.Unmarshal([]byte(firstLine), hi)
		if err != nil || hi.From == "" {
			log.Errorf("unrecognized event message, except hi")
			return
		}
	}

	log.Infof("node[%s] connected", hi.From)

	for scanner.Scan() {
		line := scanner.Text()
		log.Info(line)
		msg, err := types.DecodePolymorphicMessage([]byte(line))
		if err != nil {
			log.Errorf("decode message failed: %v", err)
			continue
		}

		switch v := msg.(type) {
		case *types.Deposit:
			log.Infof("deposit: %s %d", v.Account, v.Amount)
		case *types.Transfer:
			log.Infof("tranfer: %s -> %s %d", v.FromAccount, v.ToAccount, v.Amount)
		default:
			log.Errorf("unrecognized event message")
		}
	}

	err := scanner.Err()

	if err != nil {
		log.Errorf("node[%s] connection err: %v", hi.From, err)
	} else {
		log.Infof("node[%s] connection reach EOF", hi.From)
	}
}
