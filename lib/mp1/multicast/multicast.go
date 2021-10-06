package multicast

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"context"
)

var (
	logger = log.WithField("src", "multicast")
)

const (
	RMultiCastPath = "/r-multicast"
)

type Node struct {
	ID   string
	Addr string
}

type Group struct {
	members      []Node
	membersLock  *sync.Mutex
	SelfNodeID   string
	SelfNodeAddr string
	bmulticast   *BMulticast
	rmulticast   *RMulticast
	totalOrder   *TotalOrding
}

func (g *Group) B() *BMulticast {
	return g.bmulticast
}

func (g *Group) R() *RMulticast {
	return g.rmulticast
}

func (g *Group) TO() *TotalOrding {
	return g.totalOrder
}

func (g *Group) Start(ctx context.Context) (err error) {
	g.bmulticast = NewBMulticast(g)
	g.rmulticast = NewRMulticast(g.bmulticast)
	g.totalOrder = NewTotalOrder(g.bmulticast, g.rmulticast)

	return g.totalOrder.Start(ctx)
}
