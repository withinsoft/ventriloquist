package main

import (
	"context"

	"github.com/Xe/ln"
	"github.com/asdine/storm"
	"github.com/bwmarrin/discordgo"
	"github.com/joeshaw/envdecode"
	_ "github.com/joho/godotenv/autoload"
	bbot "github.com/withinsoft/ventriloquist/internal/bot"
)

type config struct {
	DiscordToken string `env:"DISCORD_TOKEN,required"`
	DBPath       string `env:"DB_PATH,default=var/vent.db"`
}

func main() {
	ctx := context.Background()

	var cfg config
	err := envdecode.StrictDecode(&cfg)
	if err != nil {
		ln.FatalErr(ctx, err)
	}

	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		ln.FatalErr(ctx, err)
	}

	db, err := storm.Open(cfg.DBPath)
	if err != nil {
		ln.FatalErr(ctx, err)
	}

	b := bot{
		cfg: cfg,
		db:  DB{s: db},
		dg:  dg,
	}

	cs := bbot.NewCommandSet()
	cs.AddCmd("add", "adds a systemmate to the list of proxy tags", bbot.NoPermissions, b.addSystemmate)
	cs.AddCmd("list", "lists systemmates", bbot.NoPermissions, b.listSystemmates)

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		cs.Run(s, m.Message)
	})
	dg.AddHandler(b.proxyScrape)

	err = dg.Open()
	if err != nil {
		ln.FatalErr(ctx, err)
	}

	for {
		select {}
	}
}
