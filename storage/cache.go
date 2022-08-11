package storage

import (
	"encoding/json"
	"io/ioutil"
	"wx-cli/client"
)

type Messages []*client.Message

type Cache struct {
	messages   Messages
	viewMsgCur int
	fileName   string
}

func NewCache(fileName string) *Cache {
	return &Cache{
		messages: make([]*client.Message, 0),
		fileName: fileName,
	}
}

func NewCacheFromFile(fileName string) (*Cache, error) {
	c := &Cache{}
	var err error
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}
	c.fileName = fileName
	return c, nil
}

func (c *Cache) StoreMessage(msg *client.Message) {
	c.messages = append(c.messages, msg)
}

func (c *Cache) AllMessages() Messages {
	return c.messages
}

func (c *Cache) UnreadMessages() Messages {
	cur := c.viewMsgCur
	c.viewMsgCur = len(c.messages)
	return c.messages[cur:]
}
