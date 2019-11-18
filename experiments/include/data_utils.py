from pandas import read_csv
import numpy

def read_machine_csv(f):
    cols = ['startTime','endTime','cpuUsage','cpuReq','memUsage','memReq']
    cUsages = []
    cReqs = []
    mUsages = []
    mReqs = []
    cCaps = []
    mCaps = []
    df = read_csv(f, header=None, index_col=False, names=cols)
    c = 0
    m = 0
    for index, event in df.iterrows():
        if index == 0:
            c = float(event[cols[0]])
            m = float(event[cols[1]])
            continue

        if(float(event['cpuReq'])>0): 
            cUsages.append(float(event['cpuUsage']))
            cReqs.append(float(event['cpuReq'])) 
            cCaps.append(c)
        if(float(event['memReq'])>0):     
            mUsages.append(float(event['memUsage'])) 
            mReqs.append(float(event['memReq'])) 
            mCaps.append(m) 

    res = (cUsages, cReqs, mUsages, mReqs, cCaps, mCaps)
    return res


def read_task_csv(f):
    cols = ['startTime','endTime','cpuUsage','memUsage','diskUsage']
    cUsages = []
    cReqs = []
    mUsages = []
    mReqs = []   
    cMaxUsages = []
    mMaxUsages = []
    cUsageStds = []
    mUsageStds = []

    df = read_csv(f, header=None, index_col=False, names=cols)
    for index, event in df.iterrows():
        if index == 0:
            cReqs.append(float(event[cols[1]]))
            mReqs.append(float(event[cols[2]]))
            continue

        cUsages.append(float(event['cpuUsage']))
        mUsages.append(float(event['memUsage']))

    # compute std
    cMaxUsages.append(numpy.max(cUsages))
    mMaxUsages.append(numpy.max(mUsages))
    cUsageStds.append(numpy.std(cUsages)/numpy.mean(cUsages))
    mUsageStds.append(numpy.std(mUsages)/numpy.mean(mUsages))

    ## clear memory...
    cUsages = []
    mUsages = []

    res = (cUsages, cReqs, mUsages, mReqs, cMaxUsages, mMaxUsages, cUsageStds, cUsageStds, mUsageStds )
    return res    