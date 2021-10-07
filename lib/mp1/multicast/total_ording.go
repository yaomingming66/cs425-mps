package multicast

import (
	"container/heap"
	"context"
	"sync"

	errors "github.com/pkg/errors"
)

const (
	AskProposalSeqPath       = "/total-ording/ask-proposal-seq"
	WaitProposalSeqPath      = "/total-ording/wait-proposal-seq"
	AnnounceAgreementSeqPath = "/total-ording/announce-agreement-seq"
)

type ProposalItem struct {
	ProposalSeqNum uint64
	ProcessID      string
}

type TotalOrding struct {
	bmulticast                      *BMulticast
	rmulticast                      *RMulticast
	holdQueue                       *TOHoldPriorityQueue
	holdQueueLocker                 *sync.Mutex
	router                          *TODispatcher
	maxAgreementSeqNumOfGroup       uint64
	maxAgreementSeqNumOfGroupLocker *sync.Mutex
	maxProposalSeqNumOfSelf         uint64
	maxProposalSeqNumOfSelfLocker   *sync.Mutex
	deliver                         chan []byte
	waitProposalCounter             map[string][]*ProposalItem
	waitProposalCounterLock         *sync.Mutex
	msgMsgBody                      map[string][]byte
	msgMsgBodyLock                  *sync.Mutex
	msgAgreementSeqNum              map[string]uint64
	msgAgreementSeqNumLock          *sync.Mutex
	msgProposalSeqNum               map[string]uint64
	msgProposalSeqNumLock           *sync.Mutex
}

func NewTotalOrder(b *BMulticast, r *RMulticast) *TotalOrding {
	deliver := make(chan []byte)
	holdQueue := &TOHoldPriorityQueue{}
	heap.Init(holdQueue)

	return &TotalOrding{
		bmulticast:                      b,
		rmulticast:                      r,
		deliver:                         deliver,
		holdQueue:                       holdQueue,
		holdQueueLocker:                 &sync.Mutex{},
		router:                          NewTODispatcher(deliver),
		maxAgreementSeqNumOfGroup:       0,
		maxAgreementSeqNumOfGroupLocker: &sync.Mutex{},
		maxProposalSeqNumOfSelf:         0,
		maxProposalSeqNumOfSelfLocker:   &sync.Mutex{},
		waitProposalCounter:             map[string][]*ProposalItem{},
		waitProposalCounterLock:         &sync.Mutex{},
		msgMsgBody:                      map[string][]byte{},
		msgMsgBodyLock:                  &sync.Mutex{},
		msgAgreementSeqNum:              map[string]uint64{},
		msgAgreementSeqNumLock:          &sync.Mutex{},
		msgProposalSeqNum:               map[string]uint64{},
		msgProposalSeqNumLock:           &sync.Mutex{},
	}
}

func (t *TotalOrding) Start(ctx context.Context) (err error) {
	t.bindTODeliver()
	return t.rmulticast.Start(ctx)
}

func (t *TotalOrding) Router() *TODispatcher {
	return t.router
}

func (t *TotalOrding) Multicast(path string, msg []byte) (err error) {
	tomsg := NewTOMsg(path, msg)
	tomsgBytes, err := tomsg.Encode()
	if err != nil {
		return errors.Wrap(err, "to-multicast failed")
	}
	askMsg := NewTOAskProposalSeqMsg(t.bmulticast.group.SelfNodeID, tomsgBytes)
	askMsgBytes, err := askMsg.Encode()
	if err != nil {
		return errors.Wrap(err, "to-multicast failed")
	}
	proposalSeqs := make([]*ProposalItem, 0)

	t.waitProposalCounterLock.Lock()
	defer t.waitProposalCounterLock.Unlock()
	t.waitProposalCounter[askMsg.MsgID] = proposalSeqs
	logger.Errorf("send ask %s %s", askMsg.SrcID, askMsg.MsgID)
	err = t.rmulticast.Multicast(AskProposalSeqPath, askMsgBytes)
	if err != nil {
		return errors.Wrap(err, "to-multicast failed")
	}
	return nil
}

func (t *TotalOrding) bindTODeliver() {
	rRouter := t.rmulticast.Router()
	bRouter := t.bmulticast.Router()

	rRouter.Bind(AskProposalSeqPath, func(msg RMsg) {
		askMsg := &TOAskProposalSeqMsg{}
		_, err := askMsg.Decode(msg.Body)
		if err != nil {
			logger.Errorf("%v", err)
			return
		}

		t.maxAgreementSeqNumOfGroupLocker.Lock()
		defer t.maxAgreementSeqNumOfGroupLocker.Unlock()
		proposalSeqNum := MaxUint64(t.maxAgreementSeqNumOfGroup, t.maxProposalSeqNumOfSelf) + 1
		t.maxProposalSeqNumOfSelf = proposalSeqNum

		t.msgProposalSeqNumLock.Lock()
		defer t.msgProposalSeqNumLock.Unlock()
		t.msgProposalSeqNum[askMsg.MsgID] = proposalSeqNum

		t.msgMsgBodyLock.Lock()
		defer t.msgMsgBodyLock.Unlock()
		t.msgMsgBody[askMsg.MsgID] = askMsg.Body

		t.holdQueueLocker.Lock()
		defer t.holdQueueLocker.Unlock()

		heap.Push(t.holdQueue, &TOHoldQueueItem{
			proposalSeqNum: proposalSeqNum,
			msgID:          askMsg.MsgID,
			processID:      askMsg.SrcID,
		})

		replyProposalMsg := NewTOReplyProposalSeqMsg(t.bmulticast.group.SelfNodeID, askMsg.MsgID, proposalSeqNum)
		replyProposalMsgBytes, err := replyProposalMsg.Encode()
		if err != nil {
			logger.Errorf("%v", err)
			return
		}
		err = t.bmulticast.Unicast(askMsg.SrcID, WaitProposalSeqPath, replyProposalMsgBytes)
		if err != nil {
			logger.Errorf("%v", err)
			return
		}
	})

	bRouter.Bind(WaitProposalSeqPath, func(msg BMsg) {
		replyProposalMsg := &TOReplyProposalSeqMsg{}
		_, err := replyProposalMsg.Decode(msg.Body)
		if err != nil {
			logger.Errorf("%v", err)
			return
		}

		t.waitProposalCounterLock.Lock()
		defer t.waitProposalCounterLock.Unlock()

		_, ok := t.waitProposalCounter[replyProposalMsg.MsgID]

		if !ok {
			logger.Errorf("to msg id not exist")
			return
		}

		t.waitProposalCounter[replyProposalMsg.MsgID] = append(
			t.waitProposalCounter[replyProposalMsg.MsgID],
			&ProposalItem{
				ProposalSeqNum: replyProposalMsg.ProposalSeq,
				ProcessID:      replyProposalMsg.ProcessID,
			})

		memberCount := t.bmulticast.MemberCount()
		if len(t.waitProposalCounter[replyProposalMsg.MsgID]) < memberCount {
			return
		}

		agreementSeqItem, err := MaxOfArrayProposalItem(t.waitProposalCounter[replyProposalMsg.MsgID])
		if err != nil {
			logger.Errorf("%v", err)
			return
		}

		announceAgreementMsg := NewTOAnnounceAgreementSeqMsg(agreementSeqItem.ProcessID, agreementSeqItem.ProposalSeqNum, replyProposalMsg.MsgID)
		announceAgreementMsgBytes, err := announceAgreementMsg.Encode()
		if err != nil {
			logger.Errorf("%v", err)
			return
		}
		logger.Errorf("announce %s %d", announceAgreementMsg.MsgID, announceAgreementMsg.AgreementSeq)
		err = t.rmulticast.Multicast(AnnounceAgreementSeqPath, announceAgreementMsgBytes)
		if err != nil {
			logger.Errorf("%v", err)
			return
		}
	})

	rRouter.Bind(AnnounceAgreementSeqPath, func(msg RMsg) {
		announceAgreementMsg := &TOAnnounceAgreementSeqMsg{}
		_, err := announceAgreementMsg.Decode(msg.Body)
		if err != nil {
			logger.Errorf("%v", err)
			return
		}
		logger.Errorf("get announce %s -> %d", announceAgreementMsg.MsgID, announceAgreementMsg.AgreementSeq)
		t.maxAgreementSeqNumOfGroupLocker.Lock()
		defer t.maxAgreementSeqNumOfGroupLocker.Unlock()
		t.maxAgreementSeqNumOfGroup = MaxUint64(t.maxAgreementSeqNumOfGroup, announceAgreementMsg.AgreementSeq)

		t.msgAgreementSeqNumLock.Lock()
		defer t.msgAgreementSeqNumLock.Unlock()
		t.msgAgreementSeqNum[announceAgreementMsg.MsgID] = announceAgreementMsg.AgreementSeq

		t.holdQueueLocker.Lock()
		defer t.holdQueueLocker.Unlock()

		for _, v := range *t.holdQueue {
			if v.msgID == announceAgreementMsg.MsgID {
				v.proposalSeqNum = announceAgreementMsg.AgreementSeq
				v.processID = announceAgreementMsg.ProcessID
			}
		}

		t.msgMsgBodyLock.Lock()
		defer t.msgMsgBodyLock.Unlock()

		heap.Init(t.holdQueue)

		for t.holdQueue.Len() != 0 {
			item := heap.Pop(t.holdQueue).(*TOHoldQueueItem)
			heap.Push(t.holdQueue, item)
			_, hasAgreement := t.msgAgreementSeqNum[item.msgID]
			if !hasAgreement {
				break
			}
			msgBody, ok := t.msgMsgBody[item.msgID]
			if ok {
				t.deliver <- msgBody
				delete(t.msgMsgBody, item.msgID)
				delete(t.msgAgreementSeqNum, item.msgID)
				heap.Pop(t.holdQueue)
			}
		}
	})
}
