package multicast

import (
	"context"
	"sync"

	errors "github.com/pkg/errors"
)

type RMulticast struct {
	bmulticast   *BMulticast
	received     map[string]struct{}
	receivedLock *sync.Mutex
	router       *RDispatcher
	deliver      chan []byte
}

func NewRMulticast(b *BMulticast) *RMulticast {
	deliver := make(chan []byte)

	return &RMulticast{
		bmulticast:   b,
		received:     map[string]struct{}{},
		receivedLock: &sync.Mutex{},
		router:       NewRDispatcher(deliver),
		deliver:      deliver,
	}
}

func (r *RMulticast) Multicast(path string, msg []byte) (err error) {
	rmsg := NewRMsg(path, msg)
	rmsgBytes, err := rmsg.Encode()
	if err != nil {
		return errors.Wrap(err, "r-multicast failed")
	}
	err = r.bmulticast.Multicast(RMultiCastPath, rmsgBytes)
	if err != nil {
		return errors.Wrap(err, "r-multicast failed")
	}
	return nil
}

func (r *RMulticast) bindRDeliver() {
	r.bmulticast.router.Bind(RMultiCastPath, func(msg BMsg) {
		rmsg := &RMsg{}
		_, err := rmsg.Decode(msg.Body)
		if err != nil {
			logger.Errorf("decode r message failed: %v", err)
			return
		}
		r.receivedLock.Lock()
		defer r.receivedLock.Unlock()
		_, ok := r.received[rmsg.ID]
		if !ok {
			r.received[rmsg.ID] = struct{}{}
			if msg.SrcID != r.bmulticast.group.SelfNodeID {
				rmsgCopy := &RMsg{
					ID:   rmsg.ID,
					Path: rmsg.Path,
					Body: rmsg.Body,
				}
				rmsgBytes, err := rmsgCopy.Encode()
				if err != nil {
					logger.Errorf("encode r message failed: %v", err)
					return
				}
				err = r.bmulticast.Multicast(RMultiCastPath, rmsgBytes)

				if err != nil {
					logger.Errorf("b-multicast message failed: %v", err)
					return
				}
			}
			r.deliver <- msg.Body
		}
	})
}

func (r *RMulticast) Router() *RDispatcher {
	return r.router
}

func (r *RMulticast) Start(ctx context.Context) (err error) {
	r.bindRDeliver()
	return r.bmulticast.Start(ctx)
}
