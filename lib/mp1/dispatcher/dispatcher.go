package dispatcher

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

var (
	logger = log.WithField("src", "dispatcher")
)

type Msg struct {
	Path string `json:"path"`
	Body []byte `json:"body"`
}

func NewMsg(path string, v interface{}) (msg *Msg, err error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return &Msg{
		Path: path,
		Body: data,
	}, nil
}

func (m *Msg) Encode() (data []byte, err error) {
	data, err = json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *Msg) Decode(data []byte) (msg *Msg, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return m, err
	}
	return m, nil
}

type Dispatcher struct {
	publisher chan Msg
	routers   map[string]func([]byte)
}

func New(in chan Msg) *Dispatcher {
	d := &Dispatcher{
		publisher: in,
		routers:   make(map[string]func([]byte)),
	}
	return d
}

func (d *Dispatcher) Bind(path string, f func(msg []byte)) {
	d.routers[path] = f
}

func (d *Dispatcher) Run() {
	for {
		msg := <-d.publisher
		f, ok := d.routers[msg.Path]
		if !ok {
			logger.Errorf("path [%s] don't match any router", msg.Path)
			continue
		}
		f(msg.Body)
	}
}
