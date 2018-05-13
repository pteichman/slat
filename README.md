# slat
Slack archive tool

This tool maintains a JSON archive of a Slack workspace's public channels.
The archive can be seeded with a full export from Slack, so you will get
your full message history even on a free plan.

slat is designed to run from cron and update the archive incrementally, so
as long as you don't run out of free messages between cron runs, you will
retain all of your messages.

## Motivation

slat was written in the final days of the Slack IRC Gateway to maintain a
training corpus for chatbots. This corpus had previously been maintained
by watching IRC.

The code is a little quick & dirty, but intended for production use.
If you wish to archive more than the bare minimum message history it
keeps now, submissions are welcome.

## Installation

You can build and install the slat command with "go get".

    $ go get github.com/pteichman/slat/cmd/slat

## Usage

To seed an archive directory with an export from Slack, run slat on the
archive zip file:

    $ slat -o /path/to/archive "My Slack export Jan 1 2018.zip"

To update that archive with new messages, you will need a Slack API token.
You can create one that works with slat here: https://api.slack.com/custom-integrations/legacy-tokens

Pass that token in the process environment, like this:

    $ SLACK_API_TOKEN="xoxp..." slat -o /path/to/archive
