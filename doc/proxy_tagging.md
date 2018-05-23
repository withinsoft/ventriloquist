# Proxy Tagging as Relevant to This Bot

In many communities online, it has been considered useful to introduce
additional "speakers" into the chatroom whether for roleplay, sharing a
computer between multiple people or [tulpamancy][tulpamancy]. This bot is
designed to help ease the pains of simulating the presence of multiple speakers
multiplexed across one chat connection by parsing out tags in how people type.
It is designed to help solve the problem of having multiple speakers behind a
single chat connection, especially for users of iOS.

A proxy tag is basically an envelope around the message in order to help people
decipher the speaker of the message. An example follows:

```
[Hey there, I'm talking to you via square brackets.]
```

This message is tagged to originate from the speaker that speaks in square
brackets. This information should hopefully be close at hand, as this doesn't
tell who said what, only that someone that speaks in square brackets did.

However, the combination of a proxy tag and the discord user ID functions as
a unique ID for an individual speaker. Using this and a stored avatar URL,
we can send messages to Discord via [webhooks][webhooks] and allow them to speak
in chatrooms almost like they had their own accounts and connections to begin
with.

[tulpamancy]: https://www.tulpa.info
[webhooks]: https://support.discordapp.com/hc/en-us/articles/228383668-Intro-to-Webhooks
