package bot

import "github.com/bwmarrin/discordgo"

func source(s *discordgo.Session, m *discordgo.Message, parv []string) error {
	s.ChannelMessageSend(m.ChannelID, "Source code: https://github.com/withinsoft/ventriloquist")
	return nil
}
