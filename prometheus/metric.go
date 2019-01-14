package prometheus

import (
	"math"
	"strings"
	"sync"

	"github.com/prometheus/common/model"
)

type MetricWithLabel map[string]interface{}

type NodeMetrics struct {
	sync.RWMutex
	metrics             map[string]map[string]float64
	metrics_with_labels map[string]map[string][]MetricWithLabel
}

func NewNodeMetrics() *NodeMetrics {
	return &NodeMetrics{
		metrics:             map[string]map[string]float64{}, //key: instance_id
		metrics_with_labels: map[string]map[string][]MetricWithLabel{},
	}
}

func (nm *NodeMetrics) Add(node_addr, metric string, labels model.Metric, value float64, remove_label string) {
	nm.Lock()
	defer nm.Unlock()
	if math.IsNaN(value) {
		value = -1
	}
	node_vs, ok := nm.metrics_with_labels[node_addr]
	if !ok {
		node_vs = map[string][]MetricWithLabel{}
		nm.metrics_with_labels[node_addr] = node_vs
	}
	v2, ok := node_vs[metric]
	if !ok {
		v2 = make([]MetricWithLabel, 0)
		node_vs[metric] = v2
	}
	delete(labels, model.LabelName(remove_label))

	mvl := MetricWithLabel{}
	for k, v := range labels {
		label := string(k)
		label_c := strings.ToUpper(label[:1]) + label[1:]
		mvl[label_c] = v
	}
	mvl["Value"] = round(value, 2)
	//	mvl := MetricWithLabel{
	//		Labels: labels,
	//		Value:  round(value, 2),
	//	}
	v2 = append(v2, mvl)
	node_vs[metric] = v2
}

func (nm *NodeMetrics) Set(node_addr, metric string, value float64) {
	nm.Lock()
	defer nm.Unlock()
	node_vs, ok := nm.metrics[node_addr]
	if !ok {
		node_vs = make(map[string]float64)
		nm.metrics[node_addr] = node_vs
	}
	if math.IsNaN(value) {
		value = -1
	}
	node_vs[metric] = round(value, 2)
}

func round(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10
}

func (nm *NodeMetrics) Get(node_addr, metric string) (float64, bool) {
	nm.RLock()
	defer nm.RUnlock()
	node_vs, ok := nm.metrics[node_addr]
	if ok {
		v, ok := node_vs[metric]
		if ok {
			return v, true
		}
	}
	return 0, false
}
