package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Xe/ln"
	"github.com/bwmarrin/discordgo"
	"github.com/go-kit/kit/metrics"
	"github.com/withinsoft/ventriloquist/internal/proxytag"
)

type bot struct {
	cfg             config
	db              DB
	dg              *discordgo.Session
	lastProxiedUser map[string]time.Time
	lpuLock         *sync.RWMutex

	proxiedLine      metrics.Counter
	messageDeletions metrics.Counter
	webhookDuration  metrics.Histogram
	webhookFailure   metrics.Counter
	webhookSuccess   metrics.Counter
	modForceCtr      metrics.Counter
}

func deleteLater(s *discordgo.Session, dur time.Duration, msgs ...*discordgo.Message) {
	time.Sleep(dur)
	for _, m := range msgs {
		go s.ChannelMessageDelete(m.ChannelID, m.ID)
	}
}

type cmd func(*discordgo.Session, *discordgo.Message, []string) error

func (b bot) modForce(verb, help string, parvlen int, doer cmd) func(*discordgo.Session, *discordgo.Message, []string) error {
	return func(s *discordgo.Session, m *discordgo.Message, parv []string) error {
		if parvlen != 999 {
			if len(parv) != parvlen {
				return errors.New(help)
			}
		}

		mts := m.Mentions
		if len(mts) != 1 {
			return errors.New("please mention the user you want to update as the first argument")
		}

		cparv := []string{";" + verb}
		if parvlen != 0 {
			cparv = append(cparv, parv[2:]...)
		}

		ln.Log(context.Background(), ln.Action("impersonation"), ln.F{
			"command":      verb,
			"parv":         parv,
			"mod_username": m.Author.Username + "#" + m.Author.Discriminator,
			"mod_id":       m.Author.ID,
			"channel":      m.ChannelID,
			"to_discord":   true,
		})
		m.Author.ID = mts[0].ID // hack

		b.modForceCtr.Add(1)
		return doer(s, m, cparv)
	}
}

func (b bot) modOnly(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return err
	}

	gu, err := s.State.Member(ch.GuildID, m.Author.ID)
	if err != nil {
		return err
	}

	for _, r := range gu.Roles {
		if strings.EqualFold(b.cfg.AdminRole, r) {
			return nil
		}
	}

	return errors.New("not authorized")
}

func checkAvatarURL(aurl string) error {
	avatar_url, err := url.Parse(aurl)
	if err != nil {
		return fmt.Errorf("can't parse avatar url: %v", err)
	}

	if !strings.HasPrefix(avatar_url.Scheme, "http") {
		return errors.New("must be a http:// url or https:// url")
	}

	req, err := http.NewRequest("HEAD", aurl, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "image/") {
		return fmt.Errorf("This is not an image. It is Content-Type: %s", ct)
	}

	return nil
}

func (b bot) addSystemmate(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	if len(parv) < 3 {
		return errors.New("usage: `;add <name> <avatar url> <proxy sample>`\n\n(without including the angle brackets; the proxy sample is what your systemmate saying `test` looks like. Check `;help` if in need of further assistance.)")
	}

	name := parv[1]
	aurl := parv[2]
	var err error

	if err := checkAvatarURL(aurl); err != nil {
		return err
	}

	match := proxytag.Match{
		Method: "Nameslash",
		Name:   name,
	}

	if len(parv) > 3 {
		tag := strings.Join(parv[3:], " ")

		log.Printf("tag: %v", parv)

		var err error
		match, err = proxytag.Parse(tag, proxytag.Nameslash, proxytag.Sigils, proxytag.HalfSigilStart, proxytag.HalfSigilEnd)
		if err != nil {
			return err
		}

		if match.Body != "test" {
			return fmt.Errorf("To provide a proxy sample, at the end of the command type what your systemmate saying `test` looks like, and not `%q`", match.Body)
		}

		match.Body = ""
	}

	sm := Systemmate{
		CoreDiscordID: m.Author.ID,
		Name:          name,
		AvatarURL:     aurl,
		Match:         match,
	}

	ln.Log(context.Background(), ln.Action("adding systemmate"), ln.F{
		"name":       name,
		"avatar_url": aurl,
	})

	_, err = b.db.AddSystemmate(sm)
	if err != nil {
		return err
	}

	reply, err := s.ChannelMessageSend(m.ChannelID, "Added member "+sm.Name+" with the following options: "+match.String()+". Please use ;chproxy to customize this further.")

	go deleteLater(s, 30*time.Second, m, reply)

	return err
}

func (b bot) changeProxy(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	const compPhrase = `test`

	if len(parv) == 1 {
		return errors.New("usage: ;chproxy <systemmate name> <proxy them saying '" + compPhrase + "'>\n\n(don't include the angle brackets)")
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

	reply, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s now is set to use the following proxying settings: %s", name, match))
	go deleteLater(s, 30*time.Second, m, reply)
	return err
}

func (b bot) updateAvatar(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	if l := len(parv); l <= 2 {
		return errors.New("usage: ;update <name> <avatar url> [new name]\n\n(don't include the angle/square brackets)")
	}

	name := parv[1]
	aurl := parv[2]
	_, err := url.Parse(aurl)
	if err != nil {
		return fmt.Errorf("can't parse avatar url: %v", err)
	}

	if err = checkAvatarURL(aurl); err != nil {
		return err
	}

	members, err := b.db.FindSystemmates(m.Author.ID)
	if err != nil {
		return err
	}

	var mm Systemmate
	for _, mb := range members {
		if strings.EqualFold(name, mb.Name) {
			mm = mb
		}
	}
	if mm.ID == "" {
		return errors.New("no such systemmate")
	}

	mm.AvatarURL = aurl

	if len(parv) == 4 {
		mm.Name = parv[3]
	}

	err = b.db.UpdateSystemmate(mm)
	if err != nil {
		return err
	}

	reply, err := s.ChannelMessageSend(m.ChannelID, "Updated. Thanks!")
	if err != nil {
		return err
	}
	go deleteLater(s, 30*time.Second, m, reply)

	return b.listSystemmates(s, m, parv)
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

	reply, err := s.ChannelMessageSend(m.ChannelID, sb.String())
	go deleteLater(s, 30*time.Second, m, reply)
	return err
}

func (b bot) delSystemmate(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	if len(parv) != 2 {
		return errors.New("usage: ;del <name>\n\n(don't include the angle brackets)")
	}

	name := parv[1]
	err := b.db.DeleteSystemmate(m.Author.ID, name)
	if err != nil {
		return err
	}

	reply, err := s.ChannelMessageSend(m.ChannelID, "Updated. Thanks!")
	go deleteLater(s, 30*time.Second, m, reply)
	return err
}

func (b bot) nukeSystem(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	tkn := Hash(s.State.User.ID, m.Author.ID)

	if len(parv) != 2 {
		return fmt.Errorf("usage: ;nuke %s\n\nCopy and paste that command to delete all of your systemmates forever (a really long time!)", tkn)
	}

	utkn := parv[1]
	if !strings.EqualFold(tkn, utkn) {
		return errors.New("invalid delete token, see ;nuke")
	}

	err := b.db.NukeSystem(m.Author.ID)
	if err != nil {
		return err
	}

	reply, err := s.ChannelMessageSend(m.ChannelID, "System deleted. Have a good day.")
	go deleteLater(s, 30*time.Second, m, reply)
	return err
}

func (b bot) export(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	sms, err := b.db.FindSystemmates(m.Author.ID)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)

	e := json.NewEncoder(buf)
	e.SetIndent("", "\t")
	err = e.Encode(sms)
	if err != nil {
		return err
	}

	ms := &discordgo.MessageSend{
		Content: "Here is your data in JSON format. To remove all of your data, see `;nuke`.",
		Files: []*discordgo.File{
			&discordgo.File{
				Name:        m.Author.ID + ".json",
				ContentType: "application/json",
				Reader:      buf,
			},
		},
	}

	authorChannel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		return err
	}

	msg, err := s.ChannelMessageSendComplex(authorChannel.ID, ms)
	if err != nil {
		return err
	}

	ln.Log(context.Background(), ln.Info("user data exported"), ln.F{
		"to_discord": true,
		"author_id":  m.Author.ID,
		"message_id": msg.ID,
		"channel_id": m.ChannelID,
	})

	go deleteLater(s, 30*time.Second, m)

	return nil
}

func (b *bot) proxyScrape(s *discordgo.Session, m *discordgo.MessageCreate) {
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

	msg = strings.Replace(msg, "@here", "at-here", -1)
	msg = strings.Replace(msg, "@everyone", "at-everyone", -1)
	msg = strings.Replace(msg, "@channel", "at-channel", -1)
	msg = strings.Replace(msg, "@someone", "at-someone", -1)

	member, matchBody, err := b.db.FindSystemmateByMessage(m.Author.ID, msg)
	if err != nil {
		if err.Error() != "not found" || !strings.Contains(err.Error(), "systemmate not found") {
			ln.Error(ctx, err, f, ln.Action("find systemmate by match"))
		}

		return
	}

	f["name"] = "Generic"
	f["member_id"] = member.ID
	f["member_name"] = member.Name
	f["proxy_match"] = member.Match.String()
	b.proxiedLine.Add(1)

	wh, err := b.db.FindWebhook(m.ChannelID)
	if err != nil {
		if err.Error() == "not found" {
			whook, err := s.WebhookCreate(m.ChannelID, "ventriloquist proxy bot", "https://cdn.discordapp.com/attachments/376041158842908672/442528694762864660/unknown.png")
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

	wantEbds := []embeds{
		{
			Footer: embedFooter{
				Text:    fmt.Sprintf("%s#%s", m.Author.Username, m.Author.Discriminator),
				IconURL: m.Author.AvatarURL("32"),
			},
		},
	}

	var ebds []embeds
	b.lpuLock.RLock()
	lpt, ok := b.lastProxiedUser[m.ChannelID+"/"+m.Author.ID]
	b.lpuLock.RUnlock()
	if !ok {
		ebds = wantEbds
		goto skipEbds
	}

	ln.Log(ctx, f, ln.Action("lastProxiedUserTest"), ln.F{"m_author_id": m.Author.ID, "last_proxied_time": lpt})
	if time.Now().After(lpt.Add(b.cfg.MessageTagLifetime)) {
		ebds = wantEbds
	}

skipEbds:
	dw := dWebhook{
		Content:   matchBody,
		Username:  member.Name,
		AvatarURL: member.AvatarURL,
		Embeds:    ebds,
	}

	t0 := time.Now()
	err = sendWebhook(wh, dw)
	if err != nil {
		b.webhookFailure.Add(1)
		ln.Error(context.Background(), err, f, ln.Action("sending webhook"))
		return
	}
	b.webhookDuration.Observe(float64(time.Since(t0)))
	b.webhookSuccess.Add(1)

	b.lpuLock.Lock()
	b.lastProxiedUser[m.ChannelID+"/"+m.Author.ID] = time.Now()
	b.lpuLock.Unlock()

	err = s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		f["to_discord"] = true
		ln.Error(context.Background(), err, f, ln.Action("deleting original message"))
		return
	}
	ln.Log(ctx, ln.Action("deleted message"), f)
	b.messageDeletions.Add(1)

	err = sendWebhook(b.cfg.LoggingWebhook, dWebhook{
		Content:  fmt.Sprintf("%s: %s of %s#%s (%s) in <#%s>: %s", m.ID, member.Name, m.Author.Username, m.Author.Discriminator, m.Author.ID, m.ChannelID, matchBody),
		Username: "Ventriloquist Logging",
	})
	if err != nil {
		ln.Error(ctx, err, f, ln.Info("can't send log message to discord"))
	}
}
