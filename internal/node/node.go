package node

import (
	"bufio"
	ledger "cs425_mp1/internal/account"
	"cs425_mp1/internal/config"
	manager "cs425_mp1/internal/network"
	ordering "cs425_mp1/internal/ordering"
	"cs425_mp1/internal/timing"
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
	recorder       *timing.Recorder
}

func NewNode(identifier config.NodeInfo, parsed config.Parsed) *Node {
	mgr := manager.NewManager(identifier, 1024)
	// Listen must be called before ConnectToPeers so other nodes can reach us.
	if err := mgr.Listen(); err != nil {
		log.Fatalf("listen: %v", err)
	}
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
		recorder:       timing.NewRecorder(),
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
		return manager.NewDeposit(msgID, n.identifier.ID, fields[1], amount), true
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
		return manager.NewTransfer(msgID, n.identifier.ID, fields[1], fields[3], amount), true
	default:
		log.Printf("unknown action %q", fields[0])
		return manager.Message{}, false
	}
}

func (n *Node) applyAndPrint(result ordering.DeliveryResult) {
	tx := result.Tx
	n.recorder.Record(tx.OriginTime, result.DeliveryTime)
	if tx.Kind == manager.Deposit {
		n.ledger.Deposit(tx.Account, tx.Amount)
	} else {
		n.ledger.Transfer(tx.Source, tx.Dest, tx.Amount)
	}
	fmt.Println(n.ledger.Balances())
}

func (n *Node) FlushLatencies(path string) error {
	return n.recorder.Flush(path)
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
				txCh = nil
				continue
			}
			// Sender must be node ID (used for routing propose replies back to originator).
			msg.SenderID = n.identifier.ID
			msg.Transaction.Sender = n.identifier.ID
			n.networkManager.Broadcast(msg)

			// Originator also processes the transaction locally to contribute its proposal.
			propOut := n.ordering.HandleMessage(n.identifier.ID, msg)
			if propOut != nil {
				// Feed the local proposal back into the ordering layer.
				agreeOut := n.ordering.HandleMessage(n.identifier.ID, propOut.Msg)
				if agreeOut != nil {
					// All proposals collected (single-node or last proposal was ours).
					n.networkManager.Broadcast(agreeOut.Msg)
					n.ordering.HandleMessage(n.identifier.ID, agreeOut.Msg)
					for _, result := range n.ordering.DeliveryReady() {
						n.applyAndPrint(result)
					}
				}
			}

		case msg := <-n.networkManager.Inbox():
			out := n.ordering.HandleMessage(n.identifier.ID, msg)
			if out != nil {
				if out.To == "" {
					// Originator is broadcasting TypeAgree — also apply it locally.
					n.networkManager.Broadcast(out.Msg)
					n.ordering.HandleMessage(n.identifier.ID, out.Msg)
				} else {
					n.networkManager.Send(out.To, out.Msg)
				}
			}
			for _, result := range n.ordering.DeliveryReady() {
				n.applyAndPrint(result)
			}

		case id := <-n.networkManager.Failures():
			log.Printf("peer %s died", id)
			// Reduce the quorum and check if any pending proposals can now finalize.
			for _, agreeOut := range n.ordering.PeerFailed() {
				n.networkManager.Broadcast(agreeOut.Msg)
				n.ordering.HandleMessage(n.identifier.ID, agreeOut.Msg)
				for _, result := range n.ordering.DeliveryReady() {
					n.applyAndPrint(result)
				}
			}
		}
	}
}
