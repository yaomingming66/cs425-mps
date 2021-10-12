package multicast

import (
	"container/heap"
	"context"
	"time"

	sync "github.com/sasha-s/go-deadlock"

	"github.com/bamboovir/cs425/lib/mp1/metrics"
	"github.com/bamboovir/cs425/lib/mp1/router"
	errors "github.com/pkg/errors"
)

const (
	AskProposalSeqPath       = "/total-ording/ask-proposal-seq"
	WaitProposalSeqPath      = "/total-ording/wait-proposal-seq"
	AnnounceAgreementSeqPath = "/total-ording/announce-agreement-seq"
)

const (
	RetryRemoveMax   = 25
	NodeCrashTimeout = 6 * time.Second
)

type ProposalItem struct {
	ProposalSeqNum uint64
	ProcessID      string
	MsgID          string
	Body           []byte
}

type TotalOrding struct {
	bmulticast                      *BMulticast
	rmulticast                      *RMulticast
	router                          *router.Router
	holdQueueMap                    map[string]*TOHoldQueueItem
	holdQueue                       *TOHoldPriorityQueue
	holdQueueLocker                 *sync.Mutex
	maxAgreementSeqNumOfGroup       uint64
	maxAgreementSeqNumOfGroupLocker *sync.Mutex
	maxProposalSeqNumOfSelf         uint64
	maxProposalSeqNumOfSelfLocker   *sync.Mutex
	waitProposalCounter             map[string][]*ProposalItem
	waitProposalCounterLock         *sync.Mutex
	waitVotesChannel                chan *ProposalItem
	crashNodeTimeout                map[string]time.Time
}

func NewTotalOrder(b *BMulticast, r *RMulticast) *TotalOrding {
	holdQueue := &TOHoldPriorityQueue{}
	heap.Init(holdQueue)

	return &TotalOrding{
		bmulticast:                      b,
		rmulticast:                      r,
		router:                          router.New(),
		holdQueueMap:                    map[string]*TOHoldQueueItem{},
		holdQueue:                       holdQueue,
		holdQueueLocker:                 &sync.Mutex{},
		maxAgreementSeqNumOfGroup:       0,
		maxAgreementSeqNumOfGroupLocker: &sync.Mutex{},
		maxProposalSeqNumOfSelf:         0,
		maxProposalSeqNumOfSelfLocker:   &sync.Mutex{},
		waitProposalCounter:             map[string][]*ProposalItem{},
		waitProposalCounterLock:         &sync.Mutex{},
		waitVotesChannel:                make(chan *ProposalItem, 10000),
		crashNodeTimeout:                map[string]time.Time{},
	}
}

func (t *TotalOrding) Start(ctx context.Context) (err error) {
	t.bindTODeliver()
	err = t.rmulticast.Start(ctx)
	if err != nil {
		return err
	}
	memberUpdateChannel := t.bmulticast.MembersUpdate()
	go t.collectVotes(memberUpdateChannel)
	return nil
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

	// logger.Infof("announce %s %d", announceAgreementMsg.MsgID, announceAgreementMsg.AgreementSeq)
	err = t.rmulticast.Multicast(AnnounceAgreementSeqPath, announceAgreementMsg)
	if err != nil {
		return err
	}

	return nil
}

func (t *TotalOrding) collectVotes(memberUpdateChannel chan interface{}) {
	for {
		select {
		case vote := <-t.waitVotesChannel:
			// logger.Errorf("get vote: [%s]", vote.MsgID)
			_, ok := t.waitProposalCounter[vote.MsgID]
			if !ok {
				t.waitProposalCounter[vote.MsgID] = []*ProposalItem{}
			}

			membersCount := t.bmulticast.MemberCount()

			t.waitProposalCounter[vote.MsgID] = append(t.waitProposalCounter[vote.MsgID], vote)
			// logger.Errorf("curr vote: [%d]", len(t.waitProposalCounter[vote.MsgID]))
			if len(t.waitProposalCounter[vote.MsgID]) < membersCount {
				continue
			}

			err := t.aggregateVotesAndMulticast(t.waitProposalCounter[vote.MsgID])
			if err != nil {
				logger.Errorf("aggregate votes and multicast failed for msg [%s]: %v", vote.MsgID, err)
			}
			delete(t.waitProposalCounter, vote.MsgID)
		case membersCountI := <-memberUpdateChannel:
			var membersCount int
			switch t := membersCountI.(type) {
			case int:
				membersCount = t
			default:
				membersCount = -1
			}

			logger.Infof("members count update to %d, re-check vote count", membersCount)
			for msgID, votes := range t.waitProposalCounter {
				if len(votes) < membersCount {
					continue
				}
				err := t.aggregateVotesAndMulticast(votes)
				if err != nil {
					logger.Errorf("aggregate votes and multicast failed for msg [%s]: %v", msgID, err)
				}
				delete(t.waitProposalCounter, msgID)
			}
		}
	}
}

func TOMsgDecodeWrapper(f func(*TOMsg) error) func(interface{}) error {
	return func(v interface{}) error {
		msg := v.(*TOMsg)
		return f(msg)
	}
}

func (r *TotalOrding) Bind(path string, f func(msg *TOMsg) error) {
	r.router.Bind(path, TOMsgDecodeWrapper(f))
}

func (t *TotalOrding) Multicast(path string, v interface{}) (err error) {
	tomsg, err := NewTOMsg(path, v)
	if err != nil {
		return errors.Wrap(err, "to-multicast failed")
	}
	tomsgBytes, err := tomsg.Encode()
	if err != nil {
		return errors.Wrap(err, "to-multicast failed")
	}
	askMsg := NewTOAskProposalSeqMsg(t.bmulticast.group.SelfNodeID, tomsgBytes)

	err = t.rmulticast.Multicast(AskProposalSeqPath, askMsg)
	if err != nil {
		return errors.Wrap(err, "to-multicast failed")
	}
	return nil
}

func (t *TotalOrding) bindTODeliver() {
	rRouter := t.rmulticast
	bRouter := t.bmulticast

	rRouter.Bind(AskProposalSeqPath, func(msg *RMsg) error {
		askMsg := &TOAskProposalSeqMsg{}
		_, err := askMsg.Decode(msg.Body)
		if err != nil {
			return errors.Wrap(err, "ask-proposal-seq failed")
		}

		t.maxAgreementSeqNumOfGroupLocker.Lock()
		t.maxProposalSeqNumOfSelfLocker.Lock()

		proposalSeqNum := MaxUint64(t.maxAgreementSeqNumOfGroup, t.maxProposalSeqNumOfSelf) + 1
		t.maxProposalSeqNumOfSelf = proposalSeqNum
		t.maxAgreementSeqNumOfGroupLocker.Unlock()
		t.maxProposalSeqNumOfSelfLocker.Unlock()

		t.holdQueueLocker.Lock()
		defer t.holdQueueLocker.Unlock()

		item := &TOHoldQueueItem{
			body:           askMsg.Body,
			proposalSeqNum: proposalSeqNum,
			msgID:          askMsg.MsgID,
			processID:      askMsg.SrcID,
			agreed:         false,
		}
		t.holdQueueMap[askMsg.MsgID] = item
		heap.Push(t.holdQueue, item)

		// logger.Errorf("send proposal seq [%s] [%d] to [%s]", askMsg.MsgID, proposalSeqNum, askMsg.SrcID)
		replyProposalMsg := NewTOReplyProposalSeqMsg(t.bmulticast.group.SelfNodeID, askMsg.MsgID, proposalSeqNum)
		err = t.bmulticast.Unicast(askMsg.SrcID, WaitProposalSeqPath, replyProposalMsg)
		if err != nil {
			return errors.Wrap(err, "ask-proposal-seq failed")
		}
		return nil
	})

	bRouter.Bind(WaitProposalSeqPath, func(msg *BMsg) error {
		replyProposalMsg := &TOReplyProposalSeqMsg{}
		_, err := replyProposalMsg.Decode(msg.Body)
		if err != nil {
			return errors.Wrap(err, "wait-proposal-seq failed")
		}
		// logger.Infof("get proposal seq: %s", replyProposalMsg.MsgID)
		t.waitVotesChannel <- &ProposalItem{
			ProposalSeqNum: replyProposalMsg.ProposalSeq,
			ProcessID:      replyProposalMsg.ProcessID,
			MsgID:          replyProposalMsg.MsgID,
		}

		return nil
	})

	rRouter.Bind(AnnounceAgreementSeqPath, func(msg *RMsg) error {
		announceAgreementMsg := &TOAnnounceAgreementSeqMsg{}
		_, err := announceAgreementMsg.Decode(msg.Body)
		if err != nil {
			return errors.Wrap(err, "announce-agreement-seq failed")
		}

		// logger.Infof("get announce [%d:%s] %s", announceAgreementMsg.AgreementSeq, announceAgreementMsg.ProcessID, announceAgreementMsg.MsgID)

		t.maxAgreementSeqNumOfGroupLocker.Lock()
		t.maxAgreementSeqNumOfGroup = MaxUint64(t.maxAgreementSeqNumOfGroup, announceAgreementMsg.AgreementSeq)
		t.maxAgreementSeqNumOfGroupLocker.Unlock()

		t.holdQueueLocker.Lock()
		defer t.holdQueueLocker.Unlock()

		item, ok := t.holdQueueMap[announceAgreementMsg.MsgID]
		if !ok {
			logger.Errorf("msg id [%s] not exist in hold queue map", announceAgreementMsg.MsgID)
			return errors.New("announce-agreement-seq failed")
		}

		t.holdQueue.Update(item, announceAgreementMsg.ProcessID, announceAgreementMsg.AgreementSeq)

		for t.holdQueue.Len() > 0 {
			// logger.Infof("hold queue %s", t.holdQueue.Snapshot())
			item := t.holdQueue.Peek().(*TOHoldQueueItem)
			if !item.agreed {
				ok := t.bmulticast.IsNodeAlived(item.processID)
				if !ok {
					crashTime, tok := t.crashNodeTimeout[item.processID]

					if !tok {
						t.crashNodeTimeout[item.processID] = time.Now()
						break
					}

					timeDiff := time.Since(crashTime)

					if timeDiff > NodeCrashTimeout {
						delete(t.holdQueueMap, item.msgID)
						heap.Pop(t.holdQueue)
						logger.Infof("skip crashed process [%s] msg", item.processID)
						break
					}
				}
				break
			}

			logger.Infof("TO deliver [%d:%s][%s]", item.proposalSeqNum, item.processID, item.msgID)
			metrics.NewDelayLogEntry(t.bmulticast.group.SelfNodeID, item.msgID).Log()
			delete(t.holdQueueMap, item.msgID)
			heap.Pop(t.holdQueue)
			tomsg := &TOMsg{}
			_, err = tomsg.Decode(item.body)
			if err != nil {
				return errors.Wrap(err, "announce-agreement-seq failed")
			}
			err = t.router.Run(tomsg.Path, tomsg)
			if err != nil {
				logger.Errorf("process err %v", err)
			}
		}

		return nil
	})
}
