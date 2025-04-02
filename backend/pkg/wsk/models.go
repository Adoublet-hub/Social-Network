package wsk

import "backend/pkg/models"

type userChannel chan *UserChat
type messageChannel chan *models.Message

type Channel struct {
	messageChannel messageChannel
	leaveChannel   userChannel
}
