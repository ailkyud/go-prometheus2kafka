package prometheus

import (
	"sync"

	"github.com/prometheus/common/model"
)

//metric中一个dataPoint的数据结构
type DataPoint struct {
	//存实例地址,唯一标识
	addr string
	//存prometheus标签
	tags map[string]interface{}
	//存标签扩展
	tags_extent map[string]interface{}
	//存prometheus度量域
	metricField map[string]interface{}
	//存度量域扩展,这里key存对应度量名,value存map
	metricField_extent map[string]map[string]interface{}
}

//metric数据
type DataPointArray struct {
	sync.RWMutex
	arrDataPoint []DataPoint
}

//初始化一个metric数据
func NewDataPointArray() *DataPointArray {
	return &DataPointArray{arrDataPoint: []DataPoint{}}
}

//NodeMetrics 赋值metrics方法
func (dataPointArray *DataPointArray) Add(addr, metricFieldName string, metricFieldVal float64, tags model.Metric, metricFieldExtent map[string]interface{}) {
	dataPointArray.Lock()
	defer dataPointArray.Unlock()
	//判定是否已存在
	vIndexKey := -1
	for num, vi := range dataPointArray.arrDataPoint {
		if vi.addr == addr {
			vIndexKey = num
			break
		}
	}
	if vIndexKey > -1 {
		//存在直接赋值或覆盖
		vi := dataPointArray.arrDataPoint[vIndexKey]
		//插入指标值
		vi.metricField[metricFieldName] = metricFieldVal
		//插入标签
		for k, v := range tags {
			vi.tags[string(k)] = v
		}
		//插入指标值的扩展信息
		vi.metricField_extent[metricFieldName] = map[string]interface{}{}
		for k, v := range metricFieldExtent {
			vi.metricField_extent[metricFieldName][k] = v
		}

	} else {
		//不存在新增InstanceInfo并插入InstanceInfoArray
		dataPoint := DataPoint{}
		dataPoint.addr = addr
		dataPoint.tags = map[string]interface{}{}
		dataPoint.tags_extent = map[string]interface{}{}
		dataPoint.metricField = map[string]interface{}{}
		dataPoint.metricField_extent = map[string]map[string]interface{}{}
		for k, v := range tags {
			dataPoint.tags[string(k)] = v
		}
		dataPoint.metricField[metricFieldName] = metricFieldVal
		dataPoint.metricField_extent[metricFieldName] = metricFieldExtent
		dataPointArray.arrDataPoint = append(dataPointArray.arrDataPoint, dataPoint)
	}

}
