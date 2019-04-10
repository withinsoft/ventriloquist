# Proxy a Message

In any channel the bot is in, proxy your systemmate like normal. For example given a systemmate named Nicole with Nameslash proxing:

```
Nicole\ hey
```

to create a message like this: https://i.imgur.com/wbonssx.png

## Proxy Methods

### Nameslash

Nameslash proxying is done by using the full name followed by a backslash, a space, and the text to proxy:

```
Nicole\ Hey there
```

Alternatively, you can use a shortened nickname when outlining the backspace proxy method:

```
N\ Hey there
```

### Sigils

Sigils looks for the first and last "sigils" of a message in order to figure out which systemmate is speaking. These can be any unicode symbol or punctuation character, minus discord formatting characters. Example: 

```
[Hey there]
```

This works for any number of sigils. Optionally you can also use sigils at one end only:

```
[Hey there
```

or:

```
Hey there]

```

To help with the sigil recognition, it is useful to separate any additional special characters (ie. not alphanumeric) that are not part of the sigil from the actual sigil by putting a space inbetween them. (This doesn't apply for characters in use by Discord already, such as `*` or `_`)

```
[ !!! I didn't know that]
```

Please see `${PREFIX}help list` and `${PREFIX}help chproxy` for more information on how to list systemmates and change proxy methods and `${PREFIX}help add` to add systemmates to proxy.
