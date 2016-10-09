package chat

import "encoding/json"

type Message struct {
	User      string `json:"user"`
	Timestamp string `json:"timestamp"`
	Data      string `json:"data"`
}

func MessageFromBytes(bytes []byte) (Message, error) {
	m := Message{}
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return Message{}, err
	}
	return m, nil
}

func (m Message) ToBytes() ([]byte, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
