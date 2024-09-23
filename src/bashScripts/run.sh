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

if [ -z "$2" ]; then
  echo "Usage: $0 $1 <identifier space (bit)>"
  exit 1
fi

# Check that number is positive integer
if ! [[ "$2" =~ ^[0-9]+$ ]] || [ "$2" -le 0 ]; then
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
amount_nodes=${#node_array[@]}

# Initialize and format list
node_list="["

cd ../DeployServers

go_list="["
addr_list=""

counter=0
while [[ $counter -lt $amount_nodes ]]; do

    # Extract key-value pairs from the current line
    port=$(shuf -i 49152-65535 -n1)

    # Get node from array, wrap around if needed
    node=${node_array[$((counter % amount_nodes))]}
    nodePort="$node:$port"

    # Add node to node list
    go_list+="\"$nodePort\","
    addr_list+="$nodePort "

    counter=$((counter + 1))
done

go_list=${go_list::-1}
go_list+="]"

echo $addr_list

# Get node IDs
go build && ./DeployServers $2 $go_list $amount_nodes

python3 ../DeployServers/get_ids.py

until [ -f ../DeployServers/node_ids.txt ]
do
    sleep 1
done

counter=0
while IFS= read -r line; do
    id=$line

    # Get node from array, wrap around if needed
    node=${node_array[$((counter % amount_nodes))]}

    # Start server on node
    ssh -f $node "cd $PWD/.. && go run Server.go $id && echo Server started on $node"

    counter=$((counter + 1))

done < node_ids.txt

rm node_ids.txt