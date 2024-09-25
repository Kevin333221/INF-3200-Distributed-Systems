"""Module for plotting the results of the PUT and GET requests.
"""

from collections import defaultdict
import os
import matplotlib.pyplot as plt


def average_time(file: str) -> dict:
    """Reads the line from the file and calculates the average time for each node.

    Args:
        file (str): Path to the file containing the time data.

    Returns:
        dict: A dictionary containing the average time for each number of nodes.
    """
    time_data = defaultdict(lambda: {"time": 0, "count": 0})

    with open(file, "r", encoding="utf-8") as f:
        for line in f:
            try:
                node, time = map(float, line.split())
                time_data[int(node)]["time"] += time
                time_data[int(node)]["count"] += 1
            except ValueError:
                print(f"Skipping invalid line: {line.strip()}")
                continue

        average_times = {}
        for node, data in time_data.items():
            average_times[node] = round(data["time"] / data["count"], 2)

    return average_times


def make_plot(
    put_times: dict, get_times: dict, filename: str = "time_plot.pdf"
) -> None:
    """Creates a plot of the average time for PUT and GET requests.
    Plots are saved in a PDF file.

    Args:
        put_times (dict): The average time for PUT requests.
        get_times (dict): The average time for GET requests.
        filename (str, optional): Where to store plot. Defaults to 'time_plot.pdf'.

    Raises:
        ValueError: If there is no data to plot.
    """
    if len(put_times) == 0 or len(get_times) == 0:
        print("No data to plot.")
        raise ValueError("No data to plot.")

    plt.plot(put_times.keys(), put_times.values(), color="red", label="PUT")
    plt.plot(get_times.keys(), get_times.values(), color="blue", label="GET")

    plt.xlabel("Number of nodes")
    plt.ylabel("Average time (ms)")
    plt.title("Average time for PUT and GET requests")

    plt.legend()
    plt.savefig(filename, format="pdf")
    plt.close()


if __name__ == "__main__":
    if not os.path.exists("PUT_log.txt"):
        print("'PUT_log.txt' not found.")
        exit(1)
    if not os.path.exists("GET_log.txt"):
        print("'GET_log.txt' not found.")
        exit(1)

    p_times = average_time("PUT_log.txt")
    print(f"PUT times: {p_times}")
    g_times = average_time("GET_log.txt")
    print(f"GET times: {g_times}")

    make_plot(p_times, g_times)
