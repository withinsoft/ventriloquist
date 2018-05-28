package bot

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/withinsoft/ventriloquist/help/statik"
)

func (cs *CommandSet) replace(verb string) string {
	switch strings.ToLower(verb) {
	case "prefix":
		return cs.Prefix
	}

	return "<unknown verb " + verb + ">"
}

func (cs *CommandSet) genHelp(verb string) (string, error) {
	fin, err := cs.helpFS.Open("/" + verb + ".md")
	if err != nil {
		return "error", err
	}
	defer fin.Close()

	data, err := ioutil.ReadAll(fin)
	if err != nil {
		return "error", err
	}

	return os.Expand(string(data), cs.replace), nil
}

func (cs *CommandSet) help(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	switch len(parv) {
	case 1:
		result, err := cs.genHelp("index")
		if err != nil {
			return err
		}

		authorChannel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return err
		}

		s.ChannelMessageSend(authorChannel.ID, result)

		todel, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> check direct messages, help is there!", m.Author.ID))
		if err != nil {
			return err
		}

		go func() {
			time.Sleep(30 * time.Second)
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			s.ChannelMessageDelete(todel.ChannelID, todel.ID)
		}()

	case 2:
		verb := parv[1]

		result, err := cs.genHelp(verb)
		if err != nil {
			return err
		}

		authorChannel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return err
		}

		s.ChannelMessageSend(authorChannel.ID, result)

		go func() {
			time.Sleep(30 * time.Second)
			s.ChannelMessageDelete(m.ChannelID, m.ID)
		}()
	}

	return nil
}

func (cs *CommandSet) formHelp() string {
	result := "Bot commands: \n"

	for verb, cmd := range cs.cmds {
		result += fmt.Sprintf("%s%s: %s\n", cs.Prefix, verb, cmd.Helptext())
	}

	return (result + "If there's any problems please don't hesitate to ask a server admin for help.")
}
