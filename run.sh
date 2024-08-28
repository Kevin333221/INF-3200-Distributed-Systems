#!/bin/bash

# Check arguments is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <number_of_nodes>"
  exit 1
fi

# Check that number is positive integer
if ! [[ "$1" =~ ^[0-9]+$ ]] || [ "$1" -le 0 ]; then
  echo "Error: Argument must be a positive integer."
  exit 1
fi

# Get available nodes
nodes=$(/share/ifi/available-nodes.sh)

# Exit if no available nodes
if [ -z "$nodes" ]; then
  echo "Error: No available nodes."
  exit 1
fi

# Shuffle and select arg number of nodes
random_nodes=$(echo "$nodes" | shuf -n "$1")

# Initialize and format list
node_list="["

for node in $random_nodes; 
do
    # Get random port
    port=$(shuf -i 49152-65535 -n1)
    nodePort="$node:$port"

    # Start server on node with random port
    ##ssh -f $node "cd $PWD && python3 server.py $port" && echo "Server started on $node:$port"

    # Addd node to node list
    node_list+="'$nodePort',"
done

# Format the node list
node_list=${node_list::-1}
node_list+="]"