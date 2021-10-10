package multicast

import (
	"context"
	"fmt"

	sync "github.com/sasha-s/go-deadlock"

	"time"

	"github.com/bamboovir/cs425/lib/broker"
	"github.com/bamboovir/cs425/lib/mp1/router"
	errors "github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type BMulticast struct {
	group              *Group
	memberUpdate       *broker.Broker
	senders            map[string]*TCPClient
	senderLock         *sync.Mutex
	router             *router.Router
	startSyncWaitGroup *sync.WaitGroup
}

func NewBMulticast(group *Group) *BMulticast {
	return &BMulticast{
		memberUpdate:       broker.New(),
		group:              group,
		senders:            map[string]*TCPClient{},
		senderLock:         &sync.Mutex{},
		router:             router.New(),
		startSyncWaitGroup: &sync.WaitGroup{},
	}
}

const (
	BMulticastPath = "/b-multicast"
)

func BMsgDecodeWrapper(f func(*BMsg) error) func(interface{}) error {
	return func(v interface{}) error {
		msg := v.(*BMsg)
		return f(msg)
	}
}

func (b *BMulticast) Bind(path string, f func(msg *BMsg) error) {
	b.router.Bind(path, BMsgDecodeWrapper(f))
}

func (b *BMulticast) AddMember(nodeID string, client *TCPClient) {
	b.senderLock.Lock()
	defer b.senderLock.Unlock()
	b.senders[nodeID] = client
}

func (b *BMulticast) MemberCount() int {
	b.senderLock.Lock()
	defer b.senderLock.Unlock()
	return len(b.senders)
}

func (b *BMulticast) IsNodeAlived(nodeID string) bool {
	b.senderLock.Lock()
	defer b.senderLock.Unlock()
	_, ok := b.senders[nodeID]
	return ok
}

func (b *BMulticast) Unicast(dstID string, path string, v interface{}) (err error) {
	b.senderLock.Lock()
	defer b.senderLock.Unlock()
	sender, ok := b.senders[dstID]

	if !ok {
		errmsg := fmt.Sprintf("dst node id [%s] sender not exists, unicast failed", dstID)
		return fmt.Errorf(errmsg)
	}

	bmsg, err := NewBMsg(b.group.SelfNodeID, path, v)

	if err != nil {
		return errors.Wrap(err, "b-unicast failed")
	}

	bmsgBytes, err := bmsg.Encode()

	if err != nil {
		return errors.Wrap(err, "b-unicast failed")
	}

	err = sender.Send(bmsgBytes)
	if err != nil {
		logger.Errorf("client lost connection, write error: %v", err)
		logger.Infof("eject node [%s] from group", dstID)
		sender.Close()
		delete(b.senders, dstID)
		b.memberUpdate.Publish(len(b.senders))
		return errors.Wrap(err, "b-unicast failed")
	}
	return nil
}

func (b *BMulticast) Multicast(path string, v interface{}) (err error) {
	b.senderLock.Lock()
	defer b.senderLock.Unlock()

	bmsg, err := NewBMsg(b.group.SelfNodeID, path, v)

	if err != nil {
		return errors.Wrap(err, "b-multicast failed")
	}

	bmsgBytes, err := bmsg.Encode()

	if err != nil {
		return errors.Wrap(err, "b-multicast failed")
	}

	for dstID, sender := range b.senders {
		err = sender.Send(bmsgBytes)
		if err != nil {
			logger.Errorf("client lost connection, write error: %v", err)
			logger.Infof("eject node [%s] from group", dstID)
			sender.Close()
			delete(b.senders, dstID)
			b.memberUpdate.Publish(len(b.senders))
		}
	}
	return nil
}

func (b *BMulticast) startServer() (err error) {
	for range b.group.members {
		b.startSyncWaitGroup.Add(1)
	}

	socket, err := startServer(
		b.group.SelfNodeID,
		b.group.SelfNodeAddr,
		b.router,
	)

	if err != nil {
		return err
	}

	go runServer(
		b.startSyncWaitGroup,
		b.group.SelfNodeID,
		socket,
		b.router,
	)

	return nil
}

func (b *BMulticast) startClients() (err error) {
	for _, m := range b.group.members {
		err = b.startClient(m.ID, m.Addr, 5*time.Second)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BMulticast) startClient(
	dstNodeID string,
	addr string,
	retryInterval time.Duration,
) (err error) {
	client, err := NewTCPClient(b.group.SelfNodeID, dstNodeID, addr, 5*time.Second)
	b.AddMember(dstNodeID, client)
	b.startSyncWaitGroup.Done()
	if err != nil {
		logger.Errorf("init tcp client failed: %v", err)
		return err
	}
	return nil
}

func (b *BMulticast) MembersUpdate() chan interface{} {
	return b.memberUpdate.Subscribe()
}

func (b *BMulticast) bindBDeliver() {
	b.router.Bind(BMulticastPath, func(v interface{}) error {
		msg := v.(*BMsg)
		return b.router.Run(msg.Path, msg)
	})
}

func (b *BMulticast) Start(ctx context.Context) (err error) {
	b.bindBDeliver()
	go b.memberUpdate.Start()

	errGroup, _ := errgroup.WithContext(ctx)
	errGroup.Go(
		func() error {
			return b.startServer()
		},
	)

	errGroup.Go(
		func() error {
			return b.startClients()
		},
	)

	err = errGroup.Wait()
	time.Sleep(time.Second * 5)
	return err
}
