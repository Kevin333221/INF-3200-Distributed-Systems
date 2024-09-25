import matplotlib.pyplot as plt

def plot(x, y, title, xlabel, ylabel):
    plt.plot(x, y)
    plt.title(title)
    plt.xlabel(xlabel)
    plt.ylabel(ylabel)
    plt.show()

def get_data(data):
    with open(data, 'r') as f:
        lines = f.readlines()
        x = [int(line.split()[0]) for line in lines]
        y = [float(line.split()[1]) for line in lines]
    return x, y