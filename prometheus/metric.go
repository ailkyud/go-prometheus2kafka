package prometheus

import (
	//	"math"
	//"strings"
	"sync"

	"github.com/prometheus/common/model"
)

//指标标签
type MetricWithLabel map[string]interface{}

//生成的数据结构NodeMetrics
type NodeMetrics struct {
	sync.RWMutex
	//针对指标包含子指标情况,暂不实现
	metrics_with_labels map[string]map[string][]MetricWithLabel
	metrics             map[string]MetricWithLabel
}

//初始化
func NewNodeMetrics() *NodeMetrics {
	return &NodeMetrics{
		metrics_with_labels: map[string]map[string][]MetricWithLabel{},
		metrics:             map[string]MetricWithLabel{},
	}
}

//NodeMetrics 赋值metrics方法
func (nm *NodeMetrics) Add(node_addr, metric string, labels model.Metric, value float64) {
	nm.Lock()
	defer nm.Unlock()
	node_vs, ok := nm.metrics[node_addr]
	if !ok {
		node_vs = MetricWithLabel{}
		nm.metrics[node_addr] = node_vs
	}
	//插入指标
	node_vs[metric] = value
	//插入标签
	for k, v := range labels {
		label := string(k)
		node_vs[label] = v
	}

}
