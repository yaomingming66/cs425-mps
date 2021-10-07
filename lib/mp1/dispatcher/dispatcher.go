package dispatcher

import (
	"encoding/json"
	"sync"

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
	publisher  chan Msg
	routers    map[string]func([]byte)
	routerLock *sync.Mutex
}

func New(in chan Msg) *Dispatcher {
	d := &Dispatcher{
		publisher:  in,
		routers:    make(map[string]func([]byte)),
		routerLock: &sync.Mutex{},
	}
	return d
}

func (d *Dispatcher) Bind(path string, f func(msg []byte)) {
	d.routerLock.Lock()
	defer d.routerLock.Unlock()
	d.routers[path] = f
}

func (d *Dispatcher) Run() {
	for msg := range d.publisher {
		d.routerLock.Lock()
		f, ok := d.routers[msg.Path]
		d.routerLock.Unlock()
		if !ok {
			logger.Errorf("path [%s] with body [%s] don't match any router", msg.Path, msg.Body)
			continue
		}
		f(msg.Body)
	}
}
