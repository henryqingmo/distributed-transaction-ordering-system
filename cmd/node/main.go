package main

import (
	config "cs425_mp1/internal/config"
	node "cs425_mp1/internal/node"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "./mp1_node <identifier> <configuration file>")
		os.Exit(1)
	}

	//name := os.Args[1]
	filePath := os.Args[2]

	parsed, err := config.ParseConfig(filePath)
	if err != nil {
		log.Fatalf("parse config: %v", err)
	}

	//for _, n := range parsed.Nodes {
	//fmt.Printf("ID: %s, Host: %s, Port: %s\n", n.ID, n.Host, n.Port)
	//}
	identifier := os.Args[1]

	nodeInfo, err := config.ParseIdentifier(parsed, identifier)
	if err != nil {
		log.Fatalf("parse identifier: %v", err)
	}

	n := node.NewNode(nodeInfo, parsed)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		path := fmt.Sprintf("latency_%s.txt", identifier)
		if err := n.FlushLatencies(path); err != nil {
			log.Printf("flush latencies: %v", err)
		}
		os.Exit(0)
	}()

	n.Run()

}
