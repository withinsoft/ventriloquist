package main

import (
	"context"
	"errors"
	"strings"

	"github.com/Xe/ln"
	"github.com/Xe/uuid"
	"github.com/asdine/storm"
	"github.com/withinsoft/ventriloquist/internal/proxytag"
)

type DB struct {
	s *storm.DB
}

type Systemmate struct {
	ID            string `storm:"id"`
	Name          string
	CoreDiscordID string `storm:"index"`
	AvatarURL     string
	Match         proxytag.Match
}

type Webhook struct {
	ID         string `storm:"id"`
	ChannelID  string `storm:"unique"`
	WebhookURL string
}

func (d DB) AddSystemmate(s Systemmate) (Systemmate, error) {
	sms, err := d.FindSystemmates(s.CoreDiscordID)
	if err != nil {
		return Systemmate{}, err
	}

	for _, sm := range sms {
		if strings.EqualFold(s.Name, sm.Name) {
			return Systemmate{}, errors.New("can't add duplicate systemmate")
		}
	}

	s.ID = uuid.New()
	return s, d.s.Save(&s)
}

func (d DB) UpdateSystemmate(s Systemmate) error {
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

func (d DB) DeleteSystemmate(coreDiscordID, name string) error {
	mates, err := d.FindSystemmates(coreDiscordID)
	if err != nil {
		return err
	}

	for _, m := range mates {
		if strings.EqualFold(name, m.Name) {
			return d.s.DeleteStruct(&m)
		}
	}

	return errors.New("database: systemmate not found")
}

func (d DB) FindSystemmateByMatch(coreDiscordID string, m proxytag.Match) (Systemmate, error) {
	sms, err := d.FindSystemmates(coreDiscordID)
	if err != nil {
		return Systemmate{}, err
	}

	for _, sm := range sms {
		if sm.Match.String() == m.String() {
			return sm, nil
		}
	}

	return Systemmate{}, errors.New("database: systemmate not found")
}

func (d DB) NukeSystem(coreDiscordID string) error {
	mates, err := d.FindSystemmates(coreDiscordID)
	if err != nil {
		return err
	}

	var errs []error
	for _, m := range mates {
		if err := d.s.DeleteStruct(&m); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		ctx := ln.WithF(context.Background(), ln.F{
			"core_discord_id": coreDiscordID,
		})

		for _, err := range errs {
			ln.Error(ctx, err)
		}

		return errors.New("error in deletion, contact the bot admin")
	}

	return nil
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
