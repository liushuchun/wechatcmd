package chat

import (
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
)

type Messenger struct {
	kafkaClient sarama.Client
	producer    sarama.SyncProducer
	consumer    sarama.Consumer
	topic       string
	partition   int32
}

func NewMessenger(brokerURL string, topic string, clientID string) (*Messenger, error) {
	kafkaClient, err := newKafkaClient(brokerURL, clientID)
	if err != nil {
		return nil, err
	}
	producer, err := sarama.NewSyncProducerFromClient(*kafkaClient)
	if err != nil {
		return nil, err
	}
	m := Messenger{
		kafkaClient: *kafkaClient,
		producer:    producer,
		topic:       topic,
		partition:   0,
	}
	return &m, nil
}

func newKafkaClient(brokerURL string, clientID string) (*sarama.Client, error) {
	config := sarama.NewConfig()
	config.ClientID = clientID
	client, err := sarama.NewClient([]string{brokerURL}, config)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (m *Messenger) GetAllMessages() ([]Message, error) {
	maxNumberOfMessages := 10000
	consumer, err := sarama.NewConsumerFromClient(m.kafkaClient)
	if err != nil {
		return nil, err
	}
	pConsumer, err := consumer.ConsumePartition(m.topic, m.partition, sarama.OffsetOldest)
	if err != nil {
		return nil, err
	}
	messages := make([]Message, 0, maxNumberOfMessages)
	timeout := time.Second
	timer := time.NewTimer(500 * time.Millisecond)
	for {
		select {
		case consumerMsg := <-pConsumer.Messages():
			m, err := MessageFromBytes(consumerMsg.Value)
			if err != nil {
				return nil, err
			}
			messages = append(messages, m)
			timer.Reset(timeout)
		case <-timer.C:
			return messages, nil
		}
	}
}

func (m *Messenger) Send(msg Message) error {
	msgBytes, err := msg.ToBytes()
	if err != nil {
		return err
	}
	producerMessage := &sarama.ProducerMessage{Topic: m.topic, Value: sarama.ByteEncoder(msgBytes)}
	_, _, err = m.producer.SendMessage(producerMessage)
	if err != nil {
		return err
	}
	return nil
}

func (m *Messenger) MessageChan() (chan Message, error) {
	msgChan := make(chan Message)

	consumer, err := sarama.NewConsumerFromClient(m.kafkaClient)
	if err != nil {
		return nil, err
	}
	pConsumer, err := consumer.ConsumePartition(m.topic, m.partition, sarama.OffsetNewest)
	if err != nil {
		return nil, err
	}
	go func() {
		for m := range pConsumer.Messages() {
			m, err := MessageFromBytes(m.Value)
			if err != nil {
				log.WithError(err).Error("Fail to read message")
			}
			msgChan <- m
		}
	}()
	return msgChan, nil
}
