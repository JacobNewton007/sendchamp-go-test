package rabbitmq

import (
	// "bytes"

	// "time"

	// "encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	// "github.com/JacobNewton007/sendchamp-go-test/internal/data"
	// "time"
)

func (q RabbitMQ) Worker() AddTask {

	fmt.Println("Successfully connected to RabbitMq Instance")

	// ch, err := q.conn.Channel()
	amqpChannel, err := q.conn.Channel()
	failOnError(err, "Can't create a amqpChannel")

	defer amqpChannel.Close()

	queue, err := amqpChannel.QueueDeclare("add", true, false, false, false, nil)
	failOnError(err, "Could not declare `add` queue")

	err = amqpChannel.Qos(1, 0, false)
	failOnError(err, "Could not configure QoS")

	messageChannel, err := amqpChannel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Could not register consumer")

	// forever := make(chan int)
	task := make(chan string)
	var addT AddTask
	log.Printf(" [x] preparing %+v\n", addT)
	go func() {
		// log.Printf("Consumer ready, PID: %d", os.Getpid())
		for d := range messageChannel {
			log.Printf("Received a message: %s", d.Body)
			addTask := &AddTask{}

			err := json.Unmarshal(d.Body, addTask)
			addT = *addTask
			task <- "done"
			if err != nil {
				log.Printf("Error decoding JSON: %s", err)
			}
			log.Printf("AddTask:%s\n %s", addTask.Title, addTask.CreatedBy)
			if err := d.Ack(false); err != nil {
				log.Printf("Error acknowledging message : %s", err)
			} else {
				log.Printf("Acknowledged message")
			}
		}
		// d := <-messageChannel
		// log.Printf("Received a message: %s", d.Body)

		// addTask := &AddTask{}
		// err := json.Unmarshal(d.Body, addTask)
		// addT = *addTask
		// if err != nil {
		// 	log.Printf("Error decoding JSON: %s", err)
		// }
		// log.Printf("AddTask:%s\n %s", addT.Title, addT.CreatedBy)
		// if err := d.Ack(false); err != nil {
		// 	log.Printf("Error acknowledging message : %s", err)
		// } else {
		// 	log.Printf("Acknowledged message")
		// }

		defer close(task)
	}()
	// Stop for program termination
	// <-forever
	<-task
	log.Printf(" [x] Received %+v\n", addT)
	// return addT
	return addT
}
