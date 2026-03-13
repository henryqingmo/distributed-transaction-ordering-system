package ordering

import (
	manager "cs425_mp1/internal/network"
	"sort"
)

type Queue struct {
	items []*QueueItem
}

type QueueItem struct {
	id          string
	tx          manager.MsgTransaction
	priority    float64
	deliverable bool
	sender      string
}

type isisOrdering struct {
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

func NewISISOrdering(numNodes int) *isisOrdering {
	return &isisOrdering{
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

func (o *isisOrdering) HandleMessage(nodeID string, msg manager.Message) {
	switch msg.Type {
	case manager.TypeTransaction:
		o.OnReceiveTransaction(nodeID, msg)
	case manager.TypePropose:
		o.OnReceivePropose(msg)
	case manager.TypeAgree:
		o.onReceiveAgree(msg)
	}

}

func (q *Queue) Sort() {
	sort.Slice(q.items, func(i, j int) bool {
		if q.items[i].priority == q.items[j].priority {
			return q.items[i].sender < q.items[j].sender
		}
		return q.items[i].priority < q.items[j].priority
	})
}

func (o *isisOrdering) DeliveryReady() []*QueueItem {
	var ready []*QueueItem
	for len(o.holdbackQueue.items) > 0 {
		item := o.holdbackQueue.Peek()

		if !item.deliverable {
			break
		}
		ready = append(ready, o.holdbackQueue.Dequeue())
	}
	return ready
}

type Outbound struct {
	To  string
	Msg manager.Message
}

func (o *isisOrdering) OnReceiveTransaction(nodeID string, msg manager.Message) Outbound {
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
	return Outbound{
		To:	msg.Transaction.Sender,
		Msg: manager.NewPropose(
			msg.Transaction.MsgId,
			o.counter, 
			nodeID,
		),
	}
}

func (o *isisOrdering) OnReceivePropose(msg manager.Message) *Outbound {
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

func (o *isisOrdering) onReceiveAgree(msg manager.Message) {
	item, ok := o.messageMap[msg.Agree.MsgId]
	if !ok {
		return
	}
	item.priority = msg.Agree.AgreedPriority
	item.deliverable = true
	o.holdbackQueue.Sort()

	if msg.Agree.AgreedPriority > o.counter {
		o.counter = msg.Agree.AgreedPriority
	}
}
