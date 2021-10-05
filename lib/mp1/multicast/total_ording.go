package multicast

// import (
// 	"fmt"

// 	"github.com/bamboovir/cs425/lib/mp1/broker"
// )

// const (
// 	AskProposalSeqPath       = "/total-ording/ask-proposal-seq"
// 	WaitProposalSeqPath      = "/total-ording/wait-proposal-seq"
// 	AnnounceAgreementSeqPath = "/total-ording/announce-agreement-seq"
// )

// // 只需要一个 RDeliver instance
// // 进行路径匹配
// type TotalOrding struct {
// 	maxAgreementSeqNumOfGroup uint64
// 	maxProposalSeqNumOfSelf   uint64
// 	group                     *Group
// 	deliverBroker 		  *broker.Broker
// }

// func NewTotalOrder(group *Group) *TotalOrding {
// 	return &TotalOrding{
// 		maxAgreementSeqNumOfGroup: 0,
// 		maxProposalSeqNumOfSelf:   0,
// 		group:                     group,
// 	}
// }

// func (t *TotalOrding) TOMulticast(msg []byte) {
// 	msg = ""
// 	msg.path = AskProposalSeqPath
// 	msg.id = ""
// 	msg.body = ""
// 	t.group.RMulticast(msg)
// 	go t.WaitProposal(id)

// }

// func (t *TotalOrding) WaitProposalSeqs(id string) {
// 	rDeliverChannel := t.group.RDeliver()
// 	proposalSeqs := make([]uint64, 0)
// 	for msg := range rDeliverChannel {
// 		msg := msg.([]byte)
// 		src := ""
// 		// try wait all message
// 		msg.path = WaitProposalSeqPath
// 		msg.id == id
// 		if msg.id == id {
// 			proposalSeqs = append(proposalSeqs, msg.proposalSeq)
// 		}
// 		proposalSeqs
// 	}

// 	//
// 	agreementSeqNum := MaxOfArrayUint64(proposalSeqs)

// 	msg := ""
// 	msg.path = "/total-ording/agreement-seq"
// 	msg.id = ""
// 	t.group.RMulticast(msg)
// }

// func MaxUint64(x uint64, y uint64) uint64 {
// 	if x > y {
// 		return x
// 	}
// 	return y
// }

// func MaxOfArrayUint64(arr []uint64) (uint64, error) {
// 	if len(arr) < 1 {
// 		return 0, fmt.Errorf("arr is an empty sequence")
// 	}

// 	max := arr[0]
// 	for _, value := range arr {
// 		if max < value {
// 			max = value
// 		}
// 	}
// 	return max, nil
// }

// func (t *TotalOrding) RegisterTOMulticastProposal() {
// 	t.Bind(AskProposalSeqPath, func(msg []byte) {
// 		msg.src := ""
// 		msg.id := ""
// 		msg.body := ""
// 		proposalSeqNum := MaxUint64(t.maxAgreementSeqNumOfGroup, t.maxProposalSeqNumOfSelf) + 1
// 		t.maxProposalSeqNumOfSelf = proposalSeqNum
// 		t.group.Unicast(src, {
// 			msg.id,
// 			proposalSeqNum,
// 			WaitProposalSeqPath
// 			msg.selfProcessID
// 		})
// 	})

// 	t.Bind(AnnounceAgreementSeqPath, func(msg []byte) {
// 		msg.id := ""
// 		msg.agreementSeqNum := 0
// 		t.maxAgreementSeqNumOfGroup = MaxUint64(t.maxAgreementSeqNumOfGroup, msg.agreementSeqNum)

// 		if msg.agreementSeqNum != msgIDToProposalSeqNum[id] {
// 			push into heap
// 		}

// 		if heap[0] () has agreement {
// 			t.deliverBroker.Publish(...)
// 		}
// 	})

// 	rDeliverChannel := t.group.RDeliver()
// 	for msg := range rDeliverChannel {
// 		msg := msg.([]byte)
// 		src := ""
// 		proposalSeqNum := MaxUint64(t.maxAgreementSeqNumOfGroup, t.maxProposalSeqNumOfSelf) + 1
// 		t.maxProposalSeqNumOfSelf = proposalSeqNum
// 		/*
// 			src,

// 		*/

// 	}
// }
