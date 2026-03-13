package node

import (
	"bufio"
	"cs425_mp1/internal/config"
	manager "cs425_mp1/internal/network"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Node struct {
	identifier     config.NodeInfo
	parsed         config.Parsed
	networkManager manager.Manager
	msgCounter     int
}

func NewNode(identifier config.NodeInfo, parsed config.Parsed) *Node {
	manager := manager.NewManager(identifier, 1024)
	manager.ConnectToPeers(parsed.Nodes)

	return &Node{
		identifier: identifier,
		parsed:     parsed,
		msgCounter: 0,
	}
}

func (n *Node) Run() {

	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		// Expect: "<Action> <account> <amount>"
		if len(strings.Fields(line)) < 3 {
			log.Printf("skipping malformed line: %q", line)
			continue
		}
		fields := strings.Fields(line)

		action := fields[0]

		/* TODO:
		Create network manager and start listening
		Connect to all peers
		Create the ISIS orderer and the ledger
		Start reading stdin in a goroutine, pushing transactions onto a channel
		Enter the main event loop
		*/

		n.msgCounter++

		MsgID := n.identifier.ID + "-" + strconv.Itoa(n.msgCounter)

		switch action {
		case "DEPOSIT":
			account := fields[1]
			amount, err := strconv.Atoi(fields[2])
			if err != nil {
				fmt.Println("conversion error:", err)
			}
			manager.NewDeposit(MsgID, account, amount)
		case "TRANSFER":
			account1 := fields[1]
			account2 := fields[3]
			amount, err := strconv.Atoi(fields[4])
			if err != nil {
				fmt.Println("conversion error:", err)
			}
			manager.NewTransfer(MsgID, account1, account2, amount)
		}

		select {
		case msg := <-n.networkManager.Inbox():
		}
	}
}
