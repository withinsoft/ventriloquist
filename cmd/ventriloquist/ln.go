package main

import (
	"context"
	"log"

	"github.com/Xe/ln"
)

func discordLog(webhookURL string) ln.FilterFunc {
	return func(ctx context.Context, e ln.Event) bool {
		if _, ok := e.Data["to_discord"]; ok {
			data, err := ln.DefaultFormatter.Format(ctx, e)
			if err != nil {
				log.Printf("error when formatting event: %v", err)
			}

			err = sendWebhook(webhookURL, dWebhook{
				Content:  string(data),
				Username: "Ventriloquist Logging",
			})
			if err != nil {
				log.Printf("can't send log line to discord: %v", err)
			}
		}

		return true
	}
}
