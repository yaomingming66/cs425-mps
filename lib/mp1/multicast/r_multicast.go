package multicast

import (
	"context"

	"github.com/bamboovir/cs425/lib/mp1/router"
	sync "github.com/sasha-s/go-deadlock"

	errors "github.com/pkg/errors"
)

const (
	RMulticastPath = "/r-multicast"
)

type RMulticast struct {
	bmulticast   *BMulticast
	received     map[string]struct{}
	receivedLock *sync.Mutex
	router       *router.Router
}

func NewRMulticast(b *BMulticast) *RMulticast {

	return &RMulticast{
		bmulticast:   b,
		received:     map[string]struct{}{},
		receivedLock: &sync.Mutex{},
		router:       router.New(),
	}
}

func (r *RMulticast) AddMsgIfNotExist(msgID string) bool {
	r.receivedLock.Lock()
	defer r.receivedLock.Unlock()
	_, ok := r.received[msgID]
	if !ok {
		r.received[msgID] = struct{}{}
		return true
	}
	return false
}

func (r *RMulticast) Multicast(path string, v interface{}) (err error) {
	rmsg, err := NewRMsg(path, v)

	if err != nil {
		return errors.Wrap(err, "r-multicast failed")
	}

	err = r.bmulticast.Multicast(RMulticastPath, rmsg)
	if err != nil {
		return errors.Wrap(err, "r-multicast failed")
	}

	return nil
}

func RMsgDecodeWrapper(f func(*RMsg) error) func(interface{}) error {
	return func(v interface{}) error {
		msg := v.(*RMsg)
		return f(msg)
	}
}

func (r *RMulticast) Bind(path string, f func(msg *RMsg) error) {
	r.router.Bind(path, RMsgDecodeWrapper(f))
}

func (r *RMulticast) bindRDeliver() {
	r.bmulticast.Bind(RMulticastPath, func(msg *BMsg) error {
		rmsg := &RMsg{}
		_, err := rmsg.Decode(msg.Body)
		if err != nil {
			return errors.Wrap(err, "r-deliver failed")
		}
		ok := r.AddMsgIfNotExist(rmsg.ID)
		if ok {
			if msg.SrcID != r.bmulticast.group.SelfNodeID {
				rmsgCopy := &RMsg{
					ID:   rmsg.ID,
					Path: rmsg.Path,
					Body: rmsg.Body,
				}
				err = r.bmulticast.Multicast(RMulticastPath, rmsgCopy)

				if err != nil {
					return errors.Wrap(err, "r-deliver failed")
				}
			}
			return r.router.Run(rmsg.Path, rmsg)
		}
		return nil
	},
	)
}

func (r *RMulticast) Start(ctx context.Context) (err error) {
	r.bindRDeliver()
	return r.bmulticast.Start(ctx)
}
