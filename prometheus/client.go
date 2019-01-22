package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ailkyud/go-prometheus2kafka/config"
	"github.com/ailkyud/go-prometheus2kafka/kafka"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func LoadMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	//连接Prometheus
	url := config.Config.Prometheus.Url
	client, err := api.NewClient(api.Config{Address: url})
	if err != nil {
		fmt.Printf("Error when trying to connect to Prometheus %s, err: %v \n", url, err)
	}
	query_api := v1.NewAPI(client)

	//送paas存储至es的消息
	vMq := []string{}
	//送端到端的消息
	vMqE2e := []string{}
	//创建一个存放指标数据的对象
	instanceInfoArray := NewInstanceInfoArray()
	//确定实例对应Prometheus的哪个标签
	instance_label := config.Config.Promql.Instance_id.Label
	var wg = sync.WaitGroup{}
	//匿名函数
	getMetric := func(ctx context.Context, prometheusQuery config.PromQuery, queryTime time.Time) {
		defer wg.Done()
		r, err := query_api.Query(ctx, prometheusQuery.Query, queryTime)
		if err != nil {
			fmt.Printf("Prometheus api接口查询失败, err: %v, 查询条件: %s \n", err, prometheusQuery.Query)
		}
		v, ok := r.(model.Vector)
		if ok {
			for _, vr := range v {
				//fmt.Printf("vr.Metric= %s  \n", vr.Metric)
				//fmt.Printf("vr.Value= %s  \n", vr.Value)
				//取实例地址 ip:port
				addr := string(vr.Metric[model.LabelName(instance_label)])
				//送集团的扩展值，从配置文件中读取，这里暂时写死了，有变化需要调整代码
				metricsExtent := map[string]interface{}{}
				metricsExtent["METRICCODE"] = prometheusQuery.Metriccode
				metricsExtent["METRICTYPE"] = prometheusQuery.Metrictype
				instanceInfoArray.Add(addr, prometheusQuery.Metric, (float64)(vr.Value), vr.Metric, metricsExtent)

			}
		}
	}
	//记录开始时间
	startTime := time.Now()
	//开始调用prometheus接口查询
	for _, q := range config.Config.Promql.Querys {
		wg.Add(1)
		go getMetric(ctx, q, startTime)
	}
	wg.Wait()
	//构造送paas-es存储消息
	t1 := startTime.Unix()
	count := 0
	for _, instanceInfo := range instanceInfoArray.arrInstanceInfo {
		//构造一个指标消息对象
		vi := map[string]interface{}{}

		for k, v := range instanceInfo.labels {
			vi[k] = v
		}
		for k, v := range instanceInfo.labels_extent {
			vi[k] = v
		}

		if config.Config.Promql.Instance_id.Is_ip_port {
			vi["addr_ip"] = instanceInfo.addr[:strings.LastIndex(instanceInfo.addr, ":")]
			vi["addr_port"] = instanceInfo.addr[strings.LastIndex(instanceInfo.addr, ":")+1:]
		} else {
			vi["addr_ip"] = instanceInfo.addr
		}

		sTimeProcessed := startTime.Format("2006-01-02T15:04:05.000Z")
		vi["@timestamp"] = sTimeProcessed

		vi["metrics"] = instanceInfo.metrics
		jsonBytes, err := json.Marshal(vi)
		if err != nil {
			fmt.Printf("json marshal error,   value=%v \n", vi)
		} else {
			//fmt.Println(string(jsonBytes))
			vMq = append(vMq, string(jsonBytes))
			count++
		}

	}

	//构造送端到端消息
	for _, instanceInfo := range instanceInfoArray.arrInstanceInfo {
		//构造一个指标消息对象
		vi := map[string]interface{}{}
		vi_sub_arr := []map[string]interface{}{}

		sTimeProcessed := startTime.Format("20060102150405")
		sTimeCollect := startTime.Format("2006-01-02 15:04:05")

		vi["timestamp"] = sTimeProcessed
		vi["name"] = config.Config.Prometheus.Name

		for k1, v1 := range instanceInfo.metrics {
			vi_sub := map[string]interface{}{}
			isSendE2e := false
			for k2, v2 := range instanceInfo.metrics_extent[k1] {
				if k2 == "METRICCODE" || k2 == "METRICTYPE" {
					isSendE2e = true
					vi_sub[k2] = v2
				}
			}
			//如果没有找到扩展信息，说明是本省自己需要的指标，无需送端到端
			if !isSendE2e {
				continue
			}
			vi_sub["METRICVALUE"] = v1
			vi_sub["CompType"] = config.Config.Prometheus.Comptype
			vi_sub["COLLECTTIME"] = sTimeCollect
			vi_sub["HOSTIP"] = instanceInfo.addr[:strings.LastIndex(instanceInfo.addr, ":")]
			vi_sub["CompKey"] = instanceInfo.addr[:strings.LastIndex(instanceInfo.addr, ":")] + "|" + config.Config.Prometheus.Type + "_" + instanceInfo.addr[strings.LastIndex(instanceInfo.addr, ":")+1:]
			vi_sub_arr = append(vi_sub_arr, vi_sub)
		}

		vi["message"] = vi_sub_arr
		jsonBytes, err := json.Marshal(vi)
		if err != nil {
			fmt.Printf("json marshal error,  value=%v \n", vi)
		} else {
			//fmt.Println(string(jsonBytes))
			vMqE2e = append(vMqE2e, string(jsonBytes))
		}

	}

	now := time.Now()
	if err != nil {
		fmt.Println(err)
	}
	processedTime := now.Unix() - t1
	sTimeProcessed := startTime.Format("20060102150405")

	fmt.Printf("提交 %d 条记录至kafka 当前时间 %s (%d 秒 用时)\n", count, sTimeProcessed, processedTime)

	//记录提交kafka送端到端
	go kafka.AsyncProducer(config.Config.Kafka.Brokers, config.Config.Kafka.Topice2e, vMqE2e)
	//记录提交kafka送paas
	go kafka.AsyncProducer(config.Config.Kafka.Brokers, config.Config.Kafka.Topicpaas, vMq)

}
