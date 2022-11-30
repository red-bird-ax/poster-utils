package queue

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    amqp "github.com/rabbitmq/amqp091-go"
)

// todo: learn func`s arguments

type Callback func(message *amqp.Delivery) error

type Config struct {
    Name         string
    User         string
    Password     string
    Host         string
    Port         int
}

type MessageQueue struct {
    connection   *amqp.Connection
    channel      *amqp.Channel
    queue        *amqp.Queue
    callbacks    map[string]Callback
    name         string
}

func New(config *Config) (*MessageQueue, error) {
    url := fmt.Sprintf(
        "amqp://%s:%s@%s:%d",
        config.User,
        config.Password,
        config.Host,
        config.Port,
    )

    connection, err := amqp.Dial(url)
    if err != nil {
        return nil, err
    }

    channel, err := connection.Channel()
    if err != nil {
        return nil, err
    }

    mq := MessageQueue{
        connection:   connection,
        channel:      channel,
        name:         config.Name,
        callbacks:    make(map[string]Callback),
    }

    if err = mq.declareExchange(fmt.Sprintf("%s-exchange", config.Name)); err != nil {
        return nil, err
    }

    if err = mq.declareQueue(fmt.Sprintf("%s-queue", config.Name)); err != nil {
        return nil, err
    }

    return &mq, nil
}

func (mq *MessageQueue) declareExchange(exchange string) error {
    return mq.channel.ExchangeDeclare(
        exchange,
        "direct", // exchange type
        true,     // durable
        false,    // auto-deleted
        false,    // internal
        false,    // no-wait
        nil,      // args
    )
}

func (mq *MessageQueue) declareQueue(queue string) error {
    q, err := mq.channel.QueueDeclare(
        queue,
        false, // durable
        false, // delete when unused
        true,  // exclusive?
        false, // no-wait
        nil,   // args
    )
    if err != nil {
        return err
    }

    mq.queue = &q
    return nil
}

func (mq *MessageQueue) PublishJSON(key string, message any) error {
    bytes, err := json.Marshal(message)
    if err != nil {
        return err
    }

    err = mq.channel.PublishWithContext(
        context.Background(), // todo: learn context
        fmt.Sprintf("%s-exchange", mq.name),
        key,
        false, // mandatory
        false, // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        bytes,
        },
    )
    return err
}

func (mq *MessageQueue) Subscribe(key, exchange string, callback Callback) error {
    if err := mq.declareExchange(exchange); err != nil {
        return err
    }

    err := mq.channel.QueueBind(
        mq.queue.Name,
        key,
        exchange,
        false, // no-wait
        nil,   // args
    )
    if err != nil {
        return err
    }

    mq.callbacks[key] = callback
    return nil
}

func (mq *MessageQueue) Listen() error {
    messages, err := mq.channel.Consume(
        mq.queue.Name,
        "",    // consumer
        true,  // auto-ack
        false, // exclusive
        false, // no-local
        false, // no-wait
        nil,   // args
    )
    if err != nil {
        return err
    }

    for message := range messages {
        key := message.RoutingKey

        if callback, ok := mq.callbacks[key]; ok {
            go func(message *amqp.Delivery, callback Callback) {
                if err := callback(message); err != nil {
                    _, _ = fmt.Fprintf(os.Stderr, "[Message Queue] %s: %s\n", key, err.Error())
                }
            }(&message, callback)
        } else {
            _, _ = fmt.Fprintf(os.Stderr, "[Message Queue] Callback for '%s' key not found\n", key)
        }
    }
    return nil
}

func UnmarshalJSON[T any](message *amqp.Delivery) (*T, error) {
    var body T
    if err := json.Unmarshal(message.Body, &body); err != nil {
        return nil, err
    }
    return &body, nil
}