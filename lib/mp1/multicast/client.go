package multicast

import (
	"fmt"
	"net"
	"time"

	"github.com/bamboovir/cs425/lib/mp1/types"
	"github.com/bamboovir/cs425/lib/retry"
)

func RunClient(srcID string, dstID string, addr string, in <-chan Msg, quit chan struct{}, retryInterval time.Duration) (err error) {
	var client net.Conn
	err = retry.Retry(0, retryInterval, func() error {
		logger.Infof("node [%s] tries to connect to the server [%s] in [%s]", srcID, dstID, addr)
		client, err = net.DialTimeout("tcp", addr, time.Second*10)
		if err != nil {
			logger.Errorf("node [%s] failed to connect to the server [%s] in [%s], retry", srcID, dstID, addr)
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	logger.Infof("node [%s] success connect to the server [%s] in [%s]", srcID, dstID, addr)
	defer client.Close()

	hi, _ := types.NewHi(srcID).Encode()
	hi = append(hi, '\n')
	_, err = client.Write(hi)
	if err != nil {
		errmsg := fmt.Sprintf("client lost connection, write handshake message error: %v", err)
		logger.Error(errmsg)
		quit <- struct{}{}
		return fmt.Errorf(errmsg)
	}

	for msg := range in {
		msgBytes, err := EncodeMsg(&msg)
		if err != nil {
			logger.Error("encode message failed")
		}
		msg := append(msgBytes, '\n')
		_, err = client.Write(msg)
		if err != nil {
			errmsg := fmt.Sprintf("client lost connection, write error: %v", err)
			logger.Error(errmsg)
			quit <- struct{}{}
			return fmt.Errorf(errmsg)
		}
	}

	quit <- struct{}{}
	return nil
}
