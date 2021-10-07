package multicast

import "github.com/bamboovir/cs425/lib/mp1/dispatcher"

type TODispatcher struct {
	in         chan []byte
	msgChan    chan dispatcher.Msg
	dispatcher *dispatcher.Dispatcher
}

func NewTODispatcher(in chan []byte) *TODispatcher {
	msgChan := make(chan dispatcher.Msg)
	return &TODispatcher{
		dispatcher: dispatcher.New(msgChan),
		msgChan:    msgChan,
		in:         in,
	}
}

func toMsgWrapper(f func(msg TOMsg)) func(msg []byte) {
	return func(m []byte) {
		msg := TOMsg{}
		_, err := msg.Decode(m)
		if err != nil {
			logger.Errorf("decode to-multicast msg failed")
			return
		}
		f(msg)
	}
}

func (b *TODispatcher) Bind(path string, f func(msg TOMsg)) {
	b.dispatcher.Bind(path, toMsgWrapper(f))
}

func (b *TODispatcher) Run() {
	go func() {
		for m := range b.in {
			msg := TOMsg{}
			_, err := msg.Decode(m)
			if err != nil {
				logger.Errorf("decode to-multicast dispatcher message err: %v", err)
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
