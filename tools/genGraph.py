import random
import pandas as pd
import uuid

def genGraph(nodes_num, edges_num):
    nodes = [str(i) for i in range(nodes_num)]

    l1, l2 = [random.choice(nodes) for _ in range(edges_num)], [random.choice(nodes) for _ in range(edges_num)]
    df = pd.DataFrame({
        "node1": l1, 
        "node2": l2
    })

    df.to_csv(
        f'tests/graphs/synthGraph({nodes_num}:{edges_num}){uuid.uuid1()}.csv', 
        index=False, header=False
    )


if __name__ == '__main__':
    for _ in range(100):
        nodes_num = random.randint(1, 20)
        edges_num = random.randint(1, 50)
        genGraph(nodes_num, edges_num)