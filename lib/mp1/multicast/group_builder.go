package multicast

import (
	"sync"

	"github.com/bamboovir/cs425/lib/mp1/broker"
)

type GroupBuilder struct {
	SelfNodeID   string
	SelfNodeAddr string
	Memebers     []Node
}

func NewGroupBuilder() *GroupBuilder {
	return &GroupBuilder{
		Memebers: make([]Node, 0),
	}
}

func (g *GroupBuilder) WithSelfNodeID(selfNodeID string) *GroupBuilder {
	g.SelfNodeID = selfNodeID
	return g
}

func (g *GroupBuilder) WithSelfNodeAddr(selfNodeAddr string) *GroupBuilder {
	g.SelfNodeAddr = selfNodeAddr
	return g
}

func (g *GroupBuilder) AddMember(id string, addr string) *GroupBuilder {
	g.Memebers = append(g.Memebers, Node{ID: id, Addr: addr})
	return g
}

func (g *GroupBuilder) WithMembers(members []Node) *GroupBuilder {
	g.Memebers = members
	return g
}

func (g *GroupBuilder) Build() *Group {
	return &Group{
		SelfNodeID:         g.SelfNodeID,
		SelfNodeAddr:       g.SelfNodeAddr,
		Members:            g.Memebers,
		Emitters:           map[string]Emitter{},
		EmittersLock:       &sync.Mutex{},
		BDeliverBroker:     broker.New(),
		RDeliverBroker:     broker.New(),
		Received:           map[string]struct{}{},
		ReceivedLock:       &sync.Mutex{},
		StartSyncWaitGroup: &sync.WaitGroup{},
	}
}
