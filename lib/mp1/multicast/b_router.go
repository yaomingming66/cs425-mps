package multicast

import (
	"github.com/bamboovir/cs425/lib/mp1/dispatcher"
)

type BDispatcher struct {
	in         chan []byte
	msgChan    chan dispatcher.Msg
	dispatcher *dispatcher.Dispatcher
}

func NewBDispatcher(in chan []byte) *BDispatcher {
	msgChan := make(chan dispatcher.Msg)
	return &BDispatcher{
		dispatcher: dispatcher.New(msgChan),
		msgChan:    msgChan,
		in:         in,
	}
}

func bMsgWrapper(f func(msg BMsg)) func(msg []byte) {
	return func(m []byte) {
		msg := BMsg{}
		_, err := msg.Decode(m)
		if err != nil {
			logger.Errorf("decode b-multicast msg failed")
			return
		}
		f(msg)
	}
}

func (b *BDispatcher) Bind(path string, f func(msg BMsg)) {
	b.dispatcher.Bind(path, bMsgWrapper(f))
}

func (b *BDispatcher) Run() {
	go func() {
		for m := range b.in {
			msg := BMsg{}
			_, err := msg.Decode(m)
			if err != nil {
				logger.Errorf("decode b-multicast dispatcher message err: %v", err)
				continue
			}
			b.msgChan <- dispatcher.Msg{
				Path: msg.Path,
				Body: m,
			}
		}
	}()
	b.dispatcher.Run()
}
