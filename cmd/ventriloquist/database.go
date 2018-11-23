package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Xe/ln"
	"github.com/asdine/storm"
	"github.com/gofrs/uuid"
	"github.com/golang/groupcache"
	"github.com/withinsoft/ventriloquist/internal/proxytag"
)

type DB struct {
	s *storm.DB

	systemmateCache *groupcache.Group
}

type Systemmate struct {
	ID            string `storm:"id"`
	Name          string
	CoreDiscordID string `storm:"index"`
	AvatarURL     string
	Match         proxytag.OldMatch //Bad hack to work with the old DB format
	Matchers      []proxytag.Matcher
}

type Webhook struct {
	ID         string `storm:"id"`
	ChannelID  string `storm:"unique"`
	WebhookURL string
}

func (d DB) AddSystemmate(s Systemmate) (Systemmate, error) {
	if len(s.Name) > 15 {
		return Systemmate{}, fmt.Errorf("%s is too long, 15 characters max", s.Name)
	}

	sms, err := d.FindSystemmates(s.CoreDiscordID)
	if err != nil {
		goto skip
	}

	for _, sm := range sms {
		if strings.EqualFold(s.Name, sm.Name) {
			return Systemmate{}, errors.New("can't add duplicate systemmate")
		}
	}

skip:
	id, err := uuid.NewV4()
	if err != nil {
		return Systemmate{}, err
	}

	s.ID = id.String()
	return s, d.s.Save(&s)
}

func (d DB) UpdateSystemmate(s Systemmate) error {
	return d.s.Save(&s)
}

func (d DB) FindSystemmates(id string) ([]Systemmate, error) {
	var bs []byte
	sink := groupcache.AllocatingByteSliceSink(&bs)
	key := filepath.Join(time.Now().Round(time.Second).Format(time.RFC3339), id)

	err := d.systemmateCache.Get(nil, key, sink)
	if err != nil {
		return nil, err
	}

	var result []Systemmate
	err = json.Unmarshal(bs, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d DB) cacheSystemmates(ctx groupcache.Context, key string, dest groupcache.Sink) error {
	did := filepath.Base(key)

	sms, err := d.findSystemmates(did)
	if err != nil {
		return err
	}

	data, err := json.Marshal(&sms)
	if err != nil {
		return err
	}

	dest.SetBytes(data)
	return nil
}

func (d DB) findSystemmates(id string) ([]Systemmate, error) {
	var result []Systemmate
	err := d.s.Find("CoreDiscordID", id, &result)
	if err != nil {
		return nil, err
	}
	for i, sm := range result {
		// The logic here is that if there's no matchers, then this is a legacy
		// account, so we should add the new matchers. However, if there's more than
		// zero, this account has been migrated to the new matchers, so continuing to
		// add the old matchers would just slowly grow the database for no reason.
		if len(sm.Matchers) == 0 {
			// WHAT THE FUCK `range` COPIES.
			// So, `range` *copies* the slice. That means that `sm` is not the same
			// Systemmate in the `result` slice, it's a copy. If we updated
			// `sm.Matchers` here, it wouldn't stick around. That's why we need to
			// update `result[i].Matchers` instead.
			result[i].Matchers = sm.Match.Matchers(sm.Name)
		}
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

func (d DB) FindSystemMatchers(coreDiscordID string) ([]proxytag.Matcher, error) {
	sms, err := d.FindSystemmates(coreDiscordID)
	if err != nil {
		return nil, err
	}

	matchers := make([]proxytag.Matcher, 0)
	for _, sm := range sms {
		matchers = append(matchers, sm.Matchers...)
	}
	return matchers, nil
}

func (d DB) FindSystemmateByMatch(coreDiscordID string, m proxytag.Match) (Systemmate, error) {
	sms, err := d.FindSystemmates(coreDiscordID)
	if err != nil {
		return Systemmate{}, err
	}

	for _, sm := range sms {
		if sm.Name == m.Systemmate {
			return sm, nil
		}
	}

	return Systemmate{}, errors.New("database: systemmate not found")
}

func (d DB) FindSystemmateByMessage(coreDiscordID string, message string) (Systemmate, string, error) {
	matchers, err := d.FindSystemMatchers(coreDiscordID)
	if err != nil {
		return Systemmate{}, "", err
	}

	match, err := proxytag.MatchMessage(message, matchers)
	if err != nil {
		if err.Error() == "error: no match found" {
			return Systemmate{}, "", errors.New("database: systemmate not found")
		} else {
			return Systemmate{}, "", err
		}
	}

	sm, err := d.FindSystemmateByMatch(coreDiscordID, match)
	if err != nil {
		return Systemmate{}, "", err
	}

	return sm, match.Body, nil
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
			ln.Error(ctx, err, ln.F{"to_discord": true})
		}

		return errors.New("error in deletion, contact the bot admin")
	}

	return nil
}

func (d DB) AddWebhook(channelID, whurl string) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	wh := Webhook{
		ID:         id.String(),
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
