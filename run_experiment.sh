#!/bin/bash
# run_experiment.sh nodeX
./mp1_node $1 config.txt &
NODE_PID=$!
sleep 3 # wait for all nodes to connect
python3 -u gentx.py 0.5 &
GENTX_PID=$!
sleep 100
kill $GENTX_PID 2>/dev/null
kill -SIGTERM $NODE_PID # flushes latency_$1.txt
wait $NODE_PID
