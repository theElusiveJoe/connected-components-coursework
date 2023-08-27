import random
import pandas as pd
import uuid
import sys
import math
import tqdm

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
    lines = list(map(lambda x: x[:34].strip(), lines[:len(lines)//2]))
    print("total nodes:", len(lines))
    df = pd.DataFrame({
        "node1": lines, 
        "node2": lines
    })
    filename = f'tests/biggraphs/synthBig2.csv'
    df.to_csv(
        filename, 
        index=False, header=False
    )

def genBigGraph(p, q):
    # 2*10^9 узлов
    # 1.7*10^9 ребер
    # E?N = 1.7/2 = 0.85

    dencity = 0.85
    nodesNum = p*10**q
    l = int(math.log2(nodesNum))
    edgesNum = int(nodesNum*dencity)
    targetStrLen = 34
    print('-> generating', nodesNum, 'nodes and', edgesNum, 'edges')
    
    s = f'{0:0{l}b}'
    tobinst = lambda x: s.format(x)

    l1, l2 = [], []
    nodeMax = int(nodesNum*1.42)
    nodesSet = set()
    for i in tqdm.tqdm(range(edgesNum)):
        v1, v2 = random.randint(0, nodeMax), random.randint(0, nodeMax)
        nodesSet.add(v1)
        nodesSet.add(v2)
        s1, s2 = bin(v1), bin(v2)
        s1 += 'x'*(targetStrLen - len(s1))
        s2 += 'x'*(targetStrLen - len(s2))
        l1.append(s1)
        l2.append(s2)
    
    print("-> generated", len(nodesSet), 'nodes and', edgesNum, 'edges')
    # return 

    df = pd.DataFrame({
        "node1": l1, 
        "node2": l2
    })
    filename = f'tests/biggraphs/big-{p}e{q}.csv'
    print('saving to', filename)
    df.to_csv(
        filename, 
        index=False, header=False
    )

def genBigGraph2(p, q):
    # 2*10^9 узлов
    # 1.7*10^9 ребер
    # E?N = 1.7/2 = 0.85

    dencity = 0.85
    nodesNum = p*10**q
    l = int(math.log2(nodesNum))
    edgesNum = int(nodesNum*dencity)
    targetStrLen = 34
    print('-> generating', nodesNum, 'nodes and', edgesNum, 'edges')
    
    s = f'{0:0{l}b}'

    nodeMax = int(nodesNum*1.42)
    nodesSet = set()
    filename = f'tests/biggraphs/big-{p}e{q}.csv'

    with open(filename, 'w') as f:
        for i in tqdm.tqdm(range(edgesNum)):
            v1, v2 = random.randint(0, nodeMax), random.randint(0, nodeMax)
            nodesSet.add(v1)
            nodesSet.add(v2)
            s1, s2 = bin(v1), bin(v2)
            s1 += 'x'*(targetStrLen - len(s1))
            s2 += 'x'*(targetStrLen - len(s2))
            f.write(s1)
            f.write(', ')
            f.write(s2)
            f.write('\n')
    
    print("-> generated", len(nodesSet), 'nodes and', edgesNum, 'edges')

if __name__ == '__main__':
    genBigGraph2(int(sys.argv[1]),int(sys.argv[2]))