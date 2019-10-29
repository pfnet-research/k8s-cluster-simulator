package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	MaxInt8   = 1<<7 - 1
	MinInt8   = -1 << 7
	MaxInt16  = 1<<15 - 1
	MinInt16  = -1 << 15
	MaxInt32  = 1<<31 - 1
	MinInt32  = -1 << 31
	MaxInt64  = 1<<63 - 1
	MinInt64  = -1 << 63
	MaxUint8  = 1<<8 - 1
	MaxUint16 = 1<<16 - 1
	MaxUint32 = 1<<32 - 1
	MaxUint64 = 1<<64 - 1
)

func genNormFloat64(std, mean, min, max float64, r *rand.Rand) float64 {
	res := r.NormFloat64()*std + mean
	res = math.Min(res, max)
	res = math.Max(min, res)
	return res
}

func BuildClock(startClock string, shift int64) (clock.Clock, error) {
	clk := clock.NewClock(time.Now())

	if startClock != "" {
		c, err := time.Parse(time.RFC3339, startClock)
		if err != nil {
			return clk, err
		}
		clk = clock.NewClock(c)
	}

	clk = clk.Add(time.Duration(shift) * time.Second)

	return clk, nil
}

func ConvertTraceToPod(path string, csvFile string, startTimestamp string, cpuFactor int, memFactor int, timeStep int) *v1.Pod {
	// read csv files
	phases := []int{}
	cpuUsages := []int{}
	memUsages := []int{}
	requestCpu := 0
	requestMem := 0
	// Load a TXT file.
	f, _ := os.Open(fmt.Sprintf("%s/%s", path, csvFile))

	// Create a new reader.
	r := csv.NewReader(bufio.NewReader(f))
	// read first line
	firstLine, err := r.Read()
	if err == nil {
		cpu, _ := strconv.ParseFloat(firstLine[1], 64)
		mem, _ := strconv.ParseFloat(firstLine[2], 64)
		requestCpu = int(cpu * float64(cpuFactor))
		requestMem = int(mem * float64(memFactor))
	}
	// read the rest
	phaseNum := 0
	for {
		line, err := r.Read()

		// Stop at EOF.
		if err == io.EOF {
			break
		}
		cpu, _ := strconv.ParseFloat(line[2], 64)
		mem, _ := strconv.ParseFloat(line[3], 64)
		cpuUsage := int(cpu * float64(cpuFactor))
		memusage := int(mem * float64(memFactor))
		if phaseNum > 0 && cpuUsage == cpuUsages[phaseNum-1] && memusage == memUsages[phaseNum-1] {
			phases[phaseNum-1] = phases[phaseNum-1] + int(timeStep)
		} else {
			cpuUsages = append(cpuUsages, cpuUsage)
			memUsages = append(memUsages, memusage)
			phases = append(phases, timeStep)
			phaseNum = phaseNum + 1
		}
	}

	fileName := string(f.Name())
	strs := strings.Split(fileName, "_")
	jobIdx := strings.Split(strs[1], ".")[0]

	// create pods
	simSpec := ""
	for i := 0; i < phaseNum; i++ {
		sec := phases[i]
		cpu := cpuUsages[i]
		mem := memUsages[i]
		gpu := 0

		simSpec += fmt.Sprintf(`
- seconds: %d
  resourceUsage:
    cpu: %dm
    memory: %dMi
    nvidia.com/gpu: %d
`, sec, cpu, mem, gpu)
	}
	// prio := s.myrand.Int31n(3) / 2 // 0, 0, 1
	prio := int32(1) // TODO(tanle): nil memory if we set all pods'pirority as the same priority

	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("pod-%v", jobIdx),
			Namespace: "default",
			Annotations: map[string]string{
				"simSpec": simSpec,
			},
		},
		Spec: v1.PodSpec{
			Priority: &prio,
			Containers: []v1.Container{
				{
					Name:  "container",
					Image: "container",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							"cpu":            resource.MustParse(fmt.Sprintf("%dm", int(requestCpu))),
							"memory":         resource.MustParse(fmt.Sprintf("%dMi", int(requestMem))),
							"nvidia.com/gpu": resource.MustParse("0"),
						},
						Limits: v1.ResourceList{
							"cpu":            resource.MustParse("999"),
							"memory":         resource.MustParse("999Gi"),
							"nvidia.com/gpu": resource.MustParse("0"),
						},
					},
				},
			},
		},
	}

	return &pod
}

func WritePodAsJson(pod v1.Pod, path string, clock clock.Clock) {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "\t")
	err := encoder.Encode(pod)
	if err != nil {
		return
	}
	file, err := os.OpenFile(fmt.Sprintf("%s/%s@%s.json", path, clock.ToRFC3339(), pod.Name), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return
	}
	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return
	}
}
