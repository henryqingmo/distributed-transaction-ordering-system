package ordering

import (
	manager "cs425_mp1/internal/network"
	"sort"
	"time"
)

type Queue struct {
	items []*QueueItem
}

type QueueItem struct {
	id           string
	tx           manager.MsgTransaction
	priority     float64
	deliverable  bool
	sender       string
	DeliveryTime time.Time // set when item is dequeued for delivery
}

type ISISOrdering struct {
	holdbackQueue Queue
	messageMap    map[string]*QueueItem
	counter       float64
	proposals     map[string]*proposalState
	numNodes      int
}

type proposalState struct {
	maxPriority float64
	count       int
}

func (q *Queue) Enqueue(item *QueueItem) {
	q.items = append(q.items, item)
	q.Sort()
}

func (q *Queue) Dequeue() *QueueItem {
	x := q.items[0]
	q.items = q.items[1:]
	return x
}

func (q *Queue) Peek() *QueueItem {
	if len(q.items) == 0 {
		return nil
	}
	return q.items[0]
}

func NewISISOrdering(numNodes int) *ISISOrdering {
	return &ISISOrdering{
		holdbackQueue: Queue{},
		messageMap:    make(map[string]*QueueItem),
		proposals:     make(map[string]*proposalState),
		counter:       0,
		numNodes:      numNodes,
	}
}

func NewQueueItem(id string, tx manager.MsgTransaction, priority float64, deliverable bool, sender string) *QueueItem {
	return &QueueItem{
		id:          id,
		tx:          tx,
		priority:    priority,
		deliverable: deliverable,
		sender:      sender,
	}
}

func (o *ISISOrdering) HandleMessage(nodeID string, msg manager.Message) *Outbound {
	switch msg.Type {
	case manager.TypeTransaction:
		return o.OnReceiveTransaction(nodeID, msg)
	case manager.TypePropose:
		return o.OnReceivePropose(msg)
	case manager.TypeAgree:
		return o.onReceiveAgree(msg)
	}
	return nil

}

func (q *Queue) Sort() {
	sort.Slice(q.items, func(i, j int) bool {
		if q.items[i].priority == q.items[j].priority {
			return q.items[i].sender < q.items[j].sender
		}
		return q.items[i].priority < q.items[j].priority
	})
}

type DeliveryResult struct {
	Tx           manager.MsgTransaction
	DeliveryTime time.Time
}

func (o *ISISOrdering) DeliveryReady() []DeliveryResult {
	var ready []DeliveryResult
	for len(o.holdbackQueue.items) > 0 {
		item := o.holdbackQueue.Peek()
		if !item.deliverable {
			break
		}
		item.DeliveryTime = time.Now()
		dequeued := o.holdbackQueue.Dequeue()
		ready = append(ready, DeliveryResult{Tx: dequeued.tx, DeliveryTime: dequeued.DeliveryTime})
	}
	return ready
}

type Outbound struct {
	To  string
	Msg manager.Message
}

func (o *ISISOrdering) OnReceiveTransaction(nodeID string, msg manager.Message) *Outbound {
	o.counter++
	item := NewQueueItem(
		msg.Transaction.MsgId,
		msg.Transaction,
		o.counter,
		false,
		msg.Transaction.Sender,
	)
	o.holdbackQueue.Enqueue(item)
	o.messageMap[msg.Transaction.MsgId] = item
	// Send TypePropose back to msg.Transaction.Sender
	return &Outbound{
		To: msg.Transaction.Sender,
		Msg: manager.NewPropose(
			msg.Transaction.MsgId,
			o.counter,
			nodeID,
		),
	}
}

func (o *ISISOrdering) OnReceivePropose(msg manager.Message) *Outbound {
	msgID := msg.Propose.MsgId
	state, ok := o.proposals[msgID]
	if !ok {
		state = &proposalState{}
		o.proposals[msgID] = state
	}
	state.count++
	if msg.Propose.ProposedPriority > state.maxPriority {
		state.maxPriority = msg.Propose.ProposedPriority
	}
	if state.count == o.numNodes {
		delete(o.proposals, msgID)
		return &Outbound{
			To:  "",
			Msg: manager.NewAgree(msgID, state.maxPriority),
		}
	}
	return nil
}

// PeerFailed decrements the expected proposal count and returns any TypeAgree
// messages that can now be finalized because enough proposals have been collected.
func (o *ISISOrdering) PeerFailed() []*Outbound {
	if o.numNodes > 1 {
		o.numNodes--
	}
	var out []*Outbound
	for msgID, state := range o.proposals {
		if o.numNodes > 0 && state.count >= o.numNodes {
			delete(o.proposals, msgID)
			out = append(out, &Outbound{
				To:  "",
				Msg: manager.NewAgree(msgID, state.maxPriority),
			})
		}
	}
	return out
}

func (o *ISISOrdering) onReceiveAgree(msg manager.Message) *Outbound {
	item, ok := o.messageMap[msg.Agree.MsgId]
	if !ok {
		return nil
	}
	item.priority = msg.Agree.AgreedPriority
	item.deliverable = true
	o.holdbackQueue.Sort()

	if msg.Agree.AgreedPriority > o.counter {
		o.counter = msg.Agree.AgreedPriority
	}
	return nil
}
