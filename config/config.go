package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var Config Prometheus2kafkaConfig

const (
	ENCRYPT_KEY = "node2es_encryption_key"
)

type PromQuery struct {
	Metric      string
	Query       string
	Keep_labels bool
}

type Prometheus2kafkaConfig struct {
	Listen_on  string
	Prometheus struct {
		Url string
	}
	Es struct {
		Urls     []string
		Username string
		Password string
		Index    string
	}
	Promql struct {
		Instance_id struct {
			Label, Regex, Replacement string
			Is_ip_port                bool
		}
		Ip_port_label string
		Querys        []PromQuery
	}
	Add_fields struct {
		Api_url string
	}
	Kafka struct {
		Brokers                    []string
		Topicpaas, Topice2e, Group string
	}
}

// LoadConfig loads the specified YAML configuration file.
func LoadConfig(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(content, &Config)
	if Config.Promql.Instance_id.Regex == "" {
		Config.Promql.Instance_id.Regex = "(.*)"
	}
	if Config.Promql.Instance_id.Replacement == "" {
		Config.Promql.Instance_id.Replacement = "$1"
	}
	if err != nil {
		return err
	}

	return nil
}
