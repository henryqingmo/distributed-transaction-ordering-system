package main

import (
	"bufio"
	"cs425_mp1/internal/config"
	"fmt"
	"log"
	"os"
	"strings"
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

		switch action {
		case "DEPOSIT":
			account := fields[1]
			amount := fields[2]
			fmt.Printf("Action: %s, Account: %s, Amount: %s\n", action, account, amount)
		case "TRANSFER":
			account1 := fields[1]
			account2 := fields[3]
			amount := fields[4]
			fmt.Printf("Action: %s, Account1: %s, Account2: %s, Amount: %s\n", action, account1, account2, amount)
		}
	}

	for _, n := range parsed.Nodes {
		fmt.Printf("ID: %s, Host: %s, Port: %s\n", n.ID, n.Host, n.Port)
	}

}
