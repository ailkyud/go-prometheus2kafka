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
	url := config.Config.Prometheus.Url
	client, err := api.NewClient(api.Config{Address: url})
	if err != nil {
		fmt.Printf("Error when trying to connect to Prometheus %s, err: %v \n", url, err)
	}
	vMq := []string{}
	vMqE2e := []string{}
	query_api := v1.NewAPI(client)
	nms := NewNodeMetrics()
	instance_label := config.Config.Promql.Instance_id.Label
	var wg = sync.WaitGroup{}
	//匿名函数
	getMetric := func(ctx context.Context, prometheusQuery config.PromQuery, queryTime time.Time) {
		defer wg.Done()
		r, err := query_api.Query(ctx, prometheusQuery.Query, queryTime)
		if err != nil {
			fmt.Printf("Prometheus Error occuried while querying, err: %v, query: %s \n", err, prometheusQuery.Query)
		}
		v, ok := r.(model.Vector)
		if ok {
			for _, vr := range v {
				//fmt.Println(vr.Metric)
				//fmt.Println(vr.Value)
				//这里paas保留实例完整信息 ip:port
				addr := string(vr.Metric[model.LabelName(instance_label)])
				vAddFields := MetricWithLabel{}
				vAddFields["METRICCODE"] = prometheusQuery.Metriccode
				vAddFields["METRICTYPE"] = prometheusQuery.Metrictype
				nms.Add(addr, prometheusQuery.Metric, (float64)(vr.Value), vr.Metric, vAddFields)

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
	for k, v := range nms.metrics {
		vi := MetricWithLabel{}
		for k2, v2 := range v {
			vi[k2] = v2
		}

		if config.Config.Promql.Instance_id.Is_ip_port {
			vi["addr_ip"] = k[:strings.LastIndex(k, ":")]
			vi["addr_port"] = k[strings.LastIndex(k, ":")+1:]
		} else {
			vi["addr_ip"] = k
			vi["addr_port"] = ""
		}

		if err != nil {
			fmt.Println(err)
		}
		sTimeProcessed := startTime.Format("2006-01-02T15:04:05.000Z")
		vi["@timestamp"] = sTimeProcessed

		vi["metrics"] = nms.metrics_with_labels[k]["metrics"]
		jsonBytes, err := json.Marshal(vi)
		if err != nil {
			fmt.Printf("json marshal error, key=%s, value=%v \n", k, vi)
		} else {
			//fmt.Println(string(jsonBytes))
			vMq = append(vMq, string(jsonBytes))
			count++
		}

	}
	//构造送端到端消息
	for k, v := range nms.metrics_with_labels {
		vi := MetricWithLabel{}
		vi_sub_arr := []MetricWithLabel{}

		if err != nil {
			fmt.Println(err)
		}
		sTimeProcessed := startTime.Format("20060102150405")
		sTimeCollect := startTime.Format("2006-01-02 15:04:05")

		vi["timestamp"] = sTimeProcessed
		vi["name"] = config.Config.Prometheus.Name

		for _, v2 := range v["metrics"] {
			vi_sub := MetricWithLabel{}

			for k3, v3 := range v2 {
				if k3 == "METRICCODE" || k3 == "METRICTYPE" {
					vi_sub[k3] = v3
				}
				if strings.Contains(k3, "metric@") {
					vi_sub["METRICVALUE"] = v3
				}
			}

			vi_sub["CompType"] = config.Config.Prometheus.Comptype
			vi_sub["COLLECTTIME"] = sTimeCollect
			vi_sub["HOSTIP"] = k[:strings.LastIndex(k, ":")]
			vi_sub["CompKey"] = k[:strings.LastIndex(k, ":")] + "|redis_" + k[strings.LastIndex(k, ":")+1:]
			vi_sub_arr = append(vi_sub_arr, vi_sub)
		}

		vi["message"] = vi_sub_arr
		jsonBytes, err := json.Marshal(vi)
		if err != nil {
			fmt.Printf("json marshal error, key=%s, value=%v \n", k, vi)
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
	//go kafka.AsyncProducer(config.Config.Kafka.Brokers, config.Config.Kafka.Topicpaas, vMq)

}
