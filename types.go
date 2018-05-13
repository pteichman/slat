package slat

import "regexp"

type user struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type channel struct {
	Name string `json:"name"`
}

type event struct {
	Ts      string `json:"ts"`
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	User    string `json:"user"`
	Text    string `json:"text"`
}

var (
	chanRe = regexp.MustCompile(`\x{003c}#.*?\|(.*?)\x{003e}`)
	userRe = regexp.MustCompile(`\x{003c}@(.*?)\x{003e}`)
	linkRe = regexp.MustCompile(`\x{003c}(https?://.*?)(\|.*)?\x{003e}`)
)

func cleanText(usernames map[string]string, text string) string {
	text = chanRe.ReplaceAllString(text, "$1")

	text = userRe.ReplaceAllStringFunc(text, func(match string) string {
		m := userRe.FindStringSubmatch(match)
		return "@" + usernames[m[1]]
	})

	text = linkRe.ReplaceAllString(text, "$1")

	return text
}
