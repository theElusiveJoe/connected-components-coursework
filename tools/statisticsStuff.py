import random

for _ in range(100):
    s = set()
    for _ in range(10**6):
        x = random.randint(0, 10**8)
        s.add(x)

    print(len(s)/10**7, len(s))