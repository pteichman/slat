package slat

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
)

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

// ExportArchiveFile writes the messages from a Slack archive .zip file
// to a series of .json files in outdir.
func ExportArchiveFile(outdir, archive string) error {
	z, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer z.Close()

	users, err := loadUsers(z, "users.json")
	if err != nil {
		return fmt.Errorf("loading users.json: %s", err)
	}

	usernames := make(map[string]string)
	for _, u := range users {
		usernames[u.ID] = u.Name
	}

	for _, f := range z.File {
		if !isChannelLog(f.Name) {
			continue
		}

		events, err := loadEvents(f, usernames)
		if err != nil {
			return fmt.Errorf("loading events: %s: %s", f.Name, err)
		}

		subdir, filename := path.Split(f.Name)
		if err := os.MkdirAll(path.Join(outdir, subdir), os.ModePerm); err != nil {
			return err
		}
		out, err := os.Create(path.Join(outdir, subdir, filename))
		if err != nil {
			return err
		}
		defer out.Close()

		if err := json.NewEncoder(out).Encode(events); err != nil {
			return err
		}

		out.Close()
	}

	return nil
}

var chanFilenameRe = regexp.MustCompile(`[^/]/[0-9]{4}-[0-9]{2}-[0-9]{2}\.json`)

func isChannelLog(filename string) bool {
	return chanFilenameRe.MatchString(filename)
}

func loadUsers(z *zip.ReadCloser, filename string) ([]user, error) {
	f, err := findfile(z, filename)
	if err != nil {
		return nil, err
	}

	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var ret []user
	if err := json.NewDecoder(r).Decode(&ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func loadChannels(z *zip.ReadCloser, filename string) ([]channel, error) {
	f, err := findfile(z, filename)
	if err != nil {
		return nil, err
	}

	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var ret []channel
	if err := json.NewDecoder(r).Decode(&ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func loadEvents(f *zip.File, usernames map[string]string) ([]event, error) {
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var events []event
	if err := json.NewDecoder(r).Decode(&events); err != nil {
		return nil, err
	}

	for i, e := range events {
		events[i].User = usernames[e.User]
		events[i].Text = cleanText(usernames, e.Text)
	}

	return events, nil
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

func findfile(z *zip.ReadCloser, filename string) (*zip.File, error) {
	for _, f := range z.File {
		if f.Name == filename {
			return f, nil
		}
	}
	return nil, fmt.Errorf("couldn't find %s", filename)
}
