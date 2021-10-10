package multicast

import (
	"fmt"
	"net"
	"time"

	"github.com/bamboovir/cs425/lib/mp1/types"
	"github.com/bamboovir/cs425/lib/retry"
)

type TCPClient struct {
	srcID         string
	dstID         string
	addr          string
	retryInterval time.Duration
	connection    net.Conn
}

func NewTCPClient(srcID string, dstID string, addr string, retryInterval time.Duration) (c *TCPClient, err error) {
	var connection net.Conn
	err = retry.Retry(0, retryInterval, func() error {
		logger.Infof("node [%s] tries to connect to the server [%s] in [%s]", srcID, dstID, addr)
		connection, err = net.DialTimeout("tcp", addr, time.Second*10)
		if err != nil {
			logger.Errorf("node [%s] failed to connect to the server [%s] in [%s], retry", srcID, dstID, addr)
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	logger.Infof("node [%s] success connect to the server [%s] in [%s]", srcID, dstID, addr)
	tcpConn := connection.(*net.TCPConn)

	tcpConn.SetReadBuffer(5 * MB)
	tcpConn.SetWriteBuffer(5 * MB)

	hi, _ := types.NewHi(srcID).Encode()
	hi = append(hi, '\n')
	_, err = connection.Write(hi)
	if err != nil {
		errmsg := fmt.Sprintf("client lost connection, write handshake message error: %v", err)
		logger.Error(errmsg)
		return nil, fmt.Errorf(errmsg)
	}

	return &TCPClient{
		srcID:         srcID,
		dstID:         dstID,
		addr:          addr,
		retryInterval: retryInterval,
		connection:    connection,
	}, nil
}

func (c *TCPClient) Send(msg []byte) (err error) {
	msgCopy := make([]byte, len(msg))
	copy(msgCopy, msg)
	msgCopy = append(msgCopy, '\n')
	// c.connection.SetWriteDeadline(time.Now().Add(time.Second * 100))
	_, err = c.connection.Write(msgCopy)
	if err != nil {
		// if os.IsTimeout(err) {
		// 	logger.Errorf("client send timeout: %v", err)
		// 	return nil
		// }
		return err
	}
	return nil
}

func (c *TCPClient) Close() error {
	return c.connection.Close()
}
