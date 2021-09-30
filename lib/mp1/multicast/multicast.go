package multicast

import (
	"fmt"
	"net"
	"sync"
	"time"

	"bufio"
	"encoding/json"

	"github.com/bamboovir/cs425/lib/mp1/config"
	"github.com/bamboovir/cs425/lib/mp1/types"
	"github.com/bamboovir/cs425/lib/retry"
	log "github.com/sirupsen/logrus"

	"context"

	"github.com/bamboovir/cs425/lib/mp1/transaction"
	"golang.org/x/sync/errgroup"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_TYPE = "tcp"
)

var (
	logger = log.WithField("src", "multicast")
)

type Group struct {
	SelfNodeID   string
	SelfNodePort string
	Emitters     map[string]chan []byte
	EmitterLock  *sync.Mutex
	ErrGroup     *errgroup.Group
}

func NewGroup(selfNodeID string, selfNodePort string, configs config.Config) (group *Group, err error) {
	errGroup, _ := errgroup.WithContext(context.Background())
	errGroup.Go(
		func() error {
			return RunServer(selfNodeID, selfNodePort)
		},
	)

	emitters := map[string]chan []byte{}
	emittersLock := &sync.Mutex{}

	group = &Group{
		SelfNodeID:   selfNodeID,
		SelfNodePort: selfNodePort,
		Emitters:     emitters,
		EmitterLock:  emittersLock,
		ErrGroup:     errGroup,
	}

	for _, c := range configs.ConfigItems {
		addr := net.JoinHostPort(c.NodeHost, c.NodePort)
		emitter := make(chan []byte, 1000)
		emitters[c.NodeID] = emitter
		go group.RunClient(c.NodeID, addr, emitter)
	}

	return group, nil
}

func (g *Group) Wait() (err error) {
	return g.ErrGroup.Wait()
}

func (g *Group) Unicast(dstNodeID string, msg []byte) (err error) {
	emitter, ok := g.Emitters[dstNodeID]
	if !ok {
		errmsg := fmt.Sprintf("dst node id [%s] emitter not exists", err)
		return fmt.Errorf(errmsg)
	}
	emitter <- msg
	return nil
}

func (g *Group) BMulticast(msg []byte) {
	g.EmitterLock.Lock()
	defer g.EmitterLock.Unlock()
	for _, emitter := range g.Emitters {
		emitter <- msg
	}
}

func (g *Group) RMulticast(msg []byte) {

}

func (g *Group) RunClient(dstNodeID string, addr string, in chan []byte) (err error) {
	var client net.Conn
	err = retry.Retry(0, time.Second*5, func() error {
		logger.Infof("node [%s] tries to connect to the server [%s] in [%s]", g.SelfNodeID, dstNodeID, addr)
		client, err = net.DialTimeout("tcp", addr, time.Second*10)
		if err != nil {
			logger.Errorf("node [%s] failed to connect to the server [%s] in [%s], retry", g.SelfNodeID, dstNodeID, addr)
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	logger.Infof("node [%s] success connect to the server [%s] in [%s]", g.SelfNodeID, dstNodeID, addr)
	defer client.Close()
	defer func() {
		g.EmitterLock.Lock()
		defer g.EmitterLock.Unlock()
		delete(g.Emitters, dstNodeID)
		close(in)
	}()

	hi, _ := types.NewHi(g.SelfNodeID).Encode()
	hi = append(hi, '\n')
	_, err = client.Write(hi)
	if err != nil {
		logger.Errorf("client write error: %v", err)
	}

	for msg := range in {
		msg := append(msg, '\n')
		_, err := client.Write(msg)
		if err != nil {
			errmsg := fmt.Sprintf("client write error: %v", err)
			logger.Error(errmsg)
			return fmt.Errorf(errmsg)
		}
	}

	return nil
}

var (
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
