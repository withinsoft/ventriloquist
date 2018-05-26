# Ventriloquist

In many online communities, there has been a need for there to be more than more
"voice" per chat connection. Sometimes for roleplay, sometimes to multiplex
several "people" over a single chat connection. This bot seeks to help people
wishing to (for whatever reason) multiplex several voices over a single Discord
connection.

## Requirements

### Server

Ventriloquist is currently only tested on `GOOS=linux GOARCH=amd64`. Other
architectures should work, but may not. If you run into any issues, please
file a bug report. The Docker images are only for `GOOS=linux GOARCH=amd64`.

- Docker + [Docker Compose][compose] (or a swarm)

OR

- Some server that [Go][go] can target binaries for
  - Persistent disk storage for its database file (and backups!)
- [InfluxDB][influxdb] or another [Graphite][graphite] protocol compatible metrics server
  - Runtime metrics, completely anonymous and aggregated data

OPTIONAL

- [Grafana][grafana]
  - To make pretty graphs

### Runtime

Ventriloquist is configured by environment variables. For convenience, you can
define these with a file named `.env` containing key->value environment variables
as such:

```shell
ADMIN_ROLE=442518271057592330
<... etc>
```

These are all required for any deployment of Ventriloquist.

| Environment Variable | Description |
|:-------------------- |:----------- |
| `ADMIN_ROLE` | Users with this Discord guild role will be able to use the moderator commands to impersonate user commands. It is not possible to proxy with these commands. |
| `DISCORD_TOKEN` | The Discord bot token for this bot. This bot should be joined to a Discord guild with the permissions "Manage Messages", "Manage Webhooks", and "Use Cross-Server Emoji". |
| `GRAPHITE_SERVER` | The [Graphite][graphite] server that Ventriloquist sends its anonymous aggregated runtime metrics to, TCP `host:port`. |
| `LOGGING_WEBHOOK` | If set, log the message ID, systemmate name, username + discriminator, user ID, channel ID and proxied message to this [webhook][webhook] after the message being proxied is deleted. This is for moderator accountability of messages being proxied. |

## Using

To start Ventriloquist for yourself, fill out a `.env` file into a checkout of 
this repository and then run the following [`docker-compose`][compose] command:

```console
$ docker-compose up --build -d
```

The database will be located in `/var/lib/docker/volumes/ventriloquist_data/_data/tulpas.db`.
Please configure an automated backup somehow.

To see more about using Ventriloquist as a user, see [the User section of the usage doc][usage].

[go]: https://golang.org
[influxdb]: https://en.wikipedia.org/wiki/InfluxDB
[graphite]: https://graphite.readthedocs.io/en/latest/
[grafana]: https://grafana.com
[webhook]: https://en.wikipedia.org/wiki/Webhook
[usage]: https://github.com/withinsoft/ventriloquist/blob/master/doc/usage.md#using
[compose]: https://docs.docker.com/compose/
