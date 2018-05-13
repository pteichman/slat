package slat

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
)

// ExportArchiveFile writes the messages from a Slack archive .zip file
// to a series of .json files in outdir.
func ExportArchiveFile(outdir, archive string) error {
	log.Printf("Opening %s", archive)
	z, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer z.Close()

	log.Printf("Loading users.json")
	users, err := loadUsers(z, "users.json")
	if err != nil {
		return fmt.Errorf("loading users.json: %s", err)
	}

	usernames := make(map[string]string)
	for _, u := range users {
		usernames[u.ID] = u.Name
	}

	chanEvents := make(map[string][]event)

	for _, f := range z.File {
		name := channelName(f.Name)
		if name == "" {
			continue
		}

		events, err := loadEvents(f, usernames)
		if err != nil {
			return fmt.Errorf("loading events: %s: %s", f.Name, err)
		}

		log.Printf("Loading events: %s: found %d", f.Name, len(events))
		chanEvents[name] = append(chanEvents[name], events...)
	}

	for name, events := range chanEvents {
		sort.Slice(events, func(i, j int) bool {
			a, err := strconv.ParseFloat(events[i].Ts, 64)
			if err != nil {
				panic("a")
			}
			b, err := strconv.ParseFloat(events[j].Ts, 64)
			if err != nil {
				panic("b")
			}
			return a < b
		})

		filename := path.Join(outdir, name+".json")

		log.Printf("Writing events: %s: total %d", filename, len(events))
		out, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer out.Close()

		enc := json.NewEncoder(out)
		for _, event := range events {
			if err := enc.Encode(event); err != nil {
				return err
			}
		}

		out.Close()
	}

	return nil
}

var chanFilenameRe = regexp.MustCompile(`^([^/]+)/[0-9]{4}-[0-9]{2}-[0-9]{2}\.json`)

func channelName(filename string) string {
	m := chanFilenameRe.FindStringSubmatch(filename)
	if len(m) == 0 {
		return ""
	}
	return m[1]
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

func findfile(z *zip.ReadCloser, filename string) (*zip.File, error) {
	for _, f := range z.File {
		if f.Name == filename {
			return f, nil
		}
	}
	return nil, fmt.Errorf("couldn't find %s", filename)
}
