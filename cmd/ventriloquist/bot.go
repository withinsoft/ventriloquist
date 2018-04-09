package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Xe/ln"
	"github.com/bwmarrin/discordgo"
	"github.com/withinsoft/ventriloquist/internal/proxytag"
)

type bot struct {
	cfg config
	db  DB
	dg  *discordgo.Session
}

func (b bot) addSystemmate(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	if len(parv) != 3 {
		return errors.New("usage: .add <name> <avatar url>\n\n(don't include the angle brackets)")
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
		Match: proxytag.Match{
			Method: "Nameslash",
			Name:   name,
		},
	}

	ln.Log(context.Background(), ln.Action("adding systemmate"), ln.F{
		"name":       name,
		"avatar_url": aurl,
	})

	_, err = b.db.AddSystemmate(sm)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Added member "+sm.Name+" with default Nameslash proxying. Please use .chproxy to customize this further.")
	return err
}

func (b bot) changeProxy(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	const compPhrase = `i am listening for a sound beyond sound`

	if len(parv) == 1 {
		return errors.New("usage: .chproxy <tulpa name> <proxy them saying '" + compPhrase + "'>\n\n(don't include the angle brackets)")
	}

	name := parv[1]
	line := strings.Join(parv[2:], " ")
	match, err := proxytag.Parse(line, proxytag.Nameslash, proxytag.Sigils, proxytag.HalfSigilStart, proxytag.HalfSigilEnd)
	if err != nil {
		return err
	}

	if !strings.EqualFold(match.Body, compPhrase) {
		return fmt.Errorf("please proxy %q", compPhrase)
	}

	var member Systemmate
	sms, err := b.db.FindSystemmates(m.Author.ID)
	if err != nil {
		return err
	}

	for _, sm := range sms {
		if strings.EqualFold(name, sm.Name) {
			member = sm
		}
	}

	if member.ID == "" {
		return fmt.Errorf("no such systemmate %q, check .list", name)
	}

	member.Match = match
	err = b.db.UpdateSystemmate(member)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s now is set to use the following proxying settings: %s", name, match))
	return err
}

func (b bot) updateAvatar(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	if len(parv) != 3 {
		return errors.New("usage: .update <name> <avatar url>\n\n(don't include the angle brackets)")
	}

	name := parv[1]
	aurl := parv[2]
	_, err := url.Parse(aurl)
	if err != nil {
		return fmt.Errorf("can't parse avatar url: %v", err)
	}

	members, err := b.db.FindSystemmates(m.Author.ID)
	if err != nil {
		return err
	}

	var mm Systemmate
	for _, m := range members {
		if strings.EqualFold(name, m.Name) {
			mm = m
		}
	}
	if mm.ID == "" {
		return errors.New("no such systemmate")
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Updated. Thanks!")
	return err
}

func (b bot) listSystemmates(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	members, err := b.db.FindSystemmates(m.Author.ID)
	if err != nil {
		return err
	}

	sb := strings.Builder{}
	sb.WriteString("members:\n")
	for i, m := range members {
		sb.WriteString(fmt.Sprintf("%d. %s - <%s> - proxy details: %s\n", (i + 1), m.Name, m.AvatarURL, m.Match))
	}

	_, err = s.ChannelMessageSend(m.ChannelID, sb.String())
	return err
}

func (b bot) delSystemmate(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	if len(parv) != 2 {
		return errors.New("usage: .del <name>\n\n(don't include the angle brackets)")
	}

	name := parv[1]
	err := b.db.DeleteSystemmate(m.Author.ID, name)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Updated. Thanks!")
	return err
}

func (b bot) nukeSystem(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	tkn := Hash(s.State.User.ID, m.Author.ID)

	if len(parv) != 2 {
		return fmt.Errorf("usage: .nuke %s\n\nThe token shown is your unique delete token", tkn)
	}

	utkn := parv[1]
	if !strings.EqualFold(tkn, utkn) {
		return errors.New("invalid delete token, see .nuke")
	}

	err := b.db.NukeSystem(m.Author.ID)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "System deleted. Have a good day.")
	return err
}

func (b bot) proxyScrape(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if m.Author.Bot {
		return
	}

	msg := m.Content
	f := ln.F{
		"channel_id":      m.ChannelID,
		"author_id":       m.Author.ID,
		"author_username": m.Author.Username + "#" + m.Author.Discriminator,
		"message_id":      m.ID,
	}

	match, err := proxytag.Parse(msg, proxytag.Nameslash, proxytag.Sigils, proxytag.HalfSigilStart, proxytag.HalfSigilEnd)
	if err != nil {
		if err == proxytag.ErrNoMatch {
			// don't care, not a proxied line, yolo
			ln.Log(ctx, f, ln.Action("not a proxied line"))
			return
		}

		ln.Error(ctx, err, f, ln.Action("looking for proxied lines"))
	}

	f["name"] = match.Name

	member, err := b.db.FindSystemmateByMatch(m.Author.ID, match)
	if err != nil {
		ln.Error(ctx, err, f, ln.Action("find systemmate by match"))
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
		Content:   match.Body,
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
	ln.Log(ctx, ln.Action("deleted message"), f)
}
