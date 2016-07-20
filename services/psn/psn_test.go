package psn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/sprt/anna"
	"github.com/sprt/anna/services/internal/testutil"
)

type user string

const (
	userAlreadyFriends         user = "alreadyFriends"
	userAlreadyFriendRequested      = "alreadyFriendRequested"
	userNotInFriendList             = "notInFriendList"
	userFriendRequestNotFound       = "friendRequestNotFound"
	userNotFound                    = "notFound"
	userOK                          = "ok"
)

var env = new(testutil.Env)

func TestOnlineFriends(t *testing.T) {
	env.Setup()
	defer env.Teardown()

	env.Mux.HandleFunc("/userProfile/v1/users/me/friends/profiles2", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"profiles": [
		{
			"onlineId": "user1"
		},
		{
			"onlineId": "user2",
			"presences": [{
				"platform": "PS4",
				"npTitleId": "1",
				"titleName": "Foo",
				"gameStatus": "Free roam",
				"hasBroadcastData": true
			}]
		}
		]}`)
	})

	want := []*User{
		{
			Username: "user1",
		},
		{
			Username: "user2",
			Presences: []*Presence{
				{
					Platform:       "PS4",
					GameID:         "1",
					GameName:       "Foo",
					GameStatus:     "Free roam",
					IsBroadcasting: true,
				},
			},
		},
	}

	c := NewClient(env.Client, testutil.InfRateLimiter)
	got, err := c.OnlineFriends()
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != len(want) {
		t.Fatalf("got:\n%v\nwant:\n%v", testutil.Pretty(got), testutil.Pretty(want))
	}
	for i := range got {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("i=%d:\ngot:\n%v\nwant:\n%v", i, testutil.Pretty(got[i]), testutil.Pretty(want[i]))
		}
	}
}

func TestFriends(t *testing.T) {
	env.Setup()
	defer env.Teardown()

	env.Mux.HandleFunc(fmt.Sprintf("/userProfile/v1/users/%s/friendList", anna.Config.PSNUsername), func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("friendStatus") == string(friends) {
			fmt.Fprint(w, `{"friendList": [
				{"onlineId": "foo"},
				{"onlineId": "bar"},
				{"onlineId": "baz"}
			]}`)
		}
	})

	want := []*User{
		{Username: "foo"},
		{Username: "bar"},
		{Username: "baz"},
	}

	c := NewClient(env.Client, testutil.InfRateLimiter)
	got, err := c.Friends()
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != len(want) {
		t.Fatalf("got:\n%v\nwant:\n%v", testutil.Pretty(got), testutil.Pretty(want))
	}
	for i := range got {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("i=%d:\ngot:\n%v\nwant:\n%v", i, testutil.Pretty(got[i]), testutil.Pretty(want[i]))
		}
	}
}

func TestSendFriendRequest(t *testing.T) {
	env.Setup()
	defer env.Teardown()

	reqMsg := "sup"

	pattern := fmt.Sprintf("/userProfile/v1/users/%s/friendList/", anna.Config.PSNUsername)
	env.Mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]interface{})
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			t.Fatal(err)
		}
		if msg := data["requestMessage"]; msg != reqMsg {
			t.Fatalf("requestMessage is %q, want %q", msg, reqMsg)
		}

		username := user(r.URL.Path[len(pattern):])
		switch username {
		case userOK:
			w.WriteHeader(204)
		case userAlreadyFriendRequested:
			fmt.Fprint(w, `{"error": {"code": 2107651, "message": ""}}`)
		case userAlreadyFriends:
			fmt.Fprint(w, `{"error": {"code": 2107650, "message": ""}}`)
		case userNotFound:
			fmt.Fprint(w, `{"error": {"code": 2105356, "message": ""}}`)
		default:
			t.Fatalf("not a test case: username %q", username)
		}
	})

	tests := []struct {
		username user
		err      error
	}{
		{userOK, nil},
		{userAlreadyFriendRequested, ErrAlreadyFriendRequested},
		{userAlreadyFriends, ErrAlreadyFriends},
		{userNotFound, ErrUserNotFound},
	}

	c := NewClient(env.Client, testutil.InfRateLimiter)
	for _, tt := range tests {
		if err := c.SendFriendRequest(string(tt.username), reqMsg); err != tt.err {
			t.Errorf("AcceptFriendRequest(%q) = %q, want %q", tt.username, err, tt.err)
		}
	}
}

func TestAcceptFriendRequest(t *testing.T) {
	env.Setup()
	defer env.Teardown()

	pattern := fmt.Sprintf("/userProfile/v1/users/%s/friendList/", anna.Config.PSNUsername)
	env.Mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		username := user(r.URL.Path[len(pattern):])
		switch username {
		case userOK:
			w.WriteHeader(204)
		case userFriendRequestNotFound:
			fmt.Fprint(w, `{"error": {"code": 2107648, "message": ""}}`)
		case userNotFound:
			fmt.Fprint(w, `{"error": {"code": 2105356, "message": ""}}`)
		default:
			t.Fatalf("not a test case: username %q", username)
		}
	})

	tests := []struct {
		username user
		err      error
	}{
		{userOK, nil},
		{userFriendRequestNotFound, ErrFriendRequestNotFound},
		{userNotFound, ErrUserNotFound},
	}

	c := NewClient(env.Client, testutil.InfRateLimiter)
	for _, tt := range tests {
		if err := c.AcceptFriendRequest(string(tt.username)); err != tt.err {
			t.Errorf("AcceptFriendRequest(%q) = %q, want %q", tt.username, err, tt.err)
		}
	}
}

func TestIgnoreFriendRequest(t *testing.T) {
	env.Setup()
	defer env.Teardown()

	pattern := fmt.Sprintf("/userProfile/v1/users/%s/friendList/", anna.Config.PSNUsername)
	env.Mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		username := user(r.URL.Path[len(pattern):])
		switch username {
		case userOK:
			w.WriteHeader(204)
		case userNotInFriendList:
			fmt.Fprint(w, `{"error": {"code": 2107649, "message": ""}}`)
		case userNotFound:
			fmt.Fprint(w, `{"error": {"code": 2105356, "message": ""}}`)
		default:
			t.Fatalf("not a test case: username %q", username)
		}
	})

	tests := []struct {
		username user
		err      error
	}{
		{userOK, nil},
		{userNotInFriendList, ErrNotInFriendListError},
		{userNotFound, ErrUserNotFound},
	}

	c := NewClient(env.Client, testutil.InfRateLimiter)
	for _, tt := range tests {
		if err := c.IgnoreFriendRequest(string(tt.username)); err != tt.err {
			t.Errorf("IgnoreFriendRequest(%q) = %q, want %q", tt.username, err, tt.err)
		}
	}
}
