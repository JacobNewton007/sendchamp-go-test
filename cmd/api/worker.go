package main

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Worker() <-chan []byte {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	fmt.Println("Successfully connected to RabbitMq Instance")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"task_jobs", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)

	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global

	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	failOnError(err, "Failed to register a consumer")

	// var forever chan struct{}

	task := make(chan []byte)
	go func() {
		for d := range msgs {
			task <- d.Body
			log.Printf("Received a message: %s", d.Body)
			time.Sleep(1 * time.Second)
			log.Printf("Done")
			d.Ack(false)
		}
		close(task)
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	return task
}
