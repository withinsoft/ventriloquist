package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Xe/ln"
	"github.com/bwmarrin/discordgo"
)

type bot struct {
	cfg config
	db  DB
	dg  *discordgo.Session
}

func (b bot) addSystemmate(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	if len(parv) != 3 {
		return errors.New("usage: ;add <name> <avatar url>\n\n(don't include the angle brackets)")
	}

	name := parv[1]
	aurl := parv[2]
	_, err := url.Parse(aurl)
	if err != nil {
		return fmt.Errorf("can't parse avatar url: %v", err)
	}

	sm := Systemmate{
		CoreDiscordID: m.Author.ID,
		Name:          name,
		AvatarURL:     aurl,
	}

	ln.Log(context.Background(), ln.Action("adding systemmate"), ln.F{
		"name":       name,
		"avatar_url": aurl,
	})

	err = b.db.AddSystemmate(sm)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Added member "+sm.Name)
	return err
}

func (b bot) updateAvatar(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	return errors.New("not implemented")
}

func (b bot) listSystemmates(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	members, err := b.db.FindSystemmates(m.Author.ID)
	if err != nil {
		return err
	}

	sb := strings.Builder{}
	sb.WriteString("members:\n")
	for i, m := range members {
		sb.WriteString(fmt.Sprintf("%d. %s - <%s>\n", (i + 1), m.Name, m.AvatarURL))
	}

	_, err = s.ChannelMessageSend(m.ChannelID, sb.String())
	return err
}

func (b bot) delSystemmate(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	return errors.New("not implemented")
}

func (b bot) proxyScrape(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msg := m.Content
	fl := strings.Fields(msg)
	f0 := fl[0]
	f := ln.F{
		"author_id":       m.Author.ID,
		"author_username": m.Author.Username + "#" + m.Author.Discriminator,
	}

	// Proxy tags are defined as the following:
	//   Foo\ bar
	// Is a message by "Foo" saying "bar".
	if !strings.Contains(f0, "\\") {
		return
	}

	name := f0[:len(f0)-1]
	f["name"] = name

	members, err := b.db.FindSystemmates(m.Author.ID)
	if err != nil {
		ln.Error(context.Background(), err, f, ln.Action("finding systemmate"))
		return
	}

	var member Systemmate
	for _, m := range members {
		if strings.EqualFold(name, m.Name) {
			member = m
		}
	}

	if member.Name == "" {
		return
	}
	f["member_id"] = member.ID

	wh, err := b.db.FindWebhook(m.ChannelID)
	if err != nil {
		if err.Error() == "not found" {
			whook, err := s.WebhookCreate(m.ChannelID, "ventriloquist proxy bot", "https://i.ytimg.com/vi/EzOeUXVDjSM/hqdefault.jpg")
			if err != nil {
				ln.Error(ctx, err, f, ln.Action("creating webhook"))
				return
			}

			wh = "https://discordapp.com/api/webhooks/" + whook.ID + "/" + whook.Token
			err = b.db.AddWebhook(m.ChannelID, wh)
			if err != nil {
				ln.Error(ctx, err, f, ln.Action("adding webhook to database"))
				return
			}
		} else {
			ln.Error(ctx, err, f, ln.Action("finding webhook"))
			return
		}
	}

	dw := dWebhook{
		Content:   strings.Join(fl[1:], " "),
		Username:  fmt.Sprintf("%s of %s#%s", member.Name, m.Author.Username, m.Author.Discriminator),
		AvatarURL: member.AvatarURL,
	}

	err = sendWebhook(wh, dw)
	if err != nil {
		ln.Error(context.Background(), err, f, ln.Action("sending webhook"))
		return
	}

	err = s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		ln.Error(context.Background(), err, f, ln.Action("deleting original message"))
		return
	}
}
