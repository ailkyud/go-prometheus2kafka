package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ailkyud/go-prometheus2kafka/add"
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
	var mq []string
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
				nms.Add(addr, prometheusQuery.Metric, vr.Metric, (float64)(vr.Value))

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

	t1 := startTime.Unix()
	count := 0
	for k, v := range nms.metrics {
		vi := MetricWithLabel{}
		for k2, v2 := range v {
			vi[k2] = v2
		}
		if config.Config.Add_fields.Api_url != "" {
			afs := add.AddFieldsFromExternalApi.GetInstanceAddFields(k)
			vi["addFields"] = afs
		}

		if config.Config.Promql.Instance_id.Is_ip_port {
			vi["addr_ip"] = k[:strings.LastIndex(k, ":")]
			vi["addr_port"] = k[strings.LastIndex(k, ":")+1:]
		} else {
			vi["instance_ip"] = k
			vi["instance_port"] = ""
		}
		//vi["timestamp"] = t1

		local1, err := time.LoadLocation("") //same as "UTC"
		if err != nil {
			fmt.Println(err)
		}
		sTimeProcessed := startTime.In(local1).Format("2006-01-02T15:04:05.000Z")
		vi["@timestamp"] = sTimeProcessed

		jsonBytes, err := json.Marshal(vi)
		if err != nil {
			fmt.Printf("json marshal error, key=%s, value=%v \n", k, vi)
		} else {
			mq = append(mq, string(jsonBytes))
			count++
		}

	}
	now := time.Now()
	local1, err := time.LoadLocation("Asia/Shanghai") //same as "UTC"
	if err != nil {
		fmt.Println(err)
	}
	sTimeProcessed := now.In(local1).Format("2006-01-02 15:04:05")
	processedTime := now.Unix() - t1
	fmt.Printf("提交 %d 条记录至kafka 当前时间 %s (%d 秒 用时)\n", count, sTimeProcessed, processedTime)
	//记录提交kafka送paas
	go kafka.AsyncProducer(config.Config.Kafka.Brokers, config.Config.Kafka.Topicpaas, mq)
	//记录提交kafka送端到端
	//	 kafka.

}
