import os
import sys
import json

def main():

    nodes_json = os.path.join(os.path.dirname(__file__), 'Nodes.json')

    if not os.path.exists(nodes_json):
        print("nodes.json not found")
        sys.exit(1)

    with open(nodes_json, 'r') as f:
        nodes = json.loads(f.read())

    ids = []

    for node in nodes:
        ids.append(node['id'])

    with open('node_ids.txt', 'w') as f:
        for id in ids:
            f.write(str(id) + '\n')

if __name__ == "__main__":
    main()