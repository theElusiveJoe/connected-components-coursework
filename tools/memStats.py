import psutil
import time
import pandas as pd
import sys


def fin_mpi_procs(ppid):
    procs = list()
    for proc in psutil.process_iter(['pid', 'name', 'username', 'cmdline']):
        if proc.ppid() == ppid:
            procs.append(proc)
    return procs[1:] 


def ps_mem_proc(proc):
    return proc.memory_full_info().rss / 1024 // 1024


def ps_mem(procs):
    mem_used = [ps_mem_proc(proc) for proc in procs]
    return [int(time.time()*1000)] + mem_used


def log_procs_mem_consumpiton(ppid):
    try:
        procs = fin_mpi_procs(ppid)
        print('here procs:', [p.pid for p in procs])
        mem_df = pd.DataFrame(columns=['time'] + [proc.pid for proc in procs])
        
        while True:
            try:
                new_row = ps_mem(procs)
                mem_df.loc[len(mem_df)] = new_row
                print(new_row)
                time.sleep(0.05)
            except:
                mem_df = mem_df.set_index(mem_df['time'])
                mem_df = mem_df.drop(['time'], axis=1)
                mem_df.to_csv('mem.csv')
                sys.exit(0)
    except Exception as e:
        print(e)

if __name__ == '__main__':
    ppid = int(sys.argv[1])
    log_procs_mem_consumpiton(ppid)