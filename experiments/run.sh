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
    ./genConfig.sh $1 $nodeNum $cpuPerNode $memPerNode $tick $metricsTick
    go run $(go list ./...) --config="./config/cluster_$1" \
    --workload="./config/workload"  \
    --scheduler="$1" \
    --isgen=$2 \
    --oversub=$oversub \
    &> run_${1}.out
}
#rm -rf *.out
runSim $BEST_FIT false
runSim $OVER_SUB false &
runSim $PROPOSED false &
wait
echo "Simulation tooks $SECONDS seconds"
echo "==================FINISHED=================="

echo "==================Plotting=================="
python plotResults.py
echo "==================FINISHED=================="
