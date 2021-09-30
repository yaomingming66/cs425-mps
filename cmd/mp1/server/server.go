package server

import (
	"bufio"
	"encoding/json"
	"net"

	"github.com/bamboovir/cs425/lib/mp1/transaction"
	"github.com/bamboovir/cs425/lib/mp1/types"
	log "github.com/sirupsen/logrus"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_TYPE = "tcp"
)

var (
	logger               = log.WithField("src", "server")
	transactionProcessor = transaction.NewTransactionProcessor()
)

func RunServer(nodeID string, port string) (err error) {
	addr := net.JoinHostPort(CONN_HOST, port)
	socket, err := net.Listen(CONN_TYPE, addr)

	if err != nil {
		logger.Errorf("node [%s] error listening: %v", nodeID, err)
		return err
	}

	defer socket.Close()
	logger.Infof("node [%s] success listening on: %s", nodeID, addr)
	for {
		conn, err := socket.Accept()
		if err != nil {
			logger.Errorf("node [%s] error accepting: %v", nodeID, err)
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
			logger.Errorf("unrecognized event message, except hi")
			return
		}
		logger.Infof("node [%s] connected", hi.From)
	}

	for scanner.Scan() {
		line := scanner.Text()
		msg, err := types.DecodePolymorphicMessage([]byte(line))
		if err != nil {
			logger.Errorf("decode message failed: %v", err)
			continue
		}

		switch v := msg.(type) {
		case *types.Deposit:
			logger.Infof("deposit: %s %d", v.Account, v.Amount)
			_ = transactionProcessor.Deposit(v.Account, v.Amount)
			logger.Info(transactionProcessor.BalancesSnapshotStdString())
		case *types.Transfer:
			logger.Infof("tranfer: %s -> %s %d", v.FromAccount, v.ToAccount, v.Amount)
			_ = transactionProcessor.Transfer(v.FromAccount, v.ToAccount, v.Amount)
			logger.Info(transactionProcessor.BalancesSnapshotStdString())
		default:
			logger.Errorf("unrecognized event message")
		}
	}

	err := scanner.Err()

	if err != nil {
		logger.Errorf("node [%s] connection err: %v", hi.From, err)
	} else {
		logger.Infof("node [%s] connection reach EOF", hi.From)
	}
}
