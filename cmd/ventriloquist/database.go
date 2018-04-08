package main

import (
	"github.com/Xe/uuid"
	"github.com/asdine/storm"
)

type DB struct {
	s *storm.DB
}

type Systemmate struct {
	ID            string `storm:"id"`
	Name          string `storm:"index"`
	CoreDiscordID string `storm:"index"`
	AvatarURL     string
}

type Webhook struct {
	ID         string `storm:"id"`
	ChannelID  string `storm:"unique"`
	WebhookURL string
}

func (d DB) AddSystemmate(s Systemmate) error {
	s.ID = uuid.New()
	return d.s.Save(&s)
}

func (d DB) FindSystemmates(id string) ([]Systemmate, error) {
	var result []Systemmate
	err := d.s.Find("CoreDiscordID", id, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d DB) AddWebhook(channelID, whurl string) error {
	wh := Webhook{
		ID:         uuid.New(),
		ChannelID:  channelID,
		WebhookURL: whurl,
	}

	return d.s.Save(&wh)
}

func (d DB) FindWebhook(channelID string) (string, error) {
	var result Webhook
	err := d.s.One("ChannelID", channelID, &result)
	if err != nil {
		return "", err
	}
	return result.WebhookURL, nil
}
