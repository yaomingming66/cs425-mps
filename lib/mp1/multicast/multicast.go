package multicast

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bamboovir/cs425/lib/mp1/config"
	log "github.com/sirupsen/logrus"

	"context"

	"golang.org/x/sync/errgroup"
)

var (
	logger = log.WithField("src", "multicast")
)

type Emitter struct {
	Channel chan Msg
	Quit    chan struct{}
}

type Msg struct {
	SrcID string `json:"src"`
	Body  []byte `json:"body"`
}

func EncodeMsg(msg *Msg) (data []byte, err error) {
	data, err = json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func DecodeMsg(data []byte) (msg *Msg, err error) {
	msg = &Msg{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

type Group struct {
	SelfNodeID      string
	SelfNodePort    string
	Emitters        map[string]Emitter
	EmittersLock    *sync.Mutex
	ErrGroup        *errgroup.Group
	BDeliverChannel chan Msg
	RDeliverChannel chan Msg
	Received        map[string]struct{}
	ReceivedLock    *sync.Mutex
}

func NewGroup(ctx context.Context, selfNodeID string, selfNodePort string, configs config.Config) (group *Group, err error) {
	bDeliverChannel := make(chan Msg, 1000)
	rDeliverChannel := make(chan Msg, 1000)

	errGroup, _ := errgroup.WithContext(ctx)
	errGroup.Go(
		func() error {
			return RunServer(selfNodeID, selfNodePort, bDeliverChannel)
		},
	)

	emitters := map[string]Emitter{}
	emittersLock := &sync.Mutex{}

	group = &Group{
		SelfNodeID:      selfNodeID,
		SelfNodePort:    selfNodePort,
		Emitters:        emitters,
		EmittersLock:    emittersLock,
		ErrGroup:        errGroup,
		BDeliverChannel: bDeliverChannel,
		RDeliverChannel: rDeliverChannel,
		Received:        map[string]struct{}{},
		ReceivedLock:    &sync.Mutex{},
	}

	for _, c := range configs.ConfigItems {
		addr := net.JoinHostPort(c.NodeHost, c.NodePort)

		emitter := Emitter{
			Channel: make(chan Msg, 1000),
			Quit:    make(chan struct{}),
		}

		emitters[c.NodeID] = emitter

		go group.RegisterEmitter(c.NodeID, addr, emitter, 5*time.Second)
	}

	return group, nil
}

func (g *Group) Wait() (err error) {
	return g.ErrGroup.Wait()
}

func (g *Group) Unicast(dstID string, msg []byte) (err error) {
	g.EmittersLock.Lock()
	defer g.EmittersLock.Unlock()

	emitter, ok := g.Emitters[dstID]
	if !ok {
		errmsg := fmt.Sprintf("dst node id [%s] emitter not exists", err)
		return fmt.Errorf(errmsg)
	}
	select {
	case <-emitter.Quit:
		return
	case emitter.Channel <- Msg{
		SrcID: g.SelfNodeID,
		Body:  msg,
	}:
	}
	return nil
}

func (g *Group) BMulticast(msg []byte) {
	g.EmittersLock.Lock()
	defer g.EmittersLock.Unlock()
	for _, emitter := range g.Emitters {
		select {
		case <-emitter.Quit:
			continue
		case emitter.Channel <- Msg{
			SrcID: g.SelfNodeID,
			Body:  msg,
		}:
		}
	}
}

func (g *Group) BDeliver() chan Msg {
	return g.BDeliverChannel
}

func (g *Group) RMulticast(msg []byte) {
	g.BMulticast(msg)
}

func (g *Group) RDeliver() chan Msg {
	return g.RDeliverChannel
}

func (g *Group) RegisterRDeliver() {
	go func() {
		for msg := range g.BDeliverChannel {
			g.ReceivedLock.Lock()
			_, ok := g.Received[string(msg.Body)]
			if !ok {
				g.Received[string(msg.Body)] = struct{}{}
				if msg.SrcID != g.SelfNodeID {
					g.BMulticast(msg.Body)
				}
				g.RDeliverChannel <- msg
			}
			g.ReceivedLock.Unlock()
		}
	}()
}

func (g *Group) RegisterEmitter(dstNodeID string, addr string, emitter Emitter, retryInterval time.Duration) (err error) {
	err = RunClient(g.SelfNodeID, dstNodeID, addr, emitter.Channel, emitter.Quit, retryInterval)
	logger.Infof("eject node [%s] from group", dstNodeID)
	g.EmittersLock.Lock()
	defer g.EmittersLock.Unlock()
	delete(g.Emitters, dstNodeID)
	if err != nil {
		return err
	}
	return nil
}
