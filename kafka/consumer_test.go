package main

import "fmt"

func main() {
	consumer, err := kafka.NewConsumer("group", "testNJ", "localhost:2181")
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Kafka consumer: %v", err))
	}

	err = consumer.Start()
	if err != nil {
		panic(fmt.Sprintf("Failed to start Kafka consumer: %v", err))
	}
}
