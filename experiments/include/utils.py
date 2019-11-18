import os
import numpy as np
import multiprocessing as mp
import math 

## files
def list_files(path, extension):
    files = []
    files = []
    for r, d, f in os.walk(path):
        for file in f:
            if extension in file:
                files.append(os.path.join(r, file))

    return files

## compute CDF
class CDF: 
    def __init__(self, x, p): 
        self.x = x
        self.p = p   

def compute_cdf(data):
    data_sorted = np.sort(data)
    # calculate the proportional values of samples
    p = 1. * np.arange(len(data)) / (len(data) - 1)
    return CDF(data_sorted, p)

## parrallel computing

def run_parrallel(worker_num, func, arg_array):
    num_run = len(arg_array)/worker_num
    if len(arg_array) % worker_num > 0:
        num_run = num_run + 1

    results = []
    for i in range(0, num_run):
        pool = mp.Pool(worker_num)
        start = i*worker_num
        end = (i+1)*worker_num
        if end > len(arg_array):
            end = len(arg_array)
        tmp_res = pool.map(func, [arg for arg in arg_array[start:end]])
        
        results.extend(tmp_res)
    
    return results
