package manager

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

// yanze: Length-prefixed JSON encode/decode helpers
// WriteMsg length-prefixes a JSON-encoded Message onto w.
// Wire format: [4-byte big-endian length][JSON payload]
func WriteMsg(w io.Writer, msg Message) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	length := uint32(len(payload))
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return err
	}
	_, err = w.Write(payload)
	return err
}

// ReadMsg reads one length-prefixed JSON Message from r.
func ReadMsg(r io.Reader) (Message, error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return Message{}, err
	}
	if length == 0 || length > 4*1024*1024 {
		return Message{}, fmt.Errorf("invalid message length: %d", length)
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return Message{}, err
	}
	var msg Message
	if err := json.Unmarshal(buf, &msg); err != nil {
		return Message{}, err
	}
	return msg, nil
}

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
