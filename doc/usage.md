# Usage

## Installing

### Install Go

```console
$ sudo -i
# cd /usr/local
# wget https://dl.google.com/go/go1.10.1.linux-amd64.tar.gz
# tar xf go1.10.1.linux-amd64.tar.gz
```

### Clone the repo

```console
$ git clone https://github.com/withinsoft/ventriloquist $HOME/go/src/github.com/withinsoft/ventriloquist
```

### Build the binary

```console
$ GOBIN=. /usr/local/go/bin/go install github.com/withinsoft/ventriloquist/cmd/ventriloquist
```

### Install somewhere

```console
$ sudo mv ventriloquist /usr/local/bin
```

## Running

### tmux/abduco

In a new tmux session, open up a folder and copy down the following script:

```shell
#!/bin/sh
# ventrun.sh

export DISCORD_TOKEN=<discord bot token>
export ADMIN_ROLE=<discord guild moderator permissions flag role>

while true
do
  sleep 2
  /usr/local/bin/ventriloquist
done
```

Run this script in the tmux session (`sh ./ventrun.sh`) and detach (control-B, d)

### Updating

```console
$ cd $HOME/go/src/github.com/withinsoft/ventriloquist
$ git pull
$ GOBIN=. /usr/local/go/bin/go install github.com/withinsoft/ventriloquist/cmd/ventriloquist
$ sudo mv ventriloquist /usr/local/bin
```

## Using

Ventriloquist is controlled by commands. A chat line is considered a command if it starts with the prefix `;`. Example:

```
;foo bar
```

for command `foo` with argument `bar`.

### Register a systemmate

In any channel the bot is in:

```
;add Nicole https://cdn.discordapp.com/avatars/201841370023985153/6879455d380aeb5bd9ee87c02f873e99.png
```

To add a systemmate named Nicole with [this avatar](https://cdn.discordapp.com/avatars/201841370023985153/6879455d380aeb5bd9ee87c02f873e99.png).

Optionally, you can set proxy tags for systemmates on addition:

```
;add Nicole https://cdn.discordapp.com/avatars/201841370023985153/6879455d380aeb5bd9ee87c02f873e99.png [test]
```

### Proxy a message

In any channel the bot is in:

```
Nicole\ Hey there
```

To create a message like:

![Nicole of Cadey~#1337 saying "Hey there"](https://i.imgur.com/5YeMdHg.png)

### Update systemmate details

#### Avatar

In any channel the bot is in:

```
;update Nicole https://cdn.discordapp.com/avatars/201841370023985153/6879455d380aeb5bd9ee87c02f873e99.png
```

to update the avatar of a systemmate.

#### Name

In any channel the bot is in:

```
;update Nicole https://cdn.discordapp.com/avatars/201841370023985153/6879455d380aeb5bd9ee87c02f873e99.png Twitwi
```

to update the avatar and name of a systemmate. You need to give the same avatar url.

### List systemmates

In any channel the bot is in:

```
;list
```

to get back a message like:

```
members:
1. Nicole - https://cdn.discordapp.com/avatars/201841370023985153/6879455d380aeb5bd9ee87c02f873e99.png
```

### Delete systemmates

In any channel the bot is in:

```
;del Nicole
```

to delete a systemmate named Nicole.

### Nuke all data for your system

In any channel the bot is in:

```
;nuke
```

to get your unique delete token

```
;nuke <your-token>
```

to remove all of the data.

### Changing proxy method

In any channel the bot is in:

```
;chproxy
```

then correct it with:

```
;chproxy Nicole [test]
```

and the proxying settings will be immediately updated.

## Moderation

In a perfect world, the following commands will never have to be used.

### List systemmates

In any channel the bot is in

```
;mod_list @Quora
```

to get the systemmate list for the user Quora (actually mention them please).

### Delete systemmate

In any channel the bot is in:

```
;mod_del @Quora Drake
```

to delete the systemmate Drake for the user Quora (actually mention them please).

### Update systemmate

In any channel the bot is in:

```
;mod_update @Quora Drake https://i.imgur.com/4TNNqbD.jpg Naenae
```

to update the systemmate Drake for the user Quora (actually mention them please) to change their avatar to another image and their name to Naenae.

### Changing proxy method

In any channel the bot is in:

```
;mod_chproxy @Quora Drake
```

then correct it with:

```
;mod_chproxy @Quora Drake [test]
```

and the proxying settings will be immediately updated.
