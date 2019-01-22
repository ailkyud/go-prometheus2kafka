package prometheus

import (
	"sync"

	"github.com/prometheus/common/model"
)

//指标的数据结构
type InstanceInfo struct {
	//存实例地址,唯一标识
	addr string
	//存prometheus标签
	labels map[string]interface{}
	//存标签扩展
	labels_extent map[string]interface{}
	//存prometheus指标值
	metrics map[string]interface{}
	//存指标值扩展,这里key存对应指标名,value存map
	metrics_extent map[string]map[string]interface{}
}

type InstanceInfoArray struct {
	sync.RWMutex
	arrInstanceInfo []InstanceInfo
}

func NewInstanceInfoArray() *InstanceInfoArray {
	return &InstanceInfoArray{arrInstanceInfo: []InstanceInfo{}}
}

//NodeMetrics 赋值metrics方法
func (instanceInfoArray *InstanceInfoArray) Add(addr, metricName string, metricVal float64, labels model.Metric, metricsExtent map[string]interface{}) {
	instanceInfoArray.Lock()
	defer instanceInfoArray.Unlock()
	//判定是否已存在
	vIndexKey := -1
	for num, vi := range instanceInfoArray.arrInstanceInfo {
		if vi.addr == addr {
			vIndexKey = num
			break
		}
	}
	if vIndexKey > -1 {
		//存在直接赋值或覆盖
		vi := instanceInfoArray.arrInstanceInfo[vIndexKey]
		//插入指标值
		vi.metrics[metricName] = metricVal
		//插入标签
		for k, v := range labels {
			vi.labels[string(k)] = v
		}
		//插入指标值的扩展信息
		vi.metrics_extent[metricName] = map[string]interface{}{}
		for k, v := range metricsExtent {
			vi.metrics_extent[metricName][k] = v
		}

	} else {
		//不存在新增InstanceInfo并插入InstanceInfoArray
		sInstanceInfo := InstanceInfo{}
		sInstanceInfo.addr = addr
		sInstanceInfo.labels = map[string]interface{}{}
		sInstanceInfo.labels_extent = map[string]interface{}{}
		sInstanceInfo.metrics = map[string]interface{}{}
		sInstanceInfo.metrics_extent = map[string]map[string]interface{}{}
		for k, v := range labels {
			sInstanceInfo.labels[string(k)] = v
		}
		sInstanceInfo.metrics[metricName] = metricVal
		sInstanceInfo.metrics_extent[metricName] = metricsExtent
		instanceInfoArray.arrInstanceInfo = append(instanceInfoArray.arrInstanceInfo, sInstanceInfo)
	}

}
