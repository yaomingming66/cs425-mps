package multicast

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"context"

	"github.com/bamboovir/cs425/lib/mp1/broker"
	"github.com/bamboovir/cs425/lib/mp1/dispatcher"
	errors "github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var (
	logger = log.WithField("src", "multicast")
)

const (
	RMultiCastPath = "/r-multicast"
)

type Node struct {
	ID   string
	Addr string
}

type Group struct {
	Members            []Node
	SelfNodeID         string
	SelfNodeAddr       string
	emitters           map[string]chan []byte
	emittersLock       *sync.Mutex
	bDeliverBroker     *broker.Broker
	rDeliverBroker     *broker.Broker
	received           map[string]struct{}
	receivedLock       *sync.Mutex
	startSyncWaitGroup *sync.WaitGroup
	bDispatcher        *BDispatcher
	rDispatcher        *RDispatcher
	toDispatcher       *dispatcher.Dispatcher
}

func (g *Group) BDispatcher() *BDispatcher {
	return g.bDispatcher
}

func (g *Group) RDispatcher() *RDispatcher {
	return g.rDispatcher
}

func (g *Group) TODispatcher() *dispatcher.Dispatcher {
	return g.toDispatcher
}

func (g *Group) Start(ctx context.Context) (err error) {
	go g.bDeliverBroker.Start()
	go g.rDeliverBroker.Start()

	bDeliverChannel := g.bDeliverBroker.Subscribe()
	g.bDispatcher = NewBDispatcher(bDeliverChannel)
	go g.bDispatcher.Run()

	rDeliverChannel := g.rDeliverBroker.Subscribe()
	g.rDispatcher = NewRDispatcher(rDeliverChannel)
	go g.rDispatcher.Run()

	errGroup, _ := errgroup.WithContext(ctx)
	errGroup.Go(
		func() error {
			return g.registerBDeliver()
		},
	)
	g.registerBMulticast()
	go g.registerRDeliver()
	g.registerRMulticast()

	serverChan := make(chan error)
	clientChan := make(chan struct{})

	go func() {
		g.startSyncWaitGroup.Wait()
		clientChan <- struct{}{}
	}()

	go func() {
		err := errGroup.Wait()
		serverChan <- err
	}()

	select {
	case err := <-serverChan:
		return err
	case <-clientChan:
	}

	return nil
}

func (g *Group) Unicast(dstID string, path string, msg []byte) (err error) {
	g.emittersLock.Lock()
	defer g.emittersLock.Unlock()

	emitter, ok := g.emitters[dstID]
	if !ok {
		errmsg := fmt.Sprintf("dst node id [%s] emitter not exists, unicast failed", dstID)
		return fmt.Errorf(errmsg)
	}
	bmsg := BMsg{
		SrcID: g.SelfNodeID,
		Path:  path,
		Body:  msg,
	}
	bmsgBytes, err := bmsg.Encode()
	if err != nil {
		return errors.Wrap(err, "unicast failed")
	}
	emitter <- bmsgBytes
	return nil
}

// TODO: Config Emitter Buffer Size
// TODO: Config Client reconnect interval
func (g *Group) registerBMulticast() {
	for _, m := range g.Members {
		g.startSyncWaitGroup.Add(1)
		emitter := make(chan []byte, 1)
		g.emitters[m.ID] = emitter
		go g.registerEmitter(m.ID, m.Addr, emitter, 5*time.Second)
	}
}

func (g *Group) registerBDeliver() (err error) {
	err = runServer(g.SelfNodeID, g.SelfNodeAddr, g.bDeliverBroker)
	if err != nil {
		return errors.Wrap(err, "group register b-deliver failed")
	}
	return nil
}

func (g *Group) BMulticast(path string, msg []byte) (err error) {
	g.emittersLock.Lock()
	defer g.emittersLock.Unlock()
	bmsg := BMsg{
		SrcID: g.SelfNodeID,
		Path:  path,
		Body:  msg,
	}
	bmsgBytes, err := bmsg.Encode()
	if err != nil {
		return errors.Wrap(err, "b-multicast failed")
	}
	for _, emitter := range g.emitters {
		emitter <- bmsgBytes
	}
	return nil
}

func (g *Group) BDeliver() chan interface{} {
	return g.bDeliverBroker.Subscribe()
}

func (g *Group) RMulticast(path string, msg []byte) (err error) {
	rmsg := NewRMsg(path, msg)
	rmsgBytes, err := rmsg.Encode()
	if err != nil {
		return errors.Wrap(err, "r-multicast failed")
	}
	err = g.BMulticast(RMultiCastPath, rmsgBytes)
	if err != nil {
		return errors.Wrap(err, "r-multicast failed")
	}
	return nil
}

func (g *Group) RDeliver() chan interface{} {
	return g.rDeliverBroker.Subscribe()
}

func (g *Group) registerRMulticast() {
	/*
		Multiplexing B multicast channel,
		Nothing to do
	*/
}

func (g *Group) registerRDeliver() {
	g.bDispatcher.Bind(RMultiCastPath, func(msg BMsg) {
		rmsg := &RMsg{}
		_, err := rmsg.Decode(msg.Body)
		if err != nil {
			logger.Errorf("decode r message failed: %v", err)
			return
		}
		g.receivedLock.Lock()
		defer g.receivedLock.Unlock()
		_, ok := g.received[rmsg.ID]
		if !ok {
			g.received[rmsg.ID] = struct{}{}
			if msg.SrcID != g.SelfNodeID {
				rmsgCopy := &RMsg{
					ID:   rmsg.ID,
					Body: rmsg.Body,
				}
				rmsgBytes, err := rmsgCopy.Encode()
				if err != nil {
					logger.Errorf("encode r message failed: %v", err)
					return
				}
				err = g.BMulticast(RMultiCastPath, rmsgBytes)

				if err != nil {
					logger.Errorf("b-multicast message failed: %v", err)
					return
				}
			}

			g.rDeliverBroker.Publish(msg.Body)
		}
	})

}

func (g *Group) registerEmitter(dstNodeID string, addr string, emitter chan []byte, retryInterval time.Duration) (err error) {
	err = g.runClient(g.SelfNodeID, dstNodeID, addr, emitter, retryInterval)
	logger.Infof("eject node [%s] from group", dstNodeID)
	g.emittersLock.Lock()
	defer g.emittersLock.Unlock()
	delete(g.emitters, dstNodeID)
	if err != nil {
		return err
	}
	return nil
}

// TODO
func (g *Group) TOMulticast(path string, msg []byte) (err error) {
	return nil
}

func (g *Group) TODeliver() chan interface{} {
	return nil
}
