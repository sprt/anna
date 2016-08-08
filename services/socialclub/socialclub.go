package socialclub

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/time/rate"

	"github.com/sprt/anna/services"
)

const (
	baseURL        = "https://socialclub.rockstargames.com"
	baseSupportURL = "https://support.rockstargames.com"
)

type Member struct {
	ID       int            `json:"RockstarId"`
	Username string         `json:"Name"`
	JoinDate socialClubTime `json:"DateJoined"`
}

type Config struct {
	CrewID int
}

type Client struct {
	*services.Client
	config *Config
}

func NewClient(config *Config, client *http.Client, rl *rate.Limiter) *Client {
	return &Client{
		Client: services.NewClient(client, rl),
		config: config,
	}
}

func (c *Client) Status() Status {
	req, err := http.NewRequest("GET", baseSupportURL+"/hc/en-us/articles/200426246-GTA-Online-Server-Status-Latest-Updates", nil)
	if err != nil {
		return StatusDown
	}

	resp, err := c.Do(req)
	if err != nil {
		return StatusDown
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return StatusDown
	}

	var f func(*html.Node) Status
	f = func(n *html.Node) Status {
		if n.Type == html.ElementNode && n.Data == "div" {
			var hasID bool
			var status string
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val == "ps4UpOrDown" {
					hasID = true
				}
				if attr.Key == "data-upordown" {
					status = attr.Val
				}
			}
			if hasID {
				switch status {
				case "Up":
					return StatusUp
				case "Down":
					return StatusDown
				default:
					return StatusLimited
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if s := f(c); s != StatusDown {
				return s
			}
		}
		return StatusDown
	}

	return f(doc)
}

type Status string

const (
	StatusUp      Status = "up"
	StatusLimited        = "limited"
	StatusDown           = "down"
)

func (c *Client) Members() ([]*Member, error) {
	var members []*Member
	var p int
loop:
	for {
		pMembers, err := c.memberList(p)
		switch {
		case err == errServerError:
			continue
		case err != nil:
			return nil, err
		default:
			if len(pMembers) == 0 {
				break loop
			}
			members = append(members, pMembers...)
			p++
		}
	}
	return members, nil
}

func (c *Client) memberList(page int) ([]*Member, error) {
	req, err := http.NewRequest("GET", baseURL+"/crewsapi/GetMembersList", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("crewId", strconv.Itoa(c.config.CrewID))
	q.Add("pageNumber", strconv.Itoa(page))
	req.URL.RawQuery = q.Encode()

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	list := new(memberListResponse)
	err = json.NewDecoder(resp.Body).Decode(list)
	if err != nil {
		if dump, err := httputil.DumpResponse(resp, true); err == nil {
			log.Print(string(dump))
		} else {
			log.Printf("cannot dump response: %s", err)
		}
		return nil, err
	}
	if list.Error != "" {
		return nil, errServerError
	}

	return list.Members, err
}

var errServerError = errors.New("Rockstar server error")

type memberListResponse struct {
	Members []*Member
	Error   string
}

type socialClubTime struct {
	time.Time
}

func (t *socialClubTime) UnmarshalJSON(b []byte) error {
	parsed, err := time.Parse(`"1/2/2006"`, string(b[:]))
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}
