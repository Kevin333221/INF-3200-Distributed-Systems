import sys
import json
import datetime
import http.client
import asyncio
import time

async def main():

    # Convert string to list
    test_nodes = json.loads(sys.argv[1])

    leave_all_nodes(test_nodes)

    amount = 64

    timings = test_requests(test_nodes[0:amount], 20)

    with open(f"time_taken_to_join_and_leave_size_{amount}.txt", "w") as f:
        f.write(json.dumps(timings))

def test_requests(test_nodes, test):
    main_node = test_nodes[0]
    timings = []
    for _ in range(test):

        timeing_itteration_join = []
        timeing_itteration_leave = []

        for node in test_nodes[1:]:
            
            print(f"Joining {node} to network...")

            # Start timer
            start_time = datetime.datetime.now()

            # Join network
            response, _ = join_network(node, main_node)
            
            # Stop timer
            end_time = datetime.datetime.now()

            # Calculate time taken
            time_taken = end_time - start_time

            timeing_itteration_join.append(time_taken.total_seconds())

            if response.status != 200:
                print(f"Failed to join {node} to network. Skipping...")
                continue
                
        for node in test_nodes[1:]:
            print(f"Leaving {node} from network...")

            # Start timer
            start_time = datetime.datetime.now()

            # Join network
            response, _ = leave_network(node)
            
            # Stop timer
            end_time = datetime.datetime.now()

            # Calculate time taken
            time_taken = end_time - start_time

            timeing_itteration_leave.append(time_taken.total_seconds())

            if response.status != 200:
                print(f"Failed to join {node} to network. Skipping...")
                continue

        timings.append(timeing_itteration_join)
        timings.append(timeing_itteration_leave)

    return timings

def leave_all_nodes(test_nodes):
    for node in test_nodes:
        print(f"Leaving {node} from network...")
        response, _ = leave_network(node)
        if response.status != 200:
            print(f"Failed to join {node} to network. Skipping...")
            continue

def get_node_info(node):
    conn = http.client.HTTPConnection(node)
    conn.request("GET", "/node-info")
    response = conn.getresponse()
    data = response.read().decode()
    conn.close()
    return response, data

def join_network(node, main_node):
    conn = http.client.HTTPConnection(node)
    conn.request("POST", f"/join?nprime={main_node}")
    response = conn.getresponse()
    data = response.read().decode()
    conn.close()
    return response, data

def leave_network(node):
    conn = http.client.HTTPConnection(node)
    conn.request("POST", "/leave")
    response = conn.getresponse()
    data = response.read().decode()
    conn.close()
    return response, data

if __name__ == "__main__":
    asyncio.run(main())