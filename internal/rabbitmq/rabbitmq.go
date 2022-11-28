package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn *amqp.Connection
}

type AddTask struct {
	Title     string `json:"title"`
	CreatedBy string `json:"created_by"`
}

func NewMq(connString *amqp.Connection) RabbitMQ {
	return RabbitMQ{
		conn: connString,
	}
}
