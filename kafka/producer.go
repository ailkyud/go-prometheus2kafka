package kafka

import (
	"fmt"

	"github.com/Shopify/sarama"
)

func AsyncProducer(vBrokers []string, vTopic string, vMq []string) {

	config := sarama.NewConfig()
	//等待服务器所有副本都保存成功后的响应
	config.Producer.RequiredAcks = sarama.WaitForAll
	//随机向partition发送消息
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	//是否等待成功和失败后的响应,只有上面的RequireAcks设置不是NoReponse这里才有用.
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	//设置使用的kafka版本,如果低于V0_10_0_0版本,消息中的timestrap没有作用.需要消费和生产同时配置
	//注意，版本设置不对的话，kafka会返回很奇怪的错误，并且无法成功发送消息
	config.Version = sarama.V1_1_0_0

	//fmt.Println("start make producer")
	//使用配置,新建一个异步生产者
	producer, e := sarama.NewAsyncProducer(vBrokers, config)
	if e != nil {
		fmt.Println(e)
		return
	}

	defer producer.AsyncClose()

	var value string

	for i := 0; i < len(vMq); i++ {
		//fmt.Println(i)
		value = vMq[i]
		// 发送的消息,主题。
		// 注意：这里的msg必须得是新构建的变量，不然你会发现发送过去的消息内容都是一样的，因为批次发送消息的关系。
		msg := &sarama.ProducerMessage{
			Topic: vTopic,
		}

		//将字符串转化为字节数组
		msg.Value = sarama.ByteEncoder(value)

		//fmt.Println(value)

		//使用通道发送
		producer.Input() <- msg

		// select {
		// //case suc := <-producer.Successes():
		// //fmt.Printf("offset: %d,  timestamp: %s", suc.Offset, suc.Timestamp.String())
		// case fail := <-producer.Errors():
		// 	fmt.Printf("err: %s\n", fail.Err.Error())
		// }
	}
}
