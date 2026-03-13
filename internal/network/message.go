package manager

type MessageType int

const (
	TypeTransaction MessageType = iota
	TypePropose
	TypeAgree
)

type Message struct {
	Type        MessageType
	Transaction MsgTransaction
	Propose     MsgPropose
	Agree       MsgAgree
}

type Txkind int

const (
	Deposit Txkind = iota
	Transfer
)

type MsgTransaction struct {
	MsgId  string
	Sender string

	Kind    Txkind
	Account string // for Deposit
	Amount  int
	Source  string // only for Transfer
	Dest    string // only for Transfer
}

type MsgPropose struct {
	MsgId            string
	ProposedPriority float64
	FromNode         string
}

type MsgAgree struct {
	MsgId          string
	AgreedPriority float64
}

func NewTransfer(msgId, source, dest string, amount int) Message {

	return Message{
		Type: TypeTransaction,
		Transaction: MsgTransaction{
			Kind:   Transfer,
			Source: source,
			Dest:   dest,
			Amount: amount,
		},
	}
}

func NewDeposit(msgId, account string, amount int) Message {
	return Message{
		Type: TypeTransaction,
		Transaction: MsgTransaction{
			Kind:    Deposit,
			Account: account,
			Amount:  amount,
		},
	}
}

func NewPropose(msgID string, priority float64, fromNode string) Message {
	return Message{
		Type: TypePropose,
		Propose: MsgPropose{
			MsgId:            msgID,
			ProposedPriority: priority,
			FromNode:         fromNode,
		},
	}
}

func NewAgree(msgID string, priority float64) Message {
	return Message{
		Type: TypeAgree,
		Agree: MsgAgree{
			MsgId:          msgID,
			AgreedPriority: priority,
		},
	}
}
