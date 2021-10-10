package multicast

import (
	sync "github.com/sasha-s/go-deadlock"
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
	group := &Group{
		SelfNodeID:   g.SelfNodeID,
		SelfNodeAddr: g.SelfNodeAddr,
		members:      g.Memebers,
		membersLock:  &sync.Mutex{},
	}
	group.bmulticast = NewBMulticast(group)
	group.rmulticast = NewRMulticast(group.bmulticast)
	group.totalOrder = NewTotalOrder(group.bmulticast, group.rmulticast)
	return group
}
