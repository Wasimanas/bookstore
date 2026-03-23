package mq


import ( 
    "log"
    "store/internal/config"
    amqp "github.com/rabbitmq/amqp091-go"
)


func InitMQ(cfg *config.ApplicationConfig) (*amqp.Connection, error) {
    conn, err := amqp.Dial(cfg.MessageQueueURI)
    if err != nil {
        return nil, err
    }

    log.Println("RabbitMQ connection established")
    return conn, nil
}
