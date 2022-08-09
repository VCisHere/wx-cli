package storage

import "wx-cli/client"

type Messages []*client.Message

type MsgCache struct {
	messages Messages
}

func NewMsgCache() *MsgCache {
	return &MsgCache{
		messages: make([]*client.Message, 0),
	}
}

func (c *MsgCache) Messages() Messages {
	return c.messages
}

func (c *MsgCache) StoreMessage(msg *client.Message) {
	c.messages = append(c.messages, msg)
}
