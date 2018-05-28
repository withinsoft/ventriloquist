# Proxy a Message

In any channel the bot is in, proxy your systemmate like normal. For example given a systemmate named Nicole with Nameslash proxing:

```
Nicole\ hey
```

to create a message like this: https://i.imgur.com/wbonssx.png

## Proxy Methods

### Nameslash

Nameslash proxying is fairly simple:

```
Nicole\ Hey there
```

is Nicole saying "Hey there".

### Sigils

Sigils looks for the first and last "sigils" of a message in order to figure out which systemmate is speaking. These can be any unicode symbol or punctuation character, minus discord formatting characters. Example: 

```
[Hey there]
```

This works for any number of sigils. Optionally you can proxy messages like this:

```
[Hey there
```

or like this:

```
Hey there]
```

Please see `${PREFIX}help list` and `${PREFIX}help chproxy` for more information on how to list systemmates and change proxy methods and `${PREFIX}help add` to add systemmates to proxy.
