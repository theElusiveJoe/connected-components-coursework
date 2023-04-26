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
    filename = f'tests/graphs/synthGraph({nodes_num}:{edges_num}){uuid.uuid1()}.csv'
    df.to_csv(
        filename, 
        index=False, header=False
    )
    print(filename)


if __name__ == '__main__':
    genGraph(
        int(sys.argv[1]), 
        int(sys.argv[2])
    )