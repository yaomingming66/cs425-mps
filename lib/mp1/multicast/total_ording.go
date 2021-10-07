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

type MsgItem struct {
	Body            []byte
	ProposalSeqNum  uint64
	ProcessID       string
	AgreementSeqNum uint64
	Agreed          bool
}

type TotalOrding struct {
	bmulticast                      *BMulticast
	rmulticast                      *RMulticast
	holdQueueMap                    map[string]*TOHoldQueueItem
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
	msgItems                        map[string]*MsgItem
	msgItemsLock                    *sync.Mutex
}

func NewTotalOrder(b *BMulticast, r *RMulticast) *TotalOrding {
	deliver := make(chan []byte)
	holdQueue := &TOHoldPriorityQueue{}
	heap.Init(holdQueue)

	return &TotalOrding{
		bmulticast:                      b,
		rmulticast:                      r,
		deliver:                         deliver,
		holdQueueMap:                    map[string]*TOHoldQueueItem{},
		holdQueue:                       holdQueue,
		holdQueueLocker:                 &sync.Mutex{},
		router:                          NewTODispatcher(deliver),
		maxAgreementSeqNumOfGroup:       0,
		maxAgreementSeqNumOfGroupLocker: &sync.Mutex{},
		maxProposalSeqNumOfSelf:         0,
		maxProposalSeqNumOfSelfLocker:   &sync.Mutex{},
		waitProposalCounter:             map[string][]*ProposalItem{},
		waitProposalCounterLock:         &sync.Mutex{},
		msgItems:                        map[string]*MsgItem{},
		msgItemsLock:                    &sync.Mutex{},
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
	err = t.bmulticast.Multicast(AskProposalSeqPath, askMsgBytes)
	if err != nil {
		return errors.Wrap(err, "to-multicast failed")
	}
	return nil
}

func (t *TotalOrding) bindTODeliver() {
	// rRouter := t.rmulticast.Router()
	bRouter := t.bmulticast.Router()

	bRouter.Bind(AskProposalSeqPath, func(msg BMsg) {
		askMsg := &TOAskProposalSeqMsg{}
		_, err := askMsg.Decode(msg.Body)
		if err != nil {
			logger.Errorf("%v", err)
			return
		}

		t.maxAgreementSeqNumOfGroupLocker.Lock()
		defer t.maxAgreementSeqNumOfGroupLocker.Unlock()
		t.maxProposalSeqNumOfSelfLocker.Lock()
		defer t.maxProposalSeqNumOfSelfLocker.Unlock()

		proposalSeqNum := MaxUint64(t.maxAgreementSeqNumOfGroup, t.maxProposalSeqNumOfSelf) + 1
		t.maxProposalSeqNumOfSelf = proposalSeqNum

		t.msgItemsLock.Lock()
		defer t.msgItemsLock.Unlock()

		t.msgItems[askMsg.MsgID] = &MsgItem{
			ProposalSeqNum:  proposalSeqNum,
			Body:            askMsg.Body,
			ProcessID:       askMsg.SrcID,
			AgreementSeqNum: 0,
			Agreed:          false,
		}

		t.holdQueueLocker.Lock()
		defer t.holdQueueLocker.Unlock()

		item := &TOHoldQueueItem{
			proposalSeqNum: proposalSeqNum,
			msgID:          askMsg.MsgID,
			processID:      askMsg.SrcID,
			agreed:         false,
		}
		t.holdQueueMap[askMsg.MsgID] = item
		heap.Push(t.holdQueue, item)

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
		// logger.Errorf("announce %s %d", announceAgreementMsg.MsgID, announceAgreementMsg.AgreementSeq)
		err = t.bmulticast.Multicast(AnnounceAgreementSeqPath, announceAgreementMsgBytes)
		if err != nil {
			logger.Errorf("%v", err)
			return
		}
	})

	bRouter.Bind(AnnounceAgreementSeqPath, func(msg BMsg) {
		announceAgreementMsg := &TOAnnounceAgreementSeqMsg{}
		_, err := announceAgreementMsg.Decode(msg.Body)
		if err != nil {
			logger.Errorf("%v", err)
			return
		}

		// logger.Errorf("get announce %s -> %d", announceAgreementMsg.MsgID, announceAgreementMsg.AgreementSeq)

		t.maxAgreementSeqNumOfGroupLocker.Lock()
		defer t.maxAgreementSeqNumOfGroupLocker.Unlock()
		t.msgItemsLock.Lock()
		defer t.msgItemsLock.Unlock()

		t.maxAgreementSeqNumOfGroup = MaxUint64(t.maxAgreementSeqNumOfGroup, announceAgreementMsg.AgreementSeq)

		msgItem, ok := t.msgItems[announceAgreementMsg.MsgID]
		if !ok {
			logger.Errorf("msg id [%s] not exist in msg item map", announceAgreementMsg.MsgID)
			return
		}
		msgItem.AgreementSeqNum = announceAgreementMsg.AgreementSeq
		msgItem.Agreed = true

		t.holdQueueLocker.Lock()
		defer t.holdQueueLocker.Unlock()

		item, ok := t.holdQueueMap[announceAgreementMsg.MsgID]
		if !ok {
			logger.Errorf("msg id [%s] not exist in hold queue map", announceAgreementMsg.MsgID)
			return
		}

		t.holdQueue.Update(item, announceAgreementMsg.AgreementSeq, announceAgreementMsg.ProcessID, true)

		t.bmulticast.emittersLock.Lock()
		defer t.bmulticast.emittersLock.Unlock()

		for t.holdQueue.Len() > 0 {
			item := t.holdQueue.Peek().(*TOHoldQueueItem)
			_, ok := t.bmulticast.emitters[item.processID]
			if !ok {
				delete(t.msgItems, item.msgID)
				delete(t.holdQueueMap, item.msgID)
				heap.Pop(t.holdQueue)
				logger.Infof("skip crashed process [%s] msg", item.processID)
				continue
			}
			msg, ok := t.msgItems[item.msgID]
			if !ok || !msg.Agreed {
				break
			}
			logger.Errorf("to deliver [%d:%s][%s]", item.proposalSeqNum, item.processID, item.msgID)
			msgBody := msg.Body
			t.deliver <- msgBody
			delete(t.msgItems, item.msgID)
			delete(t.holdQueueMap, item.msgID)
			heap.Pop(t.holdQueue)
		}
	})
}
