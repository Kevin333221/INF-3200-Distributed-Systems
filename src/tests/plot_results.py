import os
import matplotlib.pyplot as plt
from collections import defaultdict


def average_time(file: str) -> dict:
    time_data = defaultdict(lambda: {'time': 0, 'count': 0})

    with open(file) as f:
        for line in f:
            try:
                node, time = map(float, line.split())
                time_data[int(node)]['time'] += time
                time_data[int(node)]['count'] += 1
            except ValueError:
                print(f"Skipping invalid line: {line.strip()}")
                continue

        average_times = {}
        for node, data in time_data.items():
            average_times[node] = round(data['time'] / data['count'], 2)

    return average_times

def make_plot(put_times: dict, get_times: dict, filename: str = 'time_plot.pdf') -> None:
    if(len(put_times) == 0 or len(get_times) == 0):
        print("No data to plot.")
        raise ValueError("No data to plot.")

    plt.plot(put_times.keys(), put_times.values(), color='red', label='PUT')
    plt.plot(get_times.keys(), get_times.values(), color='blue', label='GET')

    plt.xlabel('Number of nodes')
    plt.ylabel('Average time (ms)')
    plt.title('Average time for PUT and GET requests')

    plt.legend()
    plt.savefig(filename, format='pdf')
    plt.close()
    

if __name__ == '__main__':
    if not os.path.exists('PUT_log.txt') :
        print("'PUT_log.txt' not found.")
        exit(1)
    if not os.path.exists('GET_log.txt') :
        print("'GET_log.txt' not found.")
        exit(1)

    put_times = average_time('PUT_log.txt')
    print(f"PUT times: {put_times}")
    get_times = average_time('GET_log.txt')
    print(f"GET times: {get_times}")

    make_plot(put_times, get_times)