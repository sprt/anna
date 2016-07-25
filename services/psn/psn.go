// References:
//
//  - https://github.com/sprt/oauth2-psn/blob/patch-1/DOCS.md
//  - https://github.com/Tustin/psn-php
//  - https://github.com/drasticactions/PsnLib
package psn

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"golang.org/x/time/rate"

	"github.com/sprt/anna/services"
)

const (
	baseURL     = "https://auth.api.sonyentertainmentnetwork.com"
	baseUserURL = "https://us-prof.np.community.playstation.net"

	redirectURI   = "com.playstation.PlayStationApp://redirect"
	serviceEntity = "urn:service-entity:psn"
	scope         = "psn:sceapp,user:account.get,user:account.settings.privacy.get," +
		"user:account.settings.privacy.update,user:account.realName.get," +
		"user:account.realName.update,kamaji:get_account_hash," +
		"kamaji:ugc:distributor,oauth:manage_device_usercodes"
)

var (
	ErrAlreadyFriends         = errors.New("psn: already friends")
	ErrAlreadyFriendRequested = errors.New("psn: already friend-requested")
	ErrFriendRequestNotFound  = errors.New("psn: friend request not found")
	ErrNotInFriendListError   = errors.New("psn: not in friend list")
	ErrUserNotFound           = errors.New("psn: user not found")
)

type User struct {
	Username  string      `json:"onlineId"`
	Presences []*Presence `json:"presences"`
}

type Presence struct {
	Platform       string `json:"platform"`
	GameID         string `json:"npTitleId"`
	GameName       string `json:"titleName"`
	GameStatus     string `json:"gameStatus"` // may be zero
	IsBroadcasting bool   `json:"hasBroadcastData"`
}

type Config struct {
	Username, Email, Password string
	ClientID, ClientSecret    string
	DUID                      string
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

func (c *Client) OnlineFriends() ([]*User, error) {
	req, err := http.NewRequest("GET", baseUserURL+"/userProfile/v1/users/me/friends/profiles2", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("fields", "onlineId,primaryOnlineStatus,presences(@titleInfo,hasBroadcastData)")
	q.Add("sort", "name-onlineId")
	q.Add("userFilter", "online")
	q.Add("offset", "0")
	q.Add("limit", "2000")
	req.URL.RawQuery = q.Encode()

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respData *struct {
		Profiles []*User `json:"profiles"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return nil, err
	}

	return respData.Profiles, nil
}

func (c *Client) Friends() ([]*User, error) {
	return c.friends(friends)
}

func (c *Client) FriendRequests() ([]*User, error) {
	return c.friends(friendRequests)
}

func (c *Client) SentFriendRequests() ([]*User, error) {
	return c.friends(sentFriendRequests)
}

// SendFriendRequest may return ErrAlreadyFriendRequested or ErrUserNotFound.
func (c *Client) SendFriendRequest(username, message string) error {
	data := struct {
		RequestMessage string `json:"requestMessage"`
	}{
		RequestMessage: message,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.doFriendListRequest("POST", username, bytes.NewBuffer(body))
}

// TODO: RemoveFriend, CancelFriendRequest?

// AcceptFriendRequest may return ErrFriendRequestNotFound or ErrUserNotFound.
func (c *Client) AcceptFriendRequest(username string) error {
	return c.doFriendListRequest("PUT", username, nil)
}

// IgnoreFriendRequest may return ErrNotInFriendListError or ErrUserNotFound.
func (c *Client) IgnoreFriendRequest(username string) error {
	return c.doFriendListRequest("DELETE", username, nil)
}

func (c *Client) doFriendListRequest(method, username string, body io.Reader) error {
	url := baseUserURL + fmt.Sprintf("/userProfile/v1/users/%s/friendList/%s", c.config.Username, username)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) friends(status friendStatus) ([]*User, error) {
	url := baseURL + fmt.Sprintf("/userProfile/v1/users/%s/friendList", c.config.Username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("friendStatus", string(status))
	q.Add("fields", "onlineId,region")
	q.Add("offset", "0")
	q.Add("limit", "2000")
	req.URL.RawQuery = q.Encode()

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respData *struct {
		FriendList []*User `json:"friendList"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return nil, err
	}

	return respData.FriendList, nil
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	body := io.TeeReader(resp.Body, &buf)
	resp.Body = ioutil.NopCloser(&buf)

	if resp.StatusCode != 204 {
		var respData *struct {
			Error *struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		err = json.NewDecoder(body).Decode(&respData)
		if err != nil {
			return nil, err
		}
		if e := respData.Error; e != nil {
			switch e.Code {
			case 2107650:
				return nil, ErrAlreadyFriends
			case 2107651:
				return nil, ErrAlreadyFriendRequested
			case 2107648:
				return nil, ErrFriendRequestNotFound
			case 2107649:
				return nil, ErrNotInFriendListError
			case 2105356:
				return nil, ErrUserNotFound
			default:
				return nil, fmt.Errorf("psn: %s (code: %d)", e.Message, e.Code)
			}
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("%s %s: %s", req.Method, req.URL, resp.Status)
		}
	}

	return resp, nil
}

type friendStatus string

const (
	friends            friendStatus = "friend"
	friendRequests                  = "requested"
	sentFriendRequests              = "requesting"
)
