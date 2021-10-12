package multicast

import (
	"container/heap"
	"fmt"
	"strings"
)

type TOHoldQueueItem struct {
	body           []byte
	proposalSeqNum uint64
	processID      string
	agreed         bool
	msgID          string
	index          int
}

type TOHoldPriorityQueue []*TOHoldQueueItem

func (pq TOHoldPriorityQueue) Snapshot() string {
	builder := strings.Builder{}
	for i, item := range pq {
		if i == 10 {
			break
		}
		record := fmt.Sprintf("[%d:%s] %t %s %d\n", item.proposalSeqNum, item.processID, item.agreed, item.msgID, item.index)
		builder.WriteString(record)
	}
	return builder.String()
}

func (pq TOHoldPriorityQueue) Len() int { return len(pq) }

func (pq TOHoldPriorityQueue) Less(i, j int) bool {
	if pq[i].proposalSeqNum < pq[j].proposalSeqNum {
		return true
	} else if pq[i].proposalSeqNum > pq[j].proposalSeqNum {
		return false
	}

	if pq[i].agreed == false && pq[j].agreed == true {
		return true
	} else if pq[i].agreed == true && pq[j].agreed == false {
		return false
	}

	return pq[i].processID < pq[j].processID
}

func (pq TOHoldPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *TOHoldPriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*TOHoldQueueItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *TOHoldPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func (pq *TOHoldPriorityQueue) Peek() interface{} {
	old := *pq
	return old[0]
}

func (pq *TOHoldPriorityQueue) Update(item *TOHoldQueueItem, processID string, seqNum uint64) {
	item.agreed = true
	item.processID = processID
	item.proposalSeqNum = seqNum
	heap.Fix(pq, item.index)
}
