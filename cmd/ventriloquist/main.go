package main

import (
	"context"
	"os"
	"time"

	"github.com/Xe/ln"
	"github.com/asdine/storm"
	"github.com/bwmarrin/discordgo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/graphite"
	"github.com/golang/groupcache"
	"github.com/joeshaw/envdecode"
	_ "github.com/joho/godotenv/autoload"
	bbot "github.com/withinsoft/ventriloquist/internal/bot"
)

type config struct {
	DiscordToken   string `env:"DISCORD_TOKEN,required"`
	DBPath         string `env:"DB_PATH,default=var/vent.db"`
	AdminRole      string `env:"ADMIN_ROLE,required"`
	GraphiteServer string `env:"GRAPHITE_SERVER,required"`
	LoggingWebhook string `env:"LOGGING_WEBHOOK,required"`
}

func main() {
	ctx := context.Background()
	ctx = ln.WithF(ctx, ln.F{
		"in": "main",
	})

	_ = os.MkdirAll("var", 0700)

	var cfg config
	err := envdecode.StrictDecode(&cfg)
	if err != nil {
		ln.FatalErr(ctx, err)
	}

	ln.DefaultLogger.Filters = append(ln.DefaultLogger.Filters, discordLog(cfg.LoggingWebhook))

	lg := log.NewLogfmtLogger(os.Stdout)
	prov := graphite.New("ventriloquist.", lg)
	go prov.SendLoop(time.Tick(time.Second), "tcp", cfg.GraphiteServer)
	ln.Log(ctx, ln.Action("created metrics client"), ln.F{"address": cfg.GraphiteServer, "protocol": "graphite+tcp"})

	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		ln.FatalErr(ctx, err)
	}
	ln.Log(ctx, ln.Action("discordgo session created"))

	must := func(err error) {
		if err != nil {
			ln.FatalErr(ctx, err)
		}
	}

	db, err := storm.Open(cfg.DBPath)
	if err != nil {
		ln.FatalErr(ctx, err)
	}
	ln.Log(ctx, ln.Action("database opened"))

	const fiftyMegs = 1024 * 1024 * 50
	b := bot{
		cfg: cfg,
		db: DB{
			s: db,
		},
		dg: dg,

		proxiedLine:      prov.NewCounter("discord.messages.proxied.line"),
		messageDeletions: prov.NewCounter("discord.messages.deleted"),
		webhookDuration:  prov.NewHistogram("discord.webhook.execution.ns", 50),
		webhookFailure:   prov.NewCounter("discord.webhook.failure"),
		webhookSuccess:   prov.NewCounter("discord.webhook.success"),
		modForceCtr:      prov.NewCounter("mod.force"),
	}

	b.db.systemmateCache = groupcache.NewGroup("systemmates", fiftyMegs, groupcache.GetterFunc(b.db.cacheSystemmates))

	cs := bbot.NewCommandSet()
	cs.Prefix = ";"

	must(cs.AddCmd("add", "adds a systemmate and optionally their proxy tags", bbot.NoPermissions, b.addSystemmate))
	must(cs.AddCmd("list", "lists systemmates", bbot.NoPermissions, b.listSystemmates))
	must(cs.AddCmd("update", "updates systemmates avatars and optionally name", bbot.NoPermissions, b.updateAvatar))
	must(cs.AddCmd("del", "removes a systemmate", bbot.NoPermissions, b.delSystemmate))
	must(cs.AddCmd("nuke", "removes all system data", bbot.NoPermissions, b.nukeSystem))
	must(cs.AddCmd("chproxy", "changes proxy method for a systemmate", bbot.NoPermissions, b.changeProxy))
	must(cs.AddCmd("export", "exports a copy of all of your data (GDPR compliance)", bbot.NoPermissions, b.export))
	must(cs.AddCmd("mod_list", "mod: lists systemmates for a user", b.modOnly, b.modForce(
		"list",
		"usage: ;mod_list <mention the user>\n\n(don't include the angle brackets)",
		2,
		b.listSystemmates,
	)))
	must(cs.AddCmd("mod_del", "mod: removes a systemmate for a user", b.modOnly, b.modForce(
		"del",
		"usage: ;mod_del <mention the user> <name>\n\n(don't include the angle brackets)",
		3,
		b.delSystemmate,
	)))
	must(cs.AddCmd("mod_update", "mod: removes a systemmate for a user", b.modOnly, b.modForce(
		"update",
		"usage: ;mod_update <mention the user> <name> <new avatar url> <new name>\n\n(don't include the angle brackets)",
		5,
		b.updateAvatar,
	)))
	must(cs.AddCmd("mod_chproxy", "mod: changes proxy method of a systemmate for a user", b.modOnly, b.modForce(
		"update",
		"usage: ;mod_chproxy <mention the user>\n\n(don't include the angle brackets)",
		999,
		b.changeProxy,
	)))
	ln.Log(ctx, ln.Action("added commands to mux"))

	messageCtr := prov.NewCounter("discord.messages.processed")
	botMessageCtr := prov.NewCounter("discord.bot.messages")
	cmdExecDuration := prov.NewHistogram("command.handler.exec.ns", 50)
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		messageCtr.Add(1)
		if m.Author.Bot {
			botMessageCtr.Add(1)
			return
		}

		st := time.Now()
		err := cs.Run(s, m.Message)
		if err != nil {
			ln.Error(context.Background(), err)
		}
		cmdExecDuration.Observe(float64(time.Since(st)))
	})
	dg.AddHandler(b.proxyScrape)
	ln.Log(ctx, ln.Action("added discordgo handlers"))

	err = dg.Open()
	if err != nil {
		ln.FatalErr(ctx, err)
	}
	must(dg.UpdateStatus(0, "memcpy in the cloud"))
	ln.Log(ctx, ln.Action("opened discordgo websocket"))

	ln.Log(ctx, ln.Info("waiting for lines to proxy"), ln.F{"to_discord": true})
	for {
		select {}
	}
}
