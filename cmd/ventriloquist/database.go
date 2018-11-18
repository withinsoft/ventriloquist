package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
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
	Match         proxytag.Match
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

func (d DB) FindSystemmateByMessage(coreDiscordID string, message string) (Systemmate, string, error) {
	sms, err := d.FindSystemmates(coreDiscordID)
	if err != nil {
		return Systemmate{}, "", err
	}

	matchers := make([]map[string]interface{}, len(sms))

	for i, sm := range sms {
		match := sm.Match
		var prefix, suffix string
		if match.Method == "Nameslash" {
			prefix = match.Name + "\\"
		} else if match.Method == "Sigils" {
			prefix = match.InitialSigil
			suffix = match.EndSigil
		} else if match.Method == "HalfSigilStart" {
			prefix = match.InitialSigil
		} else {
			// Dirty hack to create a matcher that won't match anything
			prefix = "\x00"
			suffix = "\x00"
		}
		if suffix == "" {
			matchers[i] = map[string]interface{}{
				"matcherPrefix": prefix,
				"matcherSuffix": suffix,
				"matcherSystemMate": sm.Name,
			}
		} else {
			matchers[i] = map[string]interface{}{
				"matcherPrefix": prefix,
				"matcherSystemMate": sm.Name,
			}

		}
	}

	matcherMessage := map[string]interface{}{
		"messageBody": message,
		"messageMatchers": matchers,
	}

	cmd := exec.Command("proxy-matcher")
	cmdStdin, err := cmd.StdinPipe()
	if err != nil {
		return Systemmate{}, "", err
	}
	err = json.NewEncoder(cmdStdin).Encode(matcherMessage)
	if err != nil {
		return Systemmate{}, "", err
	}
	cmdStdin.Close()
	cmdStdout, err := cmd.Output()
	if err != nil {
		return Systemmate{}, "", err
	}
	var matcherResponse map[string]interface{}
	err = json.Unmarshal(cmdStdout, &matcherResponse)
	if err != nil {
		return Systemmate{}, "", err
	}

	if matcherResponse["responseError"] != nil || matcherResponse["responseMatch"] == nil {
		return Systemmate{}, "", errors.New("database: systemmate not found")
	}

	match := matcherResponse["responseMatch"].(map[string]interface{})
	matchSystemMate := match["matchSystemMate"].(string)
	matchBody := match["matchBody"].(string)

	for _, sm := range sms {
		if sm.Name == matchSystemMate {
			return sm, matchBody, nil
		}
	}

	return Systemmate{}, "", errors.New("database: systemmate not found")
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
