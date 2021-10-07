package multicast

import "container/heap"

type TOHoldQueueItem struct {
	proposalSeqNum uint64
	processID      string
	agreed         bool
	msgID          string
	index          int
}

type TOHoldPriorityQueue []*TOHoldQueueItem

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

func (pq *TOHoldPriorityQueue) Update(item *TOHoldQueueItem, proposalSeqNum uint64, processID string, agreed bool) {
	item.proposalSeqNum = proposalSeqNum
	item.processID = processID
	item.agreed = agreed
	heap.Fix(pq, item.index)
}
