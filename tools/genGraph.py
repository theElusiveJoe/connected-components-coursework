import random
import pandas as pd
import uuid
import sys

def genGraph(nodes_num, edges_num):
    nodes = [str(i) for i in range(nodes_num)]

    l1, l2 = [random.choice(nodes) for _ in range(edges_num)], [random.choice(nodes) for _ in range(edges_num)]
    df = pd.DataFrame({
        "node1": l1, 
        "node2": l2
    })
    filename = f'tests/graphs/synthGraph-{nodes_num}-{edges_num}-{str(uuid.uuid1())[:8]}.csv'
    df.to_csv(
        filename, 
        index=False, header=False
    )
    print(filename)


if __name__ == '__main__':
    for i in range(2, 51, 3):
        for j in range(1, 100, 3):
            genGraph(i, j)