import random
import pandas as pd
import uuid
import sys

def genGraph1(nodes_num, edges_num):
    nodes = [str(i) for i in range(nodes_num)]

    l1, l2 = [random.choice(nodes) for _ in range(edges_num)], [random.choice(nodes) for _ in range(edges_num)]
    df = pd.DataFrame({
        "node1": l1, 
        "node2": l2
    })
    filename = f'tests/graphs1/synthGraph-{edges_num}e-{nodes_num}n.csv'
    df.to_csv(
        filename, 
        index=False, header=False
    )
    print(filename)

def genGraph2(nodes_num, edges_num, levels):
    l1, l2 = [], []
    for lvl in range(65, 65+levels):
        nodes = [chr(lvl)+str(i) for i in range(nodes_num)]
        n1, n2 = [random.choice(nodes) for _ in range(edges_num)], [random.choice(nodes) for _ in range(edges_num)]
        l1.extend(n1)
        l2.extend(n2)

    df = pd.DataFrame({
        "node1": l1, 
        "node2": l2
    })
    filename = f'tests/graphs2/synthGraph-{levels}l-{edges_num}e.csv'
    df.to_csv(
        filename, 
        index=False, header=False
    )
    print(filename)

def edgesFromBigNodes():
    with open('tests/nodes/1500k.csv') as f:
        lines = f.readlines()
    lines = list(map(lambda x: x[:34].strip(), lines))

    df = pd.DataFrame({
        "node1": lines, 
        "node2": lines
    })
    filename = f'tests/biggraphs/synthBig1.csv'
    df.to_csv(
        filename, 
        index=False, header=False
    )

if __name__ == '__main__':
    # for lvl in range(1, 30):
    #     for i in range(10, 200, 10):
    #         genGraph2(i, i, lvl)

    # for nodes in range(1000, 30000, 1000):
    #     for edges in range(1000, 30000, 1000):
    #         genGraph1(nodes, edges)
    
    edgesFromBigNodes()