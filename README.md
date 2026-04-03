# CS425 MP1 — Distributed Ledger with ISIS Total Ordering

A distributed ledger system built for CS425 (Distributed Systems) at UIUC. Multiple nodes coordinate `DEPOSIT` and `TRANSFER` transactions using the **ISIS total ordering protocol** to maintain consistent account balances across a cluster, even in the presence of node failures.

## Overview

Each node accepts transactions on stdin, broadcasts them to all peers, and uses ISIS ordering to ensure every node applies transactions in the same global order. When a node fails mid-protocol, the remaining nodes detect the failure and finalize any pending transactions.

## Architecture

```
stdin → Node.Run() → parse tx → ISIS ordering → apply to Ledger → print BALANCES
                                      ↕
                             Network Manager (TCP)
                                   ↕
                            peer nodes (same flow)
```

### Key Packages

| Package | Responsibility |
|---|---|
| `internal/node` | Orchestrates everything: reads stdin, drives ISIS, applies ordered transactions |
| `internal/ordering` | ISIS total-order broadcast — holdback queue, propose/agree phases, failure handling |
| `internal/network` | TCP manager: `Send`, `Broadcast`, `Inbox`, `Failures` |
| `internal/account` | Thread-safe ledger: `Deposit`, `Transfer`, `Balances` |
| `internal/config` | Parses `config.txt` |
| `internal/timing` | Records end-to-end transaction latency |

### ISIS Protocol

1. **Broadcast** — originator sends `TypeTransaction` to all nodes
2. **Propose** — each node assigns a local priority and replies `TypePropose` to the originator
3. **Agree** — originator collects all proposals, picks the max, broadcasts `TypeAgree`
4. **Deliver** — nodes mark the message deliverable and deliver in priority order

On node failure, the cluster decrements the expected proposal count and finalizes any messages that now have enough proposals.

## Building & Running

```bash
# Build
make build          # compiles to bin/node

# Run a node
./bin/node <nodeID> config.txt
```

### Config File Format

```
node1 sp25-cs425-0101.cs.illinois.edu 1234
node2 sp25-cs425-0102.cs.illinois.edu 1234
```

### Transaction Input (stdin)

```
DEPOSIT alice 100
TRANSFER alice -> bob 50
```

### Balances Output

```
BALANCES alice:50 bob:50
```

Accounts are sorted alphabetically; zero-balance accounts are omitted.

## Testing

```bash
go test ./internal/account   # ledger unit tests
go test ./...                # all tests

python3 gentx.py             # generate test transactions
python3 sanity.py            # validate output format
```

## Experiments

Latency data from failure and no-failure runs (3-node and 8-node clusters) are stored under `experiments/`.
