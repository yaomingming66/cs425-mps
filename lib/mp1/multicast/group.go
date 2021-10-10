package multicast

import (
	"context"

	sync "github.com/sasha-s/go-deadlock"
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
	return g.totalOrder.Start(ctx)
}
