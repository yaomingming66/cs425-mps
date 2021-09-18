package mp1

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bamboovir/cs425/lib/mp1/config"
	"github.com/bamboovir/cs425/lib/mp1/types"
	log "github.com/sirupsen/logrus"
)

func RunClients(nodeID string, configPath string) (err error) {
	nodesConfig, err := config.ConfigParser(configPath)
	if err != nil {
		return err
	}

	for i := range nodesConfig.ConfigItems {
		c := nodesConfig.ConfigItems[i]
		log.Info(c.NodeID, " ", c.NodeHost, " ", c.NodePort)
	}

	ReadStdin()

	return nil
}

func ReadStdin() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		eventMsg, err := EncodeInputMsg(line)
		if err != nil {
			log.Error("encode input msg failed, skip")
			continue
		}
		log.Info(string(eventMsg))
	}

	err := scanner.Err()

	if err != nil {
		log.Errorf("read stdin err: %v", err)
	} else {
		log.Errorf("stdin reach EOF")
	}
}

func RunClient(nodeID string, addr string) (err error) {
	log.Infof("node [%s] tries to connect to the server [%s]", nodeID, addr)
	client, err := net.DialTimeout("tcp", addr, time.Second*10)
	if err != nil {
		log.Errorf("node [%s] failed to connect to the server [%s]", nodeID, addr)
		return err
	}

	defer client.Close()
	return nil
}

const (
	DepositEvent  = "DEPOSIT"
	TransferEvent = "TRANSFER"
)

func EncodeInputMsg(msg string) (encoded []byte, err error) {
	fields := strings.Fields(msg)
	if len(fields) == 0 {
		errMsg := "invalid event format, empty fields"
		log.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	eventType := fields[0]

	switch eventType {
	case DepositEvent:
		if len(fields) != 3 {
			errMsg := "invalid deposit event format"
			log.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}

		amount, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}

		deposit := types.Deposit{
			Account: fields[1],
			Amount:  amount,
		}

		encoded, err = types.EncodePolymorphicMessage(deposit, types.DepositEventTypeID)
		if err != nil {
			return nil, err
		}
		return encoded, nil
	case TransferEvent:
		if len(fields) != 5 {
			errMsg := "invalid transfer event format"
			log.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}

		amount, err := strconv.Atoi(fields[4])
		if err != nil {
			return nil, err
		}

		transfer := types.Transfer{
			FromAccount: fields[1],
			ToAccount:   fields[3],
			Amount:      amount,
		}

		encoded, err = types.EncodePolymorphicMessage(transfer, types.TransferID)

		if err != nil {
			return nil, err
		}

		return encoded, nil
	default:
		errMsg := "unrecognized event type"
		log.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
}
