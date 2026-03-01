package main

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
