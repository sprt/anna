package socialclub

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"github.com/sprt/anna"
	"github.com/sprt/anna/services"
)

type Member struct {
	ID       int            `json:"RockstarId"`
	Username string         `json:"Name"`
	JoinDate socialClubTime `json:"DateJoined"`
}

type Client struct {
	*services.Client
}

func NewClient(client *http.Client, rl *rate.Limiter) *Client {
	return &Client{
		Client: services.NewClient(client, rl),
	}
}

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
	req, err := http.NewRequest("GET", "https://socialclub.rockstargames.com/crewsapi/GetMembersList", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("crewId", strconv.Itoa(anna.Config.SocialClubCrewID))
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
	parsed, err := time.Parse(`"1/02/2006"`, string(b[:]))
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}
