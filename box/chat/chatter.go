package chat

import (
	"time"

	log "github.com/Sirupsen/logrus"
)

type Chatter struct {
	room      string
	username  string
	messenger *Messenger
}
type Config struct {
	Room      string
	Username  string
	BrokerURL string
	ClientID  string
}

func New(config *Config) (*Chatter, error) {
	messenger, err := NewMessenger(config.BrokerURL, config.Room, config.ClientID)
	if err != nil {
		return nil, err
	}
	chatter := Chatter{
		room:      config.Room,
		username:  config.Username,
		messenger: messenger,
	}
	return &chatter, nil
}

func (b *Chatter) Start() (msgIn chan Message, textOut chan string) {
	maxChanSize := 10000
	/*
		allMessages, err := b.messenger.GetAllMessages()
		if err != nil {
			log.WithError(err).Error("Fail to get all messages")
		}
		log.WithField("count", len(allMessages)).Debug("Got all messages.")
	*/

	msgIn = make(chan Message, maxChanSize)
	textOut = make(chan string, maxChanSize)
	/*
		for _, m := range allMessages {
			msgIn <- m
		}*/
	/*b.startListening(msgIn)
	b.startProducing(textOut)
	*/
	return msgIn, textOut
}

func (b *Chatter) startProducing(textOutChan chan string) {

	go func() {
		for text := range textOutChan {
			m := Message{User: b.username, Timestamp: time.Now().Format("Mon Jan _2 15:04:05"), Data: text}
			err := b.messenger.Send(m)
			if err != nil {
				log.WithError(err).Error("Fail to send message")
			}
			log.Debug("Message sent")
		}
	}()
}

func (b *Chatter) startListening(msgInChan chan Message) {
	go func() {
		messengerChan, err := b.messenger.MessageChan()
		if err != nil {
			log.WithError(err).Error("Fail to create messenger message chan")
		}
		log.Debug("Start listening")
		for m := range messengerChan {
			msgInChan <- m
		}
	}()
}
