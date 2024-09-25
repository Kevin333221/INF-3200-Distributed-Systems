"""Module for plotting the results of the PUT and GET requests.
"""

from collections import defaultdict
import os
import matplotlib.pyplot as plt


def average_time(file: str) -> dict:
    """Reads the file and calculates the average time and standard deviation for each numer of nodes.

    Args:
        file (str): Path to the file containing the time data.

    Returns:
        dict: A dictionary containing the average time and SD for each number of nodes.
    """
    time_data = defaultdict(lambda: {"times": []})

    with open(file, "r", encoding="utf-8") as f:
        for line in f:
            try:
                node, time = map(float, line.split())
                time_data[int(node)]["times"].append(time)
            except ValueError:
                print(f"Skipping invalid line: {line.strip()}")
                continue

    statistics = {}
    for node, data in time_data.items():
        avg_time = round(sum(data["times"]) / len(data["times"]), 2)
        stddev = round(
            (sum((x - avg_time) ** 2 for x in data["times"]) / len(data["times"]))
            ** 0.5,
            2,
        )
        statistics[node] = {"average": avg_time, "stddev": stddev}

    return statistics


def make_plot(
    put_stats: dict, get_stats: dict, filename: str = "time_plot.pdf"
) -> None:
    """Creates a plot of the average time and SD for PUT and GET requests.
    Plots are saved in a PDF file.

    Args:
        put_times (dict): The average time for PUT requests.
        get_times (dict): The average time for GET requests.
        filename (str, optional): Where to store plot. Defaults to 'time_plot.pdf'.

    Raises:
        ValueError: If there is no data to plot.
    """
    if len(put_stats) == 0 or len(get_stats) == 0:
        print("No data to plot.")
        raise ValueError("No data to plot.")

    put_nodes = list(put_stats.keys())
    put_avg = [put_stats[node]["average"] for node in put_nodes]
    put_stddev = [put_stats[node]["stddev"] for node in put_nodes]

    get_nodes = list(get_stats.keys())
    get_avg = [get_stats[node]["average"] for node in get_nodes]
    get_stddev = [get_stats[node]["stddev"] for node in get_nodes]

    plt.errorbar(
        put_nodes,
        put_avg,
        yerr=put_stddev,
        fmt="-o",
        color="orange",
        label="PUT",
        capsize=10,
    )
    plt.errorbar(
        get_nodes,
        get_avg,
        yerr=get_stddev,
        fmt="-o",
        color="blue",
        label="GET",
        capsize=10,
    )

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
