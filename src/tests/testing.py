import os
import sys
import requests
import time

def check_addresses(addresses):

    # Set the failed flag to False
    failed = False

    # Iterate over the addresses
    for address in addresses:
        # Try to send a request to the address
        try:
            print(f"Requesting helloworld from {address}", end =" ")
            response = requests.get(f"http://{address}/helloworld")
            print(f"\u2714")
            if response.text != address:
                failed = True
        except Exception as e:
            print(f"\n\u2716 Request to {address} failed: {e}\n")
            failed = True

    return failed

def check_PUT_requests(addresses, keys):

    # Set the failed flag to False
    failed = False

    # Iterate over the addresses
    for index, key in enumerate(keys):

        address = addresses[index % len(addresses)]

        # Start the timer
        start_time = time.time()

        # Try to send a request to the address
        try:
            print(f"Sending key {keys[index]} to {address}", end =" ")
            response = requests.put(f"http://{address}/storage/{key}", data=f"Hello, World {index}!")
            total_time = (time.time() - start_time) * 1000

            if response.status_code != 200:
                failed = True

            print(f"\u2714 - {total_time:.2f} ms")
            with open("PUT_log.txt", "a") as f:
                f.write(f"{len(addresses)} {total_time:.2f}\n")

        except Exception as e:
            print(f"\n\u2716 Request to {address} failed: {e}\n")
            failed = True

    return failed

def check_GET_requests(addresses, keys):

    # Set the failed flag to False
    failed = False

    # Iterate over the addresses
    for index, key in enumerate(keys):

        # Get the address
        address = addresses[index % len(addresses)]

        # Start the timer
        start_time = time.time()

        # Try to send a request to the address
        try:
            print(f"Requesting key {key} from {address}", end =" ")
            response = requests.get(f"http://{address}/storage/{key}")
            total_time = (time.time() - start_time) * 1000

            if response.text != f"Hello, World {index}!":
                failed = True

            print(f"\u2714 - {total_time:.2f} ms")
            with open("GET_log.txt", "a") as f:
                f.write(f"{len(addresses)} {total_time:.2f}\n")

        except Exception as e:
            print(f"\n\u2716 Request to {address} failed: {e}\n")
            failed = True

    return failed

def check_PUT_requests_All(addresses, keys):
    for _ in range(3):
        start_time = time.time()
        check_PUT_requests(addresses, keys)
        total_time = time.time() - start_time
        print(f"PUT time: {total_time:.2f} seconds")
        
        with open("PUT_ALL_log.txt", "a") as f:
            f.write(f"{amount_of_addresses} {total_time:.2f}\n")

def check_GET_requests_All(addresses, keys):
    for _ in range(3):
        start_time = time.time()
        check_GET_requests(addresses, keys)
        total_time = time.time() - start_time
        print(f"GET time: {total_time:.2f} seconds")

        with open("GET_ALL_log.txt", "a") as f:
            f.write(f"{amount_of_addresses} {total_time:.2f}\n")

if __name__ == "__main__":
    
    # Get the list of addresses from the command line arguments one by one
    addresses = sys.argv[1:]
    amount_of_addresses = len(addresses)
    test_amount = 100
    keys = [os.urandom(8).hex() for _ in range(test_amount)]

    # print("Success!") if not check_addresses(addresses) else print("Failure")
    # print()

    # print("Success!") if not check_PUT_requests(addresses, keys) else print("Failure")
    # print()

    # print("Success!") if not check_GET_requests(addresses, keys) else print("Failure")

    check_PUT_requests_All(addresses, keys)
    check_GET_requests_All(addresses, keys)