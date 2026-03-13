package node

import (
	"bufio"
	ledger "cs425_mp1/internal/account"
	"cs425_mp1/internal/config"
	manager "cs425_mp1/internal/network"
	ordering "cs425_mp1/internal/ordering"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Node struct {
	identifier     config.NodeInfo
	parsed         config.Parsed
	networkManager *manager.Manager
	msgCounter     int
	ordering       *ordering.ISISOrdering
	ledger         *ledger.Ledger
}

func NewNode(identifier config.NodeInfo, parsed config.Parsed) *Node {
	mgr := manager.NewManager(identifier, 1024)
	mgr.ConnectToPeers(parsed.Nodes)
	ord := ordering.NewISISOrdering(len(parsed.Nodes))
	led := ledger.NewLedger()

	return &Node{
		identifier:     identifier,
		parsed:         parsed,
		msgCounter:     0,
		networkManager: mgr,
		ordering:       ord,
		ledger:         led,
	}
}

func (n *Node) parseLine(line string) (manager.Message, bool) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		log.Printf("skipping malformed line: %q", line)
		return manager.Message{}, false
	}
	n.msgCounter++
	msgID := n.identifier.ID + "-" + strconv.Itoa(n.msgCounter)

	switch fields[0] {
	case "DEPOSIT":
		amount, err := strconv.Atoi(fields[2])
		if err != nil {
			log.Printf("bad amount in: %q", line)
			return manager.Message{}, false
		}
		return manager.NewDeposit(msgID, fields[1], amount), true
	case "TRANSFER":
		if len(fields) < 5 {
			log.Printf("skipping malformed line: %q", line)
			return manager.Message{}, false
		}
		amount, err := strconv.Atoi(fields[4])
		if err != nil {
			log.Printf("bad amount in: %q", line)
			return manager.Message{}, false
		}
		return manager.NewTransfer(msgID, fields[1], fields[3], amount), true
	default:
		log.Printf("unknown action %q", fields[0])
		return manager.Message{}, false
	}
}

func (n *Node) Run() {
	txCh := make(chan manager.Message, 64)

	go func() {
		sc := bufio.NewScanner(os.Stdin)
		sc.Buffer(make([]byte, 0, 1024), 1024*1024)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			msg, ok := n.parseLine(line)
			if !ok {
				continue
			}
			txCh <- msg
		}
		close(txCh)
	}()

	for {
		select {
		case msg, ok := <-txCh:
			if !ok {
				return
			}
			msg.SenderID = n.identifier.ID
			msg.Transaction.Sender = n.identifier.Host
			n.networkManager.Broadcast(msg)
			out := n.ordering.HandleMessage(n.identifier.ID, msg)
			n.ordering.HandleMessage(n.identifier.ID, out.Msg)

			_ = msg
		case msg := <-n.networkManager.Inbox():
			out := n.ordering.HandleMessage(n.identifier.ID, msg)
			if out != nil {
				if out.To == "" {
					n.networkManager.Broadcast(out.Msg)
				} else {
					n.networkManager.Send(out.To, out.Msg)
				}
			}
			for _, tx := range n.ordering.DeliveryReady() {
				if tx.Kind == manager.Deposit {
					n.ledger.Deposit(tx.Account, tx.Amount)
				} else {
					n.ledger.Transfer(tx.Source, tx.Dest, tx.Amount)
				}
				fmt.Println(n.ledger.Balances())
			}
		case id := <-n.networkManager.Failures():
			log.Printf("peer %s died", id)
		}
	}
}
