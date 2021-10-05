package multicast

import "github.com/bamboovir/cs425/lib/mp1/dispatcher"

type RDispatcher struct {
	dispatcher *dispatcher.Dispatcher
}

func NewRDispatcher(in chan interface{}) *RDispatcher {
	msgChan := make(chan dispatcher.Msg)
	go func() {
		for m := range in {
			m := m.([]byte)
			msg := RMsg{}
			_, err := msg.Decode(m)
			if err != nil {
				logger.Errorf("decode r-multicast dispatcher message err: %v", err)
				continue
			}
			msgChan <- dispatcher.Msg{
				Path: msg.Path,
				Body: m,
			}
		}
	}()

	return &RDispatcher{
		dispatcher: dispatcher.New(msgChan),
	}
}

func rMsgWrapper(f func(msg RMsg)) func(msg []byte) {
	return func(m []byte) {
		msg := RMsg{}
		_, err := msg.Decode(m)
		if err != nil {
			logger.Errorf("decode r-multicast msg failed")
			return
		}
		f(msg)
	}
}

func (r *RDispatcher) Bind(path string, f func(msg RMsg)) {
	r.dispatcher.Bind(path, rMsgWrapper(f))
}

func (r *RDispatcher) Run() {
	r.dispatcher.Run()
}
