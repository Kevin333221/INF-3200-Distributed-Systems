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
shuffled_nodes=$(echo "$nodes" | shuf -n "$1")
IFS=$'\n' read -r -d '' -a node_array <<< "$shuffled_nodes"

# Get number of available nodes
num_available_nodes=${#node_array[@]}

# Initialize and format list
node_list="["

# Get node IDs
go build && ./go_server $num_available_nodes

echo "Number of available nodes: $num_available_nodes"
echo ""

# Read and process each line of Nodes.json
counter=0
while IFS= read -r line; do

    # Extract key-value pairs from the current line
    id=$(echo "$line" | sed -n 's/.*"id":\([0-9]*\).*/\1/p')
    port=$(shuf -i 49152-65535 -n1)

    # Get node from array, wrap around if needed
    node=${node_array[$((counter % num_available_nodes))]}
    nodePort="$node:$port"
    
    # Start server on node with random port in either Python or Go
    ssh -f $node "cd $PWD && go build && cd .. && go run server.go $port $id && echo Server started on $node:$port"

    # echo "Server started on $node:$port with ID $id"

    # Add node to node list
    node_list+="\"$nodePort\","
    counter=$((counter + 1))

done < Nodes.json

# # Format the node list
# node_list=${node_list::-1}
# node_list+="]"

# # Print the list
# echo "Node list: $node_list"