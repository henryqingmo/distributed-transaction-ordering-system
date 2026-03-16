#!/bin/bash
# run_experiment.sh <nodeID> [hz]
# Example: ./run_experiment.sh node1 0.5
HZ=${2:-0.5}
python3 -u gentx.py $HZ | timeout 100 ./mp1_node $1 config.txt
echo "Done. Latency saved to latency_$1.txt"
