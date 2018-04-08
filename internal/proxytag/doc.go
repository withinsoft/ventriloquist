/*
Package proxytag scrapes incoming text messages for "proxy tags" as used by the
tulpamancy and other plurality communites. At a high level, "proxy tags" are a
hack to get around the fact that there are legitimately multiple voices speaking
through one connection, and the tags signify who is speaking. There are many
styles of this, but the most common boils down to a prefix/suffix sigil on each
message.

As an example:

    [Hey there]
    Hi there

The words inside the square brackets are proxied for whoever the context of the
user doing the proxying has square brackets. There is no way to know who is
speaking from only the text, only that the user is in square brackets.

Another common style is known as the "name-slash" method.

As an example:

    Nicole\ Hey there
    Hi there

The name prefixed by the backslash indicates who is speaking.

This package hopefully will be useful when parsing these lines for bots or other
automated processing.
*/
package proxytag
