package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
}

func (rabbit RabbitMQ) NewRabbitMq(connString string) *amqp.Connection {
	conn, err := amqp.Dial(connString)
	if err != nil {
		panic(err)
	}
	return conn
}
