package multicast

import (
	"container/heap"
	"context"
	"sync"

	"github.com/bamboovir/cs425/lib/mp1/metrics"
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
	MsgID          string
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
	waitProposalCounter             map[string]chan ProposalItem
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
		waitProposalCounter:             map[string]chan ProposalItem{},
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

func (t *TotalOrding) aggregateVotesAndMulticast(votes []*ProposalItem) error {
	agreementSeqItem, err := MaxOfArrayProposalItem(votes)
	if err != nil {
		return err
	}

	announceAgreementMsg := NewTOAnnounceAgreementSeqMsg(
		agreementSeqItem.ProcessID,
		agreementSeqItem.ProposalSeqNum,
		agreementSeqItem.MsgID,
	)

	announceAgreementMsgBytes, err := announceAgreementMsg.Encode()
	if err != nil {
		return err
	}

	// logger.Errorf("announce %s %d", announceAgreementMsg.MsgID, announceAgreementMsg.AgreementSeq)
	err = t.rmulticast.Multicast(AnnounceAgreementSeqPath, announceAgreementMsgBytes)
	if err != nil {
		return err
	}

	return nil
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
	waitCounterChannel := make(chan ProposalItem)

	t.waitProposalCounterLock.Lock()
	defer t.waitProposalCounterLock.Unlock()

	t.waitProposalCounter[askMsg.MsgID] = waitCounterChannel

	go func() {
		votes := make([]*ProposalItem, 0)
		memberUpdateNotification := t.bmulticast.MembersUpdate()

		for {
			select {
			case membersCountI := <-memberUpdateNotification:
				var membersCount int

				switch t := membersCountI.(type) {
				case int:
					membersCount = t
				default:
					membersCount = -1
				}

				if len(votes) < membersCount {
					continue
				}

				t.aggregateVotesAndMulticast(votes)
				return
			case vote := <-waitCounterChannel:
				votes = append(votes, &vote)

				membersCount := t.bmulticast.MemberCount()
				if len(votes) < membersCount {
					continue
				}

				t.aggregateVotesAndMulticast(votes)
				return
			}
		}
	}()

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

		waitCounterChannel, ok := t.waitProposalCounter[replyProposalMsg.MsgID]

		if !ok {
			logger.Errorf("to msg id not exist")
			return
		}

		waitCounterChannel <- ProposalItem{
			ProposalSeqNum: replyProposalMsg.ProposalSeq,
			ProcessID:      replyProposalMsg.ProcessID,
			MsgID:          replyProposalMsg.MsgID,
		}
	})

	rRouter.Bind(AnnounceAgreementSeqPath, func(msg RMsg) {
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
			logger.Infof("TO deliver [%d:%s][%s]", item.proposalSeqNum, item.processID, item.msgID)
			metrics.NewDelayLogEntry(t.bmulticast.group.SelfNodeID, item.msgID).Log()
			msgBody := msg.Body
			t.deliver <- msgBody
			delete(t.msgItems, item.msgID)
			delete(t.holdQueueMap, item.msgID)
			heap.Pop(t.holdQueue)
		}
	})
}
