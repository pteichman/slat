package slat

import "regexp"

// user is the JSON struct from Slack export users.json files.
// Here we only need a restricted set of properties.
type user struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// event is the JSON struct from Slack export channel/YYYY-DD-MM.json files.
// This program also uses event as its own output (channel.json).
type event struct {
	Ts      string `json:"ts"`
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	User    string `json:"user"`
	Text    string `json:"text"`
}

// Regexps for recovering what a user might have typed when Slack makes more
// full-featured text.
var (
	// chanRe is for mentioning channels like #foo.
	chanRe = regexp.MustCompile(`\x{003c}#.*?\|(.*?)\x{003e}`)
	// userRe is for mentioning users like @foo.
	userRe = regexp.MustCompile(`\x{003c}@(.*?)\x{003e}`)
	// linkRe is for mentioning http or https links.
	linkRe = regexp.MustCompile(`\x{003c}(https?://.*?)(\|.*)?\x{003e}`)
)

// cleanText attempts to recover what a user actually typed from text.
func cleanText(usernames map[string]string, text string) string {
	text = chanRe.ReplaceAllString(text, "$1")

	text = userRe.ReplaceAllStringFunc(text, func(match string) string {
		m := userRe.FindStringSubmatch(match)
		return "@" + usernames[m[1]]
	})

	text = linkRe.ReplaceAllString(text, "$1")

	return text
}
