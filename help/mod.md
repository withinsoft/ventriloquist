# Moderation

In a perfect world, the following commands will never have to be used.

## List systemmates

In any channel the bot is in

```
${PREFIX}mod_list at_mention_the_user
${PREFIX}mod_list @Quora
```

to get the systemmate list for the user Quora (actually mention them please).

## Delete systemmate

In any channel the bot is in:

```
${PREFIX}mod_del at_mention_the_user name
${PREFIX}mod_del @Quora Drake
```
to delete the systemmate Drake for the user Quora (actually mention them please).

## Update systemmate

In any channel the bot is in:

```
${PREFIX}mod_update at_mention_the_user name direct_image_link new_name
${PREFIX}mod_update @Quora Drake https://i.imgur.com/4TNNqbD.jpg Naenae
```
to update the systemmate Drake for the user Quora (actually mention them please) to change their avatar to another image and their name to Naenae.

## Changing proxy method

In any channel the bot is in:

```
${PREFIX}mod_chproxy at_mention_the_user name proxy_settings
${PREFIX}mod_chproxy @Quora Drake
```

and the proxying settings will be immediately updated.
