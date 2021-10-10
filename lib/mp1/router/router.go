package router

import (
	"fmt"

	sync "github.com/sasha-s/go-deadlock"

	log "github.com/sirupsen/logrus"
)

var (
	logger = log.WithField("src", "router")
)

type Msg struct {
	Path string
	Body interface{}
}

func NewMsg(path string, v interface{}) *Msg {
	return &Msg{
		Path: path,
		Body: v,
	}
}

type Router struct {
	routers    map[string]func(interface{}) error
	routerLock *sync.Mutex
}

func New() *Router {
	d := &Router{
		routers:    map[string]func(interface{}) error{},
		routerLock: &sync.Mutex{},
	}
	return d
}

func (d *Router) Bind(path string, f func(msg interface{}) error) {
	d.routerLock.Lock()
	defer d.routerLock.Unlock()
	d.routers[path] = f
}

func (d *Router) Run(path string, msg interface{}) error {
	d.routerLock.Lock()
	f, ok := d.routers[path]
	d.routerLock.Unlock()
	if !ok {
		errmsg := fmt.Sprintf("path [%s] with body [%s] don't match any router", path, msg)
		logger.Errorf("%s", errmsg)
		return fmt.Errorf(errmsg)
	}
	return f(msg)
}
