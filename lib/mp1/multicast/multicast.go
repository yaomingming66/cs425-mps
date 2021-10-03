package multicast

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"context"

	"github.com/bamboovir/cs425/lib/mp1/broker"
	"golang.org/x/sync/errgroup"
)

var (
	logger = log.WithField("src", "multicast")
)

type Emitter struct {
	Channel chan Msg
	Quit    chan struct{}
}

type Node struct {
	ID   string
	Addr string
}

type Group struct {
	Members            []Node
	SelfNodeID         string
	SelfNodeAddr       string
	Emitters           map[string]Emitter
	EmittersLock       *sync.Mutex
	BDeliverBroker     *broker.Broker
	RDeliverBroker     *broker.Broker
	Received           map[string]struct{}
	ReceivedLock       *sync.Mutex
	StartSyncWaitGroup *sync.WaitGroup
}

func (g *Group) Start(ctx context.Context) (err error) {
	go g.BDeliverBroker.Start()
	go g.RDeliverBroker.Start()

	errGroup, _ := errgroup.WithContext(ctx)
	errGroup.Go(
		func() error {
			return g.RegisterBDeliver()
		},
	)
	g.RegisterBMulticast()

	go g.RegisterRDeliver()
	g.RegisterRMulticast()

	serverChan := make(chan error)
	clientChan := make(chan struct{})

	go func() {
		g.StartSyncWaitGroup.Wait()
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

func (g *Group) Unicast(dstID string, msg []byte) (err error) {
	g.EmittersLock.Lock()
	defer g.EmittersLock.Unlock()

	emitter, ok := g.Emitters[dstID]
	if !ok {
		errmsg := fmt.Sprintf("dst node id [%s] emitter not exists", dstID)
		return fmt.Errorf(errmsg)
	}

	emitter.Channel <- Msg{
		SrcID: g.SelfNodeID,
		Body:  msg,
	}
	return nil
}

func (g *Group) RegisterBMulticast() {
	for _, m := range g.Members {
		g.StartSyncWaitGroup.Add(1)

		emitter := Emitter{
			Channel: make(chan Msg, 1000),
			Quit:    make(chan struct{}, 1),
		}

		g.Emitters[m.ID] = emitter

		go g.RegisterEmitter(m.ID, m.Addr, emitter, 5*time.Second)
	}

}

func (g *Group) RegisterBDeliver() (err error) {
	return RunServer(g.SelfNodeID, g.SelfNodeAddr, g.BDeliverBroker)
}

func (g *Group) BMulticast(msg []byte) {
	g.EmittersLock.Lock()
	defer g.EmittersLock.Unlock()
	for _, emitter := range g.Emitters {
		emitter.Channel <- Msg{
			SrcID: g.SelfNodeID,
			Body:  msg,
		}
	}
}

func (g *Group) BDeliver() chan interface{} {
	return g.BDeliverBroker.Subscribe()
}

func (g *Group) RMulticast(msg []byte) {
	rmsg := NewRMsg(msg)
	rmsgBytes, err := EncodeRMsg(rmsg)
	if err != nil {
		logger.Errorf("encode r message failed: %v", err)
		return
	}
	g.BMulticast(rmsgBytes)
}

func (g *Group) RDeliver() chan interface{} {
	return g.RDeliverBroker.Subscribe()
}

func (g *Group) RegisterRMulticast() {
	// Multiplexing B multicast channel
}

func (g *Group) RegisterRDeliver() {
	bDeliverChannel := g.BDeliver()
	for msg := range bDeliverChannel {
		msg := msg.(Msg)
		rmsg, err := DecodeRMsg(msg.Body)
		if err != nil {
			logger.Errorf("decode r message failed: %v", err)
			continue
		}
		g.ReceivedLock.Lock()
		_, ok := g.Received[rmsg.ID]
		if !ok {
			g.Received[rmsg.ID] = struct{}{}
			if msg.SrcID != g.SelfNodeID {
				rmsgCopy := &RMsg{
					ID:   rmsg.ID,
					Body: rmsg.Body,
				}
				rmsgBytes, err := EncodeRMsg(rmsgCopy)
				if err != nil {
					logger.Errorf("encode r message failed: %v", err)
					return
				}
				g.BMulticast(rmsgBytes)
			}

			g.RDeliverBroker.Publish(rmsg.Body)
		}
		g.ReceivedLock.Unlock()
	}
}

func (g *Group) RegisterEmitter(dstNodeID string, addr string, emitter Emitter, retryInterval time.Duration) (err error) {
	err = g.RunClient(g.SelfNodeID, dstNodeID, addr, emitter.Channel, emitter.Quit, retryInterval)
	logger.Infof("eject node [%s] from group", dstNodeID)
	g.EmittersLock.Lock()
	defer g.EmittersLock.Unlock()
	delete(g.Emitters, dstNodeID)
	if err != nil {
		return err
	}
	return nil
}

// TODO
func (g *Group) TOMulticast() {
}
