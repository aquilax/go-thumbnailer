package tasks

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type RabbitMQBackend struct {
	tasks *amqp.Channel
	queue *amqp.Queue
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func get_connection() (*amqp.Channel, *amqp.Queue) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	// defer ch.Close()  TODO: Don't forget to close the channel

	q, err := ch.QueueDeclare(
		"tasks", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")
	return ch, &q
}

func (mb *RabbitMQBackend) Get() *Task {
	msgs, err := mb.tasks.Consume(
		mb.queue.Name, // queue
		"",            // consumer
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	failOnError(err, "Failed to register a consumer")

	var t *Task
	msg := <-msgs
	log.Printf("Received a message: %s", msg.Body)
	err = json.Unmarshal(msg.Body, &t)
	failOnError(err, "Failed to unmarshal data")
	return t

}

func (mb *RabbitMQBackend) Put(t *Task) {
	data, err := json.Marshal(*t)
	failOnError(err, "cannot marshal data")
	err = mb.tasks.Publish(
		"",            // exchange
		mb.queue.Name, // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		})
	failOnError(err, "Failed to publish a message")
	return
}
