package roster

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sprt/anna/services/internal/testutil"
)

var (
	env = new(testutil.Env)
	key = "foo"
)

func TestEntries(t *testing.T) {
	env.Setup()
	defer env.Teardown()

	env.Mux.HandleFunc("/spreadsheet/ccc", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, strings.TrimSpace(`
Date/Time,PSN ID,Social Club ID,reddit,Scheduled Invite Status,Nickname,Microphone,Time Zone,Are you currently a member of GTAA or GTAX?,Notes
2/14/2016 23:43:00,naraic42,dogeshibe,naraic42,Invited 14/02,Ciaran,Yes,GMT,GTAX Xbox,"I was in GTAX and now it's dead, so I know the score (and the truth) when it comes to crew sessions and all that. "
6/4/2016 13:14:09,Gnayt,iJohnson,quatroquesodosfritos,Invited 6/04,Nate,Yes,GMT -8,"NO, I am currently not.",I used to be a member on 360 and just recently got back into GTA on PS4.
5/20/2016 14:06:32,IanKC,IanKSee,iangator,,Piccy,Yes,GMT -5,"NO, I am currently not.",
12/16/2014,CSP2006,ChrisP2006,ChrisP2006,,Chris,Yes,GMT -5,"NO, I am currently not.",I applied on 12/16 but seem to have been deleted.
2/19/2016 17:05:00,hodder99,hoddernut,hodder99,Invited 2/19,hodder,Yes,GMT -3:30,"NO, I am currently not.",
		`))
	})

	want := []*Entry{
		{time.Date(2014, 12, 16, 0, 0, 0, 0, time.UTC), "CSP2006", "ChrisP2006", "ChrisP2006", "Chris", -5 * 3600},
		{time.Date(2016, 2, 14, 23, 43, 0, 0, time.UTC), "naraic42", "dogeshibe", "naraic42", "Ciaran", 0},
		{time.Date(2016, 2, 19, 17, 5, 0, 0, time.UTC), "hodder99", "hoddernut", "hodder99", "hodder", -3.5 * 3600},
		{time.Date(2016, 5, 20, 14, 6, 32, 0, time.UTC), "IanKC", "IanKSee", "iangator", "Piccy", -5 * 3600},
		{time.Date(2016, 6, 4, 13, 14, 9, 0, time.UTC), "Gnayt", "iJohnson", "quatroquesodosfritos", "Nate", -8 * 3600},
	}

	s := NewClient(key, env.Client, testutil.InfRateLimiter)
	got, err := s.Entries()
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
