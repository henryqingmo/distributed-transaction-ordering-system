package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type NodeInfo struct {
	ID   string
	Host string
	Port string
}

type Parsed struct {
	Nodes []NodeInfo
}

func ParseConfig(filePath string) (Parsed, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return Parsed{}, err
	}
	defer f.Close()

	var parsed Parsed

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 1 {
			// First line of spec format is the node count; skip it.
			continue
		}
		if len(fields) != 3 {
			return Parsed{}, fmt.Errorf("malformed line: %q", line)
		}

		n := NodeInfo{
			ID:   fields[0],
			Host: fields[1],
			Port: fields[2],
		}
		parsed.Nodes = append(parsed.Nodes, n)
	}

	if err := scanner.Err(); err != nil {
		return Parsed{}, err
	}
	return parsed, nil
}

func ParseIdentifier(parsed Parsed, identifier string) (NodeInfo, error) {
	for _, n := range parsed.Nodes {
		if n.ID == identifier {
			return n, nil
		}
	}
	return NodeInfo{}, fmt.Errorf("identifier %q not found", identifier)
}
