# whgbot-slack
An outgoing webhook bot for slack (https://slack.com/), that spits out information about twitch streams and videos linked.

![Preview](http://i.imgur.com/cVSFhrJ.png "Preview")

# I forked this, what now?
## Development
### PC environment
Make sure you are in the right directory
    1. `go get github.com/mrshankly/go-twitch/twitch`
    2. `go build`
    3. `whgbot-slack.exe -port=3000`

The bot will now be listening on port 3000. (Default is 27015)

### Slack setup
Go to `Team Settings -> Integrations -> Outgoing WebHooks -> Add
    * Channel: Create your own channel for testing purposes
    * Trigger Words: None
    * URL(s): IP/host of your server e.g. yourdomain.com:27015
    * Token: Can be ignored

## Production
You can use goxc (https://github.com/laher/goxc) to cross compile, but 64bit is needed because twitch streamid is int above 13 billion at the moment.

