echo "================== RUNNING=================="

BEST_FIT="bestfit"
OVER_SUB="oversub"
PROPOSED="proposed"
ONE_SHOT="oneshot"
WORST_FIT="worstfit"

oversub=1.5
nodeNum=500
cpuPerNode=64
memPerNode=128
tick=60
metricsTick=60
maxTaskLengthSeconds=3600 # seconds.
totalPodNumber=10000
workloadSubsetFactor=10
isDebug=true
path="/ssd/projects/google-trace-data"
runSim(){
    start="2019-01-01T00:00:00+09:00"
    end="2019-01-31T00:00:00+09:00"
    startTrace="600000000"
    ./gen_config.sh $1 "." $nodeNum $cpuPerNode $memPerNode $tick $metricsTick "$start"
    go run $(go list ./...) --config="./config/cluster_$1" \
    --workload="$path/workload"  \
    --scheduler="$1" \
    --isgen=$2 \
    --oversub=$oversub \
    --istrace=$3 \
    --trace="$path/tasks" \
    --start="$start" \
    --end="$end" \
    --trace-start="$startTrace" \
    --tick="$tick" \
    --max-task-length="$maxTaskLengthSeconds" \
    --total-pods-num=$totalPodNumber \
    --subset-factor=$workloadSubsetFactor\
    &> run_${1}.out
}
#rm -rf *.out
SECONDS=0
runSim $WORST_FIT false true
echo "$WORST_FIT took $SECONDS seconds"

SECONDS=0 
runSim $OVER_SUB false false
echo "$OVER_SUB took $SECONDS seconds"

SECONDS=0 
runSim $ONE_SHOT false false
echo "$ONE_SHOT took $SECONDS seconds"

SECONDS=0 
echo "==================Plotting=================="
python plotResults.py
echo "plotResults.py took $SECONDS seconds"
echo "==================FINISHED=================="
