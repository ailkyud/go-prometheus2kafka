package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ailkyud/go-prometheus2kafka/add"
	"github.com/ailkyud/go-prometheus2kafka/config"
	"github.com/ailkyud/go-prometheus2kafka/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	VERSION    string
	BUILD_TIME string
	GO_VERSION string
)

func main() {
	var configFile = kingpin.Flag("config.file", "Configuration file").Default("D:/go_code/prometheus2kafka/go-prometheus2kafka/prometheus2kafka.yml").String()
	var interval = kingpin.Flag("interval", "Interval time (Unit second)").Default("60").Int()
	kingpin.HelpFlag.Short('h')
	kingpin.Version(fmt.Sprintf("%s\n%s\n%s", VERSION, BUILD_TIME, GO_VERSION))
	kingpin.Parse()

	err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//刷新缓存
	if config.Config.Add_fields.Api_url != "" {
		addFieldsEndPoint := add.NewAddFields()
		http.HandleFunc("/-/reload", func(w http.ResponseWriter, r *http.Request) {
			addFieldsEndPoint.SetReloadFlag()
			w.Write([]byte(`{"acknowledged":"true"}`))
		})
		http.HandleFunc("/-/instancesMapping", func(w http.ResponseWriter, r *http.Request) {
			w.Write(addFieldsEndPoint.GetInstancesMapping())
		})
	}

	//处理指标
	c := time.Tick(time.Duration(*interval) * time.Second)
	go func() {
		for {
			prometheus.LoadMetrics()
			<-c
		}
	}()
	fmt.Println("监听端口", config.Config.Listen_on)
	err = http.ListenAndServe(config.Config.Listen_on, nil)
	if err != nil {
		fmt.Println(err)
	}
}
