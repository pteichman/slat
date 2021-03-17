package slat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"time"

	"github.com/slack-go/slack"
)

// ExportHistory updates the slat archive in outdir to include any new
// messages since the archive was last updated.
func ExportHistory(outdir, token string) error {
	log.Printf("Catching up: %s", outdir)
	api := slack.New(token)

	channels, err := getChannels(api)
	if err != nil {
		return err
	}

	for _, c := range channels {
		err := exportChannelHistory(outdir, api, c)
		if err != nil {
			return err
		}
	}

	return nil
}

func getChannels(api *slack.Client) ([]slack.Channel, error) {
	params := slack.GetConversationsParameters{
		ExcludeArchived: "true",
		Types:           []string{"public_channel"},
	}

	var ret []slack.Channel

	channels, nextCursor, err := api.GetConversations(&params)
	if err != nil {
		return nil, fmt.Errorf("GetConversations: %w", err)
	}
	ret = append(ret, channels...)

	for nextCursor != "" {
		channels, nextCursor, err = api.GetConversations(&params)
		if err != nil {
			return nil, fmt.Errorf("GetConversations: %w", err)
		}
		ret = append(ret, channels...)
	}

	return channels, nil
}

func exportChannelHistory(outdir string, api *slack.Client, channel slack.Channel) error {
	users, err := api.GetUsers()
	if err != nil {
		return fmt.Errorf("GetUsers: %w", err)
	}

	usernames := make(map[string]string)
	for _, user := range users {
		usernames[user.ID] = user.Name
	}

	filename := path.Join(outdir, channel.Name+".json")
	oldest := latestTimestamp(filename)

	history, err := getChannelHistory(api, channel.ID, oldest, "")
	if err != nil {
		return err
	}

	msgs := make([]slack.Message, len(history.Messages))
	copy(msgs, history.Messages)

	for history.HasMore {
		history, err = getChannelHistory(api, channel.ID, oldest, minTimestamp(msgs))
		if err != nil {
			return err
		}
		msgs = append(msgs, history.Messages...)
	}

	sort.Slice(msgs, func(i, j int) bool {
		a, err := strconv.ParseFloat(msgs[i].Timestamp, 64)
		if err != nil {
			panic("a")
		}
		b, err := strconv.ParseFloat(msgs[j].Timestamp, 64)
		if err != nil {
			panic("b")
		}
		return a < b
	})

	log.Printf("Updating: %s: from %s: %d new", channel.Name, oldest, len(msgs))
	if len(msgs) == 0 {
		return nil
	}

	out, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	for _, msg := range msgs {
		if err := enc.Encode(msgEvent(usernames, msg)); err != nil {
			return err
		}
	}

	return nil
}

func msgEvent(usernames map[string]string, msg slack.Message) *event {
	return &event{
		Ts:      msg.Timestamp,
		Type:    msg.Type,
		Subtype: msg.SubType,
		User:    usernames[msg.User],
		Text:    cleanText(usernames, msg.Text),
	}
}

func getChannelHistory(api *slack.Client, chanID string, oldest string, latest string) (*slack.GetConversationHistoryResponse, error) {
	params := slack.GetConversationHistoryParameters{
		ChannelID: chanID,
		Oldest:    oldest,
		Latest:    latest,
		Limit:     1000,
	}

	for {
		history, err := api.GetConversationHistory(&params)
		if rate, ok := err.(*slack.RateLimitedError); ok {
			time.Sleep(rate.RetryAfter)
			continue
		}

		if err != nil {
			return nil, err
		}

		return history, nil
	}
}

func latestTimestamp(filename string) string {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return ""
	}
	buf = bytes.TrimRight(buf, "\n")

	index := bytes.LastIndexByte(buf, '\n')
	if index < 0 {
		index = 0
	}
	buf = buf[index:]

	var e *event
	if err = json.Unmarshal(buf, &e); err != nil {
		return ""
	}
	return e.Ts
}

func minTimestamp(msgs []slack.Message) string {
	min := msgs[0].Timestamp
	for _, msg := range msgs {
		if msg.Timestamp < min {
			min = msg.Timestamp
		}
	}
	return min
}
