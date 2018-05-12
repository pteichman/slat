# slat
Slack tail / Slack archive tool

Some early design notes:

This tool provides JSON output from Slack history, either an archive file (export)
or from the history API. It is designed to keep a full archive up to date, so
after a dump from an export you can run it with an API key and archive new messages.

The output directory is filled with one file per month per channel. These files
contain newline separated JSON objects, one message per object.

It is designed to run from cron.

## Usage

    slat "My Slack export Jan 1 2018.zip"

    SLACK_API_TOKEN="xoxp..." slat
