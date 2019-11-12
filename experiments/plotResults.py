import json
import re
import matplotlib.pyplot as plt

cpuStr = 'cpu'

def loadLog(filepath) :
    cpuUsages = []
    maxCpuUsages = []
    cpuRequests = []
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
            totalCpuUsage = 0
            totalCapacity = 0
            maxCpuUsage = 0
            totalCpuRequest = 0
            for nodeName, node in nodeDict.items():
                cpuUsage = 0
                cpuAllocatable = 0
                cpuRequest = 0                
                runningPodsNum = int(node['RunningPodsNum'])

                usageDict = node['TotalResourceUsage']
                for rsName in usageDict:
                    if(rsName==cpuStr):
                        cpuUsage = formatQuatity(usageDict[rsName])
                        totalCpuUsage = totalCpuUsage+ cpuUsage
                        if cpuUsage > maxCpuUsage:
                            maxCpuUsage = cpuUsage

                allocatableDict = node['Allocatable']    
                for rsName in allocatableDict:
                    if(rsName==cpuStr):
                        cpuAllocatable = formatQuatity(allocatableDict[rsName])
                        totalCapacity = totalCapacity + cpuAllocatable
                
                requestDict = node['TotalResourceRequest']    
                for rsName in requestDict:
                    if(rsName==cpuStr):
                        cpuRequest = formatQuatity(requestDict[rsName])
                        totalCpuRequest = totalCpuRequest + cpuRequest

                if(cpuUsage > cpuAllocatable):
                    overloadNode = overloadNode+1
           
                if(cpuRequest > cpuAllocatable):
                    overBookNode = overBookNode +1
           
                if(runningPodsNum > 0):
                    busyNode = busyNode + 1

            cpuUsages.append(totalCpuUsage)
            cpuAllocatables.append(totalCapacity)
            busyNodes.append(busyNode)
            overloadNodes.append(overloadNode) 
            overBookNodes.append(overBookNode)
            maxCpuUsages.append(maxCpuUsage)
            cpuRequests.append(totalCpuRequest)

            # line = fp.readline()
            i=i+1
    fp.close()

    return busyNodes, overloadNodes, overBookNodes, cpuUsages, cpuRequests, maxCpuUsages, cpuAllocatables

def formatQuatity(str):
    strArray = re.split('(\d+)', str)
    val = float(strArray[1])
    scaleStr = strArray[2]
    if scaleStr != "":
        if(scaleStr == "m"):
            val = val/1000        
        elif (scaleStr == "Mi"):
            val = val/1024

    return val

methods = ["oneshot","worstfit","oversub"]
# methods = {"oneshot","worstfit"}
methodsNum = len(methods)
busyNodes = []
overloadNodes = []
overbookNodes = []
cpuUsages = []
maxCpuUsages = []
cpuAllocatables = []
cpuRequests = []

for m in methods:
    b, ol, ob, u, ur, mu, a = loadLog("kubesim_"+m+".log")
    busyNodes.append(b)
    overloadNodes.append(ol)
    overbookNodes.append(ob)
    cpuUsages.append(u)
    maxCpuUsages.append(mu)
    cpuAllocatables.append(a)
    cpuRequests.append(ur)

# busyNodesBest, overloadNodesBest, overBookNodesBest, cpuUsagesBest, cpuAllocatablesBest = loadLog('kubesim_bestfit.log')
# busyNodesOver, overloadNodesOver, overBookNodesOver, cpuUsagesOver, cpuAllocatablesOver = loadLog('kubesim_oversub.log')
# busyNodesProp, overloadNodesProp, overBookNodesProp, cpuUsagesProp, cpuAllocatablesProp = loadLog('kubesim_proposed.log')

# print("Overload nodes's num: oversub: "+str(sum(overloadNodesOver)) + " proposed: "+str(sum(overloadNodesProp)))
## resource usage
############# PLOTTING ##############

tick = 1
figPath = "/ssd/projects/cluster/figs/"
FIG_ONE_COL = (4,3)
plots = [True, False, False]
## plot utilization: number of busy nodes.
cap = 64
if(plots[0]):
    Y_MAX = cap*1.5
    fig_util = plt.figure(figsize=FIG_ONE_COL)
    max_len = 0
    for i in range(methodsNum):
        plt.plot(range(0,len(maxCpuUsages[i])*tick,tick), maxCpuUsages[i])
        if max_len < len(maxCpuUsages[i]):
            max_len = len(maxCpuUsages[i])
    
    plt.plot(range(0,max_len*tick,tick), [cap] * max_len)
    legends = methods
    legends.append('capacity')
    plt.legend(legends, loc='best')
    plt.xlabel('time (seconds)')
    plt.ylabel('max usage (cpu cores)')
    plt.ylim(0,Y_MAX)

    fig_util.savefig(figPath+"util.pdf", bbox_inches='tight')

if False:
    Y_MAX = cap*20
    fig_util = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(cpuRequests[i])*tick,tick), cpuRequests[i])
    
    plt.legend(methods, loc='best')
    plt.xlabel('time (seconds)')
    plt.ylabel('total cpu request')
    plt.ylim(0,Y_MAX)

    fig_util.savefig(figPath+"request.pdf", bbox_inches='tight')

## plot performance: number of overload nodes.
if False:
    fig_perform = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(maxCpuUsages[i])*tick,tick), maxCpuUsages[i])

    plt.legend(methods, loc='best')
    plt.xlabel('time (seconds)')
    plt.ylabel('Num. overload nodes')
    # plt.ylim(0,Y_MAX)

    fig_util.savefig(figPath+"perf.pdf", bbox_inches='tight')
    
## plot performance: number of overload nodes.
if False:
    fig_overbook = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(overloadNodes[i])*tick,tick), overloadNodes[i])

    plt.legend(methods, loc='best')
    plt.xlabel('time (seconds)')
    plt.ylabel('Num. overbook nodes')
    # plt.ylim(0,Y_MAX)

    fig_util.savefig(figPath+"overbook.pdf", bbox_inches='tight')

## show figures
plt.show()