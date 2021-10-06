package multicast

import (
	"fmt"
	"net"
	"time"

	"github.com/bamboovir/cs425/lib/mp1/types"
	"github.com/bamboovir/cs425/lib/retry"
)

func (b *BMulticast) runClient(srcID string, dstID string, addr string, in <-chan []byte, retryInterval time.Duration) (err error) {
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

	b.startSyncWaitGroup.Done()

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
		return fmt.Errorf(errmsg)
	}

	for msg := range in {
		msgCopy := make([]byte, len(msg))
		copy(msgCopy, msg)
		msgCopy = append(msgCopy, '\n')
		_, err = client.Write(msgCopy)
		if err != nil {
			errmsg := fmt.Sprintf("client lost connection, write error: %v", err)
			logger.Error(errmsg)
			return fmt.Errorf(errmsg)
		}
	}

	return nil
}
