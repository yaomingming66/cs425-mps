package multicast

type TOHoldQueueItem struct {
	proposalSeqNum uint64
	processID      string
	msgID          string
}

type TOHoldPriorityQueue []*TOHoldQueueItem

func (pq TOHoldPriorityQueue) Len() int { return len(pq) }

func (pq TOHoldPriorityQueue) Less(i, j int) bool {
	if pq[i].proposalSeqNum < pq[j].proposalSeqNum {
		return true
	} else if pq[i].proposalSeqNum > pq[j].proposalSeqNum {
		return false
	}

	return pq[i].processID < pq[j].processID
}

func (pq TOHoldPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *TOHoldPriorityQueue) Push(x interface{}) {
	item := x.(*TOHoldQueueItem)
	*pq = append(*pq, item)
}

func (pq *TOHoldPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[0 : n-1]
	return item
}
