echo "================== RUNNING=================="
SECONDS=0; 
BEST_FIT="bestfit"
OVER_SUB="oversub"
PROPOSED="proposed"

oversub=1.5
nodeNum=20
cpuPerNode=64
memPerNode=128
tick=1
metricsTick=1
runSim(){
    clock="2019-01-01T00:00:00+09:00"
    startTrace="0"
    ./genConfig.sh $1 $nodeNum $cpuPerNode $memPerNode $tick $metricsTick "$clock"
    go run $(go list ./...) --config="./config/cluster_$1" \
    --workload="./config/workload"  \
    --scheduler="$1" \
    --isgen=$2 \
    --oversub=$oversub \
    --istrace=$3 \
    --trace="/Users/tanle/projects/google-trace-analysis/results/tasks" \
    --clock="$clock" \
    --trace-start="$startTrace" \
    &> run_${1}.out
}
#rm -rf *.out
runSim $BEST_FIT false true
# runSim $OVER_SUB false false &
# runSim $PROPOSED false false &
wait
echo "Simulation tooks $SECONDS seconds"
echo "==================FINISHED=================="

echo "==================Plotting=================="
python plotResults.py
echo "==================FINISHED=================="
