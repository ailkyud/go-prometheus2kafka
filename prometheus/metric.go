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
	metrics map[string]MetricWithLabel
	//针对指标包含子指标情况,暂不实现
	metrics_with_labels map[string]map[string][]MetricWithLabel
}

//初始化
func NewNodeMetrics() *NodeMetrics {
	return &NodeMetrics{
		metrics:             map[string]MetricWithLabel{},
		metrics_with_labels: map[string]map[string][]MetricWithLabel{},
	}
}

//NodeMetrics 赋值metrics方法
func (nm *NodeMetrics) Add(node_addr, metric string, value float64, labels model.Metric, add_fields MetricWithLabel) {
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
	//插入新增字段
	for k, v := range add_fields {
		label := string(k)
		node_vs[label] = v
	}

}
