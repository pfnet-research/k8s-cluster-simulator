
scheduler=$1
log_file="kubesim_$scheduler.log"
filePath="$2/config/cluster_$scheduler.yaml"
node_num=$3
cpu=$4
mem=$5
tick=$6
metricsTick=$7
clock="$8"

echo """# Log level defined by sirupsen/logrus.
# Optional (info, debug)
logLevel: info

# Interval duration for scheduling and updating the cluster, in seconds.
# Optional (default: 10)
tick: $tick

# Start time at which the simulation starts, in RFC3339 format.
# Optional (default: now)
startClock: $clock

# Interval duration for logging metrics of the cluster, in seconds.
# Optional (default: same as tick)
metricsTick: $metricsTick

# Metrics of simulated kubernetes cluster is written
# to standard out, standard error or files at given paths.
# The metrics is formatted with the given formatter.
# Optional (default: not writing metrics)
metricsLogger:
- dest: $log_file
  formatter: JSON
#- dest: kubesim-hr-${scheduler}.log
#  formatter: humanReadable
#- dest: stdout
#  formatter: table

# Write configuration of each node.
cluster:
""" > $filePath

for i in `seq 1 $node_num`
do
    echo """
- metadata:
    name: node-$i
    labels:
      beta.kubernetes.io/os: simulated
    annotations:
      foo: bar
  spec:
    unschedulable: false
    # taints:
    # - key: k
    #   value: v
    #   effect: NoSchedule
  status:
    allocatable:
      cpu: $cpu
      memory: ${mem}Gi
      nvidia.com/gpu: 0
      pods: 99999
""">> $filePath
done