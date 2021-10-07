package multicast

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bamboovir/cs425/lib/broker"
	errors "github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type BMulticast struct {
	group              *Group
	memberUpdate       *broker.Broker
	emitters           map[string]chan []byte
	emittersLock       *sync.Mutex
	router             *BDispatcher
	deliver            chan []byte
	startSyncWaitGroup *sync.WaitGroup
}

func NewBMulticast(group *Group) *BMulticast {
	deliver := make(chan []byte)

	return &BMulticast{
		memberUpdate:       broker.New(),
		group:              group,
		emitters:           map[string]chan []byte{},
		emittersLock:       &sync.Mutex{},
		router:             NewBDispatcher(deliver),
		deliver:            deliver,
		startSyncWaitGroup: &sync.WaitGroup{},
	}
}

func (b *BMulticast) MemberCount() int {
	b.emittersLock.Lock()
	defer b.emittersLock.Unlock()
	return len(b.emitters)
}

func (b *BMulticast) Unicast(dstID string, path string, msg []byte) (err error) {
	b.emittersLock.Lock()
	defer b.emittersLock.Unlock()

	emitter, ok := b.emitters[dstID]
	if !ok {
		errmsg := fmt.Sprintf("dst node id [%s] emitter not exists, unicast failed", dstID)
		return fmt.Errorf(errmsg)
	}
	bmsg := BMsg{
		SrcID: b.group.SelfNodeID,
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

func (b *BMulticast) Multicast(path string, msg []byte) (err error) {
	b.emittersLock.Lock()
	defer b.emittersLock.Unlock()
	bmsg := BMsg{
		SrcID: b.group.SelfNodeID,
		Path:  path,
		Body:  msg,
	}
	bmsgBytes, err := bmsg.Encode()
	if err != nil {
		return errors.Wrap(err, "b-multicast failed")
	}
	for _, emitter := range b.emitters {
		emitter <- bmsgBytes
	}
	return nil
}

func (b *BMulticast) Router() *BDispatcher {
	return b.router
}

func (b *BMulticast) startServer() (err error) {
	return runServer(
		b.group.SelfNodeID,
		b.group.SelfNodeAddr,
		b.deliver,
	)
}

// TODO: Config Emitter Buffer Size
// TODO: Config Client reconnect interval
func (b *BMulticast) startClients() {
	for _, m := range b.group.members {
		b.startSyncWaitGroup.Add(1)
		emitter := make(chan []byte)
		b.emitters[m.ID] = emitter
		go b.startClient(m.ID, m.Addr, emitter, 5*time.Second)
	}
}

func (b *BMulticast) startClient(
	dstNodeID string,
	addr string,
	emitter chan []byte,
	retryInterval time.Duration,
) (err error) {
	err = b.runClient(b.group.SelfNodeID, dstNodeID, addr, emitter, retryInterval)
	logger.Infof("eject node [%s] from group", dstNodeID)
	b.emittersLock.Lock()
	defer b.emittersLock.Unlock()
	delete(b.emitters, dstNodeID)
	b.memberUpdate.Publish(len(b.emitters))
	if err != nil {
		return err
	}
	return nil
}

func (b *BMulticast) MembersUpdate() chan interface{} {
	return b.memberUpdate.Subscribe()
}

func (b *BMulticast) Start(ctx context.Context) (err error) {
	go b.memberUpdate.Start()

	errGroup, _ := errgroup.WithContext(ctx)
	errGroup.Go(
		func() error {
			return b.startServer()
		},
	)

	b.startClients()

	serverChan := make(chan error)
	clientChan := make(chan struct{})

	go func() {
		b.startSyncWaitGroup.Wait()
		clientChan <- struct{}{}
	}()

	go func() {
		err := errGroup.Wait()
		if err != nil {
			serverChan <- err
		}
	}()

	select {
	case err := <-serverChan:
		return err
	case <-clientChan:
		time.Sleep(5 * time.Second)
	}
	return nil
}
