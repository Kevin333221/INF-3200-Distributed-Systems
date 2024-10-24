import sys
import json
import datetime
import http.client
import asyncio

async def main():

    # Convert string to list
    test_nodes = json.loads(sys.argv[1])

    # test_key(test_nodes)
    test_join_leave(test_nodes, 20, 2)

def test_key(test_nodes):
    join_all_nodes(test_nodes)

    key = "test_key"
    value = "test_value"

    response, data = PUT_key(test_nodes[0], key, value)
    if response.status != 200:
        print(f"Failed to PUT key {key} to node {test_nodes[0]}")
        return
    
    response, data = GET_key(test_nodes[0], key)
    if response.status != 200:
        print(f"Failed to GET key {key} from node {test_nodes[0]}")
        return
    

    if data != value:
        print(f"GET key {key} from node {test_nodes[0]} returned incorrect value")
        return
    
    key2 = "test_key2"
    value2 = "test_value2"

    response, data2 = PUT_key(test_nodes[1], key2, value2)
    if response.status != 200:
        print(f"Failed to PUT key {key2} to node {test_nodes[1]}")
        return
    
    response, data2 = GET_key(test_nodes[1], key2)
    if response.status != 200:
        print(f"Failed to GET key {key2} from node {test_nodes[1]}")
        return

    if data2 != value2:
        print(f"GET key {key} from node {test_nodes[0]} returned incorrect value")
        return
    
    leave_all_nodes(test_nodes)
    join_all_nodes(test_nodes)

    response, data = GET_key(test_nodes[0], key)
    if response.status != 200:
        print(f"Failed to GET key {key} from node {test_nodes[0]}")
        return
    
    if data != value:
        print(f"GET key {key} from node {test_nodes[0]} returned incorrect value")
        return
    
    response, data2 = GET_key(test_nodes[1], key2)
    if response.status != 200:
        print(f"Failed to GET key {key2} from node {test_nodes[1]}")
        return
    
    if data2 != value2:
        print(f"GET key {key2} from node {test_nodes[1]} returned incorrect value")
        return
    
    leave_all_nodes(test_nodes)

def PUT_key(node, key, value):
    conn = http.client.HTTPConnection(node)
    conn.request("PUT", f"/storage/{key}", value)
    response = conn.getresponse()
    data = response.read().decode()
    conn.close()
    return response, data

def GET_key(node, key):
    conn = http.client.HTTPConnection(node)
    conn.request("GET", f"/storage/{key}")
    response = conn.getresponse()
    data = response.read().decode()
    conn.close()
    return response, data

def join_all_nodes(test_nodes):
    for node in test_nodes[1:]:
        response, _ = join_network(node, test_nodes[0])
        if response.status != 200:
            print(f"Failed to join {node} to network. Skipping...")
            continue

def test_join_leave(test_nodes, test, amount=10):

    leave_all_nodes(test_nodes)

    timings = test_requests(test_nodes[0:amount], test)

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