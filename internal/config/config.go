package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type node struct {
	ID   string
	Host string
	Port string
}

type Parsed struct {
	Nodes []node
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
		if len(fields) != 3 {
			return Parsed{}, fmt.Errorf("malformed line: %q", line)
		}

		n := node{
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
