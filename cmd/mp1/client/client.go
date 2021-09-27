package client

import (
	"bufio"
	"io"
	"net"
	"os"
	"time"

	"github.com/bamboovir/cs425/lib/mp1/config"
	"github.com/bamboovir/cs425/lib/mp1/types"
	"github.com/bamboovir/cs425/lib/retry"
	log "github.com/sirupsen/logrus"
)

var (
	logger = log.WithField("src", "client")
)

func RunClients(nodeID string, configPath string) (err error) {
	nodesConfig, err := config.ConfigParser(configPath)
	if err != nil {
		logger.Errorf("config parse err: %v", err)
		return err
	}

	rootEmitter := ReaderPipeline(os.Stdin)

	subEmitters := make([]chan []byte, 0)

	for _, c := range nodesConfig.ConfigItems {
		addr := net.JoinHostPort(c.NodeHost, c.NodePort)
		subEmitter := make(chan []byte, 1000)
		subEmitters = append(subEmitters, subEmitter)
		go RunClient(nodeID, c.NodeID, addr, subEmitter)
	}

	for msg := range rootEmitter {
		for _, subEmitter := range subEmitters {
			subEmitter <- msg
		}
	}

	return nil
}

func ReaderPipeline(reader io.Reader) <-chan []byte {
	out := make(chan []byte)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			eventMsg, err := types.EncodeTransactionsMsg(line)
			if err != nil {
				logger.Errorf("encode input msg failed with err :%v, skip", err)
				continue
			}
			out <- eventMsg
		}

		err := scanner.Err()

		if err != nil {
			logger.Errorf("read err: %v", err)
		} else {
			logger.Info("reach EOF")
		}
		close(out)
	}()

	return out
}

func RunClient(srcNodeID string, dstNodeID string, addr string, in <-chan []byte) (err error) {
	var client net.Conn
	err = retry.Retry(0, time.Second*5, func() error {
		logger.Infof("node [%s] tries to connect to the server [%s] in [%s]", srcNodeID, dstNodeID, addr)
		client, err = net.DialTimeout("tcp", addr, time.Second*10)
		if err != nil {
			logger.Errorf("node [%s] failed to connect to the server [%s] in [%s], retry", srcNodeID, dstNodeID, addr)
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	logger.Infof("node [%s] success connect to the server [%s] in [%s]", srcNodeID, dstNodeID, addr)
	defer client.Close()

	hi, _ := types.NewHi(srcNodeID).Encode()
	hi = append(hi, '\n')
	_, err = client.Write(hi)
	if err != nil {
		logger.Errorf("client write error: %v", err)
	}

	for msg := range in {
		msg := append(msg, '\n')
		_, err := client.Write(msg)
		if err != nil {
			logger.Errorf("client write error: %v", err)
		}
	}

	return nil
}
