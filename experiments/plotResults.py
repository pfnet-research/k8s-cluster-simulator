import json
import matplotlib.pyplot as plt

LOG_FILE = '/Users/tanle/go/src/github.com/pfnet-research/k8s-cluster-simulator/experiments/kubesim_bestfit.log'
cpuStr = 'cpu'

def loadLog(filepath) :
    cpuUsages = []
    memUsages = []
    gpuUsages = []
    cpuAllocatables = []
    requests = []
    busyNodes = []
    overloadNodes = []
    overBookNodes = []

    with open(filepath) as fp:
        content = fp.readlines()
        i = 0
        for line in content:
            data = json.loads(line)
            nodeDict = data['Nodes']
            busyNode = 0
            overloadNode = 0
            overBookNode = 0
            for nodeName, node in nodeDict.items():
                cpuUsage = 0
                cpuAllocatable = 0
                cpuRequest = 0
                
                runningPodsNum = int(node['RunningPodsNum'])

                usageDict = node['TotalResourceUsage']
                for rsName in usageDict:
                    if(rsName==cpuStr):
                        cpuUsage = int(usageDict[rsName])

                allocatableDict = node['Allocatable']    
                for rsName in allocatableDict:
                    if(rsName==cpuStr):
                        cpuAllocatable = int(allocatableDict[rsName])
                
                requestDict = node['TotalResourceRequest']    
                for rsName in requestDict:
                    if(rsName==cpuStr):
                        cpuRequest = int(requestDict[rsName])

                if(cpuUsage > cpuAllocatable):
                    overloadNode = overloadNode+1
           
                if(cpuRequest > cpuAllocatable):
                    overBookNode = overBookNode +1
           
                if(runningPodsNum > 0):
                    busyNode = busyNode + 1

            cpuUsages.append(cpuUsage)
            cpuAllocatables.append(cpuAllocatable)
            busyNodes.append(busyNode)
            overloadNodes.append(overloadNode) 
            overBookNodes.append(overBookNode)

            # line = fp.readline()
            i=i+1
    fp.close()

    return busyNodes, overloadNodes, overBookNodes, cpuUsages, cpuAllocatables

busyNodesBest, overloadNodesBest, overBookNodesBest, cpuUsagesBest, cpuAllocatablesBest = loadLog('kubesim_bestfit.log')
busyNodesOver, overloadNodesOver, overBookNodesOver, cpuUsagesOver, cpuAllocatablesOver = loadLog('kubesim_oversub.log')
busyNodesProp, overloadNodesProp, overBookNodesProp, cpuUsagesProp, cpuAllocatablesProp = loadLog('kubesim_proposed.log')
print("Overload nodes's num: oversub: "+str(sum(overloadNodesOver)) + " proposed: "+str(sum(overloadNodesProp)))
## resource usage
############# PLOTTING ##############
Y_MAX = 10
tick = 1
figPath = "./figs/"
FIG_ONE_COL = (8,6)
plots = [True, False, False]
## plot utilization: number of busy nodes.
if(plots[0]):
    fig_util = plt.figure(figsize=FIG_ONE_COL)
    plt.plot(range(0,len(busyNodesBest)*tick,tick), busyNodesBest)
    plt.plot(range(0,len(busyNodesOver)*tick,tick), busyNodesOver)
    plt.plot(range(0,len(busyNodesProp)*tick,tick), busyNodesProp)
    plt.legend(['bestfit', 'oversub', 'proposed'], loc='best')
    plt.xlabel('time (seconds)')
    plt.ylabel('Num. busy nodes')
    plt.ylim(0,Y_MAX)

    fig_util.savefig(figPath+"util.pdf", bbox_inches='tight')

## plot performance: number of overload nodes.
if(plots[1]):
    fig_perform = plt.figure(figsize=FIG_ONE_COL)
    plt.plot(range(0,len(overloadNodesBest)*tick,tick), overloadNodesBest)
    plt.plot(range(0,len(overloadNodesOver)*tick,tick), overloadNodesOver)
    plt.plot(range(0,len(overloadNodesProp)*tick,tick), overloadNodesProp)
    plt.legend(['bestfit', 'oversub', 'proposed'], loc='best')
    plt.xlabel('time (seconds)')
    plt.ylabel('Num. overload nodes')
    # plt.ylim(0,Y_MAX)

    fig_util.savefig(figPath+"perf.pdf", bbox_inches='tight')
    
## plot performance: number of overload nodes.
if(plots[2]):
    fig_overbook = plt.figure(figsize=FIG_ONE_COL)
    plt.plot(range(0,len(overBookNodesBest)*tick,tick), overBookNodesBest)
    plt.plot(range(0,len(overBookNodesOver)*tick,tick), overBookNodesOver)
    plt.plot(range(0,len(overBookNodesProp)*tick,tick), overBookNodesProp)
    plt.legend(['bestfit', 'oversub', 'proposed'], loc='best')
    plt.xlabel('time (seconds)')
    plt.ylabel('Num. overbook nodes')
    # plt.ylim(0,Y_MAX)

    fig_util.savefig(figPath+"overbook.pdf", bbox_inches='tight')


## show figures
plt.show()