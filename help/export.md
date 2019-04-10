# Exporting your data

To export all of your user data:

```
${PREFIX}export
```

Ventriloquist will reply with your data in JSON format. The format of the data is an array of Systemmates. A systemmate contains a unique ID internal to ventriloquist that is used for logging, their name, the discord ID of their host, the URL to their avatar, and a Match object describing how Ventriloquist matches proxied messages.

For more information about proxying methods, see `${PREFIX}help proxy`.
