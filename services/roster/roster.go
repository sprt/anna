package roster

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/sprt/anna/services"
)

type Entry struct {
	Timestamp      time.Time
	PSNUsername    string
	SCUsername     string
	RedditUsername string
	Nickname       string
	TZOffset       int // in seconds
}

func (e *Entry) Timezone() string {
	h, m := e.TZOffset/3600, e.TZOffset/60%60
	if m < 0 {
		m = -m
	}
	switch {
	case m != 0:
		return fmt.Sprintf("UTC%+d:%d", h, m)
	case h != 0:
		return fmt.Sprintf("UTC%+d", h)
	default:
		return "UTC"
	}
}

func (e *Entry) LocalTimeAt(t time.Time) time.Time {
	return t.Add(time.Duration(e.TZOffset) * time.Second)
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
	e.TZOffset = tz

	return nil
}

type Client struct {
	*services.Client
	key string
}

func NewClient(key string, client *http.Client, rl *rate.Limiter) *Client {
	return &Client{
		Client: services.NewClient(client, rl),
		key:    key,
	}
}

func (c *Client) Entries() ([]*Entry, error) {
	req, err := http.NewRequest("GET", "https://docs.google.com/spreadsheet/ccc", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("key", c.key)
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

	sort.Sort(byTimestamp(entries))
	return entries, nil
}

type byTimestamp []*Entry

func (r byTimestamp) Len() int           { return len(r) }
func (r byTimestamp) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r byTimestamp) Less(i, j int) bool { return r[i].Timestamp.Before(r[j].Timestamp) }

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
