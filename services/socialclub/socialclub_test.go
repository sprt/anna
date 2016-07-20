package socialclub

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/sprt/anna/services/internal/testutil"
)

var env = new(testutil.Env)

func TestMembers(t *testing.T) {
	env.Setup()
	defer env.Teardown()

	showError := true
	env.Mux.HandleFunc("/crewsapi/GetMembersList", func(w http.ResponseWriter, r *http.Request) {
		p, err := strconv.Atoi(r.URL.Query().Get("pageNumber"))
		if err != nil {
			t.Error(err)
		}

		switch p {
		case 0:
			fmt.Fprint(w, `{"Members": [
			{
				"RockstarId": 29190387,
				"MemberId": 33218356,
				"Name": "VitoCorleone85",
				"AvatarUrl": "n/vitocorleone85",
				"Relationship": "None",
				"AllowAdd": false,
				"AllowBan": false,
				"AllowKick": false,
				"DateJoined": "6/10/2014"
			},
			{
				"RockstarId": 29913246,
				"MemberId": 33222202,
				"Name": "BR-700",
				"AvatarUrl": "n/br-700",
				"Relationship": "None",
				"AllowAdd": false,
				"AllowBan": false,
				"AllowKick": false,
				"DateJoined": "6/10/2014"
			}
			]}`)

		case 1:
			if showError {
				fmt.Fprint(w, `{"Error": "Internal error."}`)
			} else {
				fmt.Fprint(w, `{"Members": [
				{
					"RockstarId": 26556154,
					"MemberId": 41146293,
					"Name": "GENGgar",
					"AvatarUrl": "n/genggar",
					"Relationship": "Friend",
					"AllowAdd": false,
					"AllowBan": false,
					"AllowKick": false,
					"DateJoined": "9/18/2014"
				}
				]}`)
			}
			showError = !showError

		case 2:
			fmt.Fprint(w, `{"Members": []}`)
		}
	})

	want := []*Member{
		{29190387, "VitoCorleone85", socialClubTime{time.Date(2014, 6, 10, 0, 0, 0, 0, time.UTC)}},
		{29913246, "BR-700", socialClubTime{time.Date(2014, 6, 10, 0, 0, 0, 0, time.UTC)}},
		{26556154, "GENGgar", socialClubTime{time.Date(2014, 9, 18, 0, 0, 0, 0, time.UTC)}},
	}

	s := NewClient(env.Client, testutil.InfRateLimiter)
	got, err := s.Members()
	if err != nil {
		t.Error(err)
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
