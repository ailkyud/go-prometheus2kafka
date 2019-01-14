package es

import (
	"context"
	"errors"
	"fmt"

	"github.com/ailkyud/go-prometheus2kafka/config"
	"gopkg.in/olivere/elastic.v5"
)

var Client *EsClient

type EsClient struct {
	client *elastic.Client
	//	bs     *elastic.BulkService
}

func NewEsClient() error {
	var esclient *elastic.Client
	configuration := config.Config
	if len(configuration.Es.Urls) == 0 {
		return errors.New("no es configuration")
	}
	var err error
	decyptedPassword, err := config.GetDecryptedPassword(configuration.Es.Password)
	if err != nil {
		return err
	}
	if decyptedPassword != "" {
		esclient, err = elastic.NewClient(elastic.SetURL(configuration.Es.Urls...),
			elastic.SetBasicAuth(configuration.Es.Username, decyptedPassword), elastic.SetSniff(false))
		if err != nil {
			return err
		}
	} else {
		esclient, err = elastic.NewClient(elastic.SetURL(configuration.Es.Urls...), elastic.SetSniff(false))
		if err != nil {
			return err
		}
	}
	Client = &EsClient{client: esclient}
	return nil
}

func (ec *EsClient) NewBulkService() *elastic.BulkService {
	return elastic.NewBulkService(ec.client)
}

func (ec *EsClient) AddBulkRequest(bs *elastic.BulkService, index, msg string) {
	if bs == nil {
		return
	}

	bir := elastic.NewBulkIndexRequest().Index(index).Type("doc").Doc(
		string(msg))

	bs.Add(bir)
}

func (ec *EsClient) SubmitBulkRequest(bs *elastic.BulkService) {
	ctx := context.Background()
	_, err := bs.Do(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
