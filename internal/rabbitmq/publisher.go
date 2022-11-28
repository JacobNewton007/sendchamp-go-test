package rabbitmq

import (
	// "bytes"
	"context"
	"math/rand"

	// "encoding/gob"
	"encoding/json"

	// "encoding/json"
	"fmt"
	"log"
	"time"

	// "github.com/JacobNewton007/sendchamp-go-test/internal/data"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func (q RabbitMQ) Publisher(input AddTask) {
	fmt.Println("Successfully connected to RabbitMq Instance")
	amqpChannel, err := q.conn.Channel()
	failOnError(err, "Can't create a amqpChannel")

	defer amqpChannel.Close()

	queue, err := amqpChannel.QueueDeclare("add", true, false, false, false, nil)
	failOnError(err, "Could not declare `add` queue")

	rand.Seed(time.Now().UnixNano())

	addTask := AddTask{Title: input.Title, CreatedBy: input.CreatedBy}
	body, err := json.Marshal(addTask)
	if err != nil {
		failOnError(err, "Error encoding JSON")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = amqpChannel.PublishWithContext(ctx, "", queue.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         body,
	})

	if err != nil {
		log.Fatalf("Error publishing message: %s", err)
	}

	log.Printf("AddTask: %s\n %s", addTask.Title, addTask.CreatedBy)
	log.Printf(" [x] Sent %+v\n", string(body))
}
