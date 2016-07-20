package roster

import (
	"encoding/csv"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/sprt/anna"
	"github.com/sprt/anna/services"
)

type Entry struct {
	Timestamp      time.Time
	PSNUsername    string
	SCUsername     string
	RedditUsername string
	Nickname       string
	TimezoneOffset int // in seconds
}

func (e *Entry) unmarshalCSV(fields []string) error {
	t, err := time.Parse("1/2/2006 15:04:05", fields[0])
	if err != nil {
		t, err = time.Parse("1/2/2006", fields[0])
		if err != nil {
			return err
		}
	}
	e.Timestamp = t

	e.PSNUsername = fields[1]
	e.SCUsername = fields[2]
	e.RedditUsername = fields[3]
	e.Nickname = fields[5]

	tz, err := parseTimezone(fields[7])
	if err != nil {
		return err
	}
	e.TimezoneOffset = tz

	return nil
}

type Client struct {
	*services.Client
}

func NewClient(client *http.Client, rl *rate.Limiter) *Client {
	return &Client{
		Client: services.NewClient(client, rl),
	}
}

func (c *Client) Entries() ([]*Entry, error) {
	req, err := http.NewRequest("GET", "https://docs.google.com/spreadsheet/ccc", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("key", anna.Config.GoogleDriveRosterID)
	q.Add("output", "csv")
	req.URL.RawQuery = q.Encode()

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := csv.NewReader(resp.Body)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	entries := make([]*Entry, 0, len(records)-1)
	for _, fields := range records[1:] {
		entry := new(Entry)
		if err := entry.unmarshalCSV(fields); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func parseTimezone(s string) (sec int, err error) {
	parts := strings.Fields(s)
	if len(parts) == 1 {
		return 0, nil
	}
	ds := strings.Replace(parts[1], ":", "h", 1)
	if !strings.Contains(ds, "h") {
		ds += "h0m"
	} else {
		ds += "m"
	}
	d, err := time.ParseDuration(ds)
	if err != nil {
		return 0, err
	}
	return int(d / time.Second), nil
}
