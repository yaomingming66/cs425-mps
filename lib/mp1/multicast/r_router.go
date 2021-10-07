package multicast

import "github.com/bamboovir/cs425/lib/mp1/dispatcher"

type RDispatcher struct {
	in         chan []byte
	msgChan    chan dispatcher.Msg
	dispatcher *dispatcher.Dispatcher
}

func NewRDispatcher(in chan []byte) *RDispatcher {
	msgChan := make(chan dispatcher.Msg)

	return &RDispatcher{
		dispatcher: dispatcher.New(msgChan),
		msgChan:    msgChan,
		in:         in,
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
	go func() {
		for m := range r.in {
			msg := RMsg{}
			_, err := msg.Decode(m)
			if err != nil {
				logger.Errorf("decode r-multicast dispatcher message err: %v", err)
				continue
			}
			r.msgChan <- dispatcher.Msg{
				Path: msg.Path,
				Body: m,
			}
		}
	}()
	r.dispatcher.Run()
}
