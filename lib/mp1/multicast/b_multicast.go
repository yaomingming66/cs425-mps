package multicast

import (
	"github.com/bamboovir/cs425/lib/mp1/dispatcher"
)

type BDispatcher struct {
	dispatcher *dispatcher.Dispatcher
}

func NewBDispatcher(in chan interface{}) *BDispatcher {
	msgChan := make(chan dispatcher.Msg)
	go func() {
		for m := range in {
			m := m.([]byte)
			msg := BMsg{}
			_, err := msg.Decode(m)
			if err != nil {
				logger.Errorf("decode b-multicast dispatcher message err: %v", err)
				continue
			}
			msgChan <- dispatcher.Msg{
				Path: msg.Path,
				Body: m,
			}
		}
	}()

	return &BDispatcher{
		dispatcher: dispatcher.New(msgChan),
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
	b.dispatcher.Run()
}
