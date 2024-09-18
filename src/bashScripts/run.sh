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
num_available_nodes=${#node_array[@]}

# Initialize and format list
node_list="["

cd ../DeployServers

address_list="["

counter=0
while [[ $counter -lt $num_available_nodes ]]; do

    # Extract key-value pairs from the current line
    port=$(shuf -i 49152-65535 -n1)

    # Get node from array, wrap around if needed
    node=${node_array[$((counter % num_available_nodes))]}
    nodePort="$node:$port"

    # Add node to node list
    address_list+="\"$nodePort\","

    counter=$((counter + 1))
done

address_list=${address_list::-1}
address_list+="]"

echo $address_list

# Get node IDs
go build && ./DeployServers $2 $address_list $num_available_nodes

python3 ../DeployServers/get_ids.py

until [ -f ../DeployServers/node_ids.txt ]
do
    sleep 1
done

counter=0
while IFS= read -r line; do
    id=$line

    # Get node from array, wrap around if needed
    node=${node_array[$((counter % num_available_nodes))]}

    # Start server on node
    ssh -f $node "cd $PWD/.. && go run Server.go $id $2 && echo Server started on $node"

    counter=$((counter + 1))

done < node_ids.txt

rm node_ids.txt

# counter=0
# while counter < $num_available_nodes; do

#     # Get node from array, wrap around if needed
#     node=${node_array[$((counter % num_available_nodes))]}

#     # Start server on node with random port in either Python or Go
#     ssh -f $node "cd $PWD/.. && go run server.go $port $id && echo Server started on $node:$port"

#     counter=$((counter + 1))
# done

# # Read and process each line of Nodes.json
# counter=0
# while IFS= read -r line; do

#     # Extract key-value pairs from the current line
#     id=$(echo "$line" | sed -n 's/.*"id":\([0-9]*\).*/\1/p')
#     port=$(shuf -i 49152-65535 -n1)

#     # Get node from array, wrap around if needed
#     node=${node_array[$((counter % num_available_nodes))]}
#     nodePort="$node:$port"
    
#     # Start server on node with random port in either Python or Go
#     ssh -f $node "cd $PWD/.. && go run server.go $port $id && echo Server started on $node:$port"

#     # Add node to node list
#     node_list+="\"$nodePort\","
#     counter=$((counter + 1))

# done < Nodes.json