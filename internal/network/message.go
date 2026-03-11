package manager

type MessageType int

const (
	TypeTransaction MessageType = iota
	TypePropose
	TypeAgree
)

type Message struct {
	Type        MessageType
	Transaction *MsgTransaction
	Propose     *MsgPropose
	Agree       *MsgAgree
}

type Txkind int

const (
	Deposit Txkind = iota
	Transfer
)

type MsgTransaction struct {
	Kind    Txkind
	Account string // for Deposit
	Amount  int
	Source  string // only for Transfer
	Dest    string // only for Transfer
}

type MsgPropose struct {
	MsgId            string
	ProposedPriority int
	FromNode         string
}

type MsgAgree struct {
	MsgId          string
	AgreedPriority int
	FinalNodeID    string
}

func NewTransfer(source, dest string, amount int) Message {

	return Message{
		Type: TypeTransaction,
		Transaction: &MsgTransaction{
			Kind:   Transfer,
			Source: source,
			Dest:   dest,
			Amount: amount,
		},
	}
}

func NewDeposit(account string, amount int) Message {
	return Message{
		Type: TypeTransaction,
		Transaction: &MsgTransaction{
			Kind:    Deposit,
			Account: account,
			Amount:  amount,
		},
	}
}
