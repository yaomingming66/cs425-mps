package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	sync "github.com/sasha-s/go-deadlock"
)

var counter int32
var mutex sync.Mutex

func Add(x int32) {
	mutex.Lock()
	defer mutex.Unlock()
	counter += x
	fmt.Printf("%d\n", counter)
}

type Msg struct {
	body string
	from int
}

type N struct {
	ID    int
	RLock *sync.Mutex
	R     map[string]struct{}
	OLock *sync.Mutex
	O     map[int]chan Msg
	ILock *sync.Mutex
	I     map[int]chan Msg
}

type G struct {
	NS map[int]*N
}

func (g *G) Run() {
	for _, n := range g.NS {
		n.BD()
	}
}

func (n *N) BM(msg string) {
	n.OLock.Lock()
	defer n.OLock.Unlock()
	fmt.Printf("%d -> %d\n", n.ID, n.ID)
	Add(1)
	n.O[n.ID] <- Msg{
		body: msg,
		from: n.ID,
	}
	for to, v := range n.O {
		if to == n.ID {
			continue
		}
		fmt.Printf("%d -> %d\n", n.ID, to)
		Add(1)
		v <- Msg{
			body: msg,
			from: n.ID,
		}
	}
}

func (n *N) BMA(msg string) {
	n.OLock.Lock()
	defer n.OLock.Unlock()
	for memberID, vc := range n.O {
		if memberID > n.ID {
			fmt.Printf("%d -> %d\n", n.ID, memberID)
			Add(1)
			vc <- Msg{
				body: msg,
				from: n.ID,
			}
		}
	}
}

func (n *N) BD() {
	n.ILock.Lock()
	defer n.ILock.Unlock()
	for _, v := range n.I {
		go func(in chan Msg) {
			for msg := range in {
				n.RLock.Lock()
				_, ok := n.R[msg.body]
				if !ok {
					n.R[msg.body] = struct{}{}
					if msg.from != n.ID {
						// n.BM(msg.body)
						n.BMA(msg.body)
					}
				}
				n.RLock.Unlock()
			}
		}(v)
	}
}

func NewG(n int) *G {
	ns := map[int]*N{}

	for i := 0; i < n; i++ {
		node := &N{
			ID:    i,
			RLock: &sync.Mutex{},
			R:     map[string]struct{}{},
			OLock: &sync.Mutex{},
			O:     map[int]chan Msg{},
			ILock: &sync.Mutex{},
			I:     map[int]chan Msg{},
		}
		ns[i] = node
	}

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			ch := make(chan Msg, 100000)
			ns[i].O[j] = ch
			ns[j].I[i] = ch
		}
	}

	return &G{
		NS: ns,
	}
}

func main() {
	nodeCount, _ := strconv.Atoi(os.Args[1])
	srcNodeNum, _ := strconv.Atoi(os.Args[2])
	g := NewG(nodeCount)
	g.Run()
	g.NS[srcNodeNum].BM("msg")
	time.Sleep(time.Duration(1<<63 - 1))
}
