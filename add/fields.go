package add

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/hongxincn/promexp/node2es/config"
)

var AddFieldsFromExternalApi *AddFields

type AddFields struct {
	IntanceAddfields map[string]map[string]string
	LoadTime         int64
	api_url          string
	needReload       bool
	sync.Mutex
}

func NewAddFields() *AddFields {
	AddFieldsFromExternalApi = &AddFields{
		api_url: config.Config.Add_fields.Api_url,
	}
	AddFieldsFromExternalApi.reload()
	go AddFieldsFromExternalApi.checkIfNeedReload()
	return AddFieldsFromExternalApi
}

func (af *AddFields) checkIfNeedReload() {
	c := time.Tick(time.Duration(60) * time.Second)
	for {
		<-c
		if af.needReload || time.Now().Unix()-af.LoadTime >= 3600*24 {
			af.reload()
		}
	}
}

func (af *AddFields) SetReloadFlag() {
	af.needReload = true
}

func (af *AddFields) GetInstancesMapping() []byte {
	v, err := json.Marshal(af)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}
	return v
}

func (af *AddFields) GetInstanceAddFields(instance_id string) map[string]string {

	result := map[string]string{}
	if af.api_url == "" {
		return result
	}
	af.Lock()
	defer af.Unlock()
	value, ok := af.IntanceAddfields[instance_id]
	if ok {
		result = value
	}
	return result
}

func (af *AddFields) reload() {
	iaf := af.getAddFieldsData()
	af.Lock()
	af.IntanceAddfields = iaf
	af.LoadTime = time.Now().Unix()
	af.needReload = false
	af.Unlock()
}

func (af *AddFields) getAddFieldsData() map[string]map[string]string {
	iaf := map[string]map[string]string{}
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := httpClient.Get(af.api_url)
	if err != nil {
		fmt.Printf("Error when trying to connect to %s, err: %v \n", af.api_url, err)
		return iaf
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error when reading the result %v\n", resp, err)
		return iaf
	}

	var result []map[string]string
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Printf("Error when unmarshal the result body %s\n", body, err)
		return iaf
	}

	for _, rec := range result {
		iid := rec["instance_id"]
		delete(rec, "instance_id")
		iaf[iid] = rec
	}
	return iaf
}
