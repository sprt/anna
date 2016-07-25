package tasks

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/sprt/anna"
	"github.com/sprt/anna/services/roster"
	"github.com/sprt/anna/services/socialclub"
)

func init() {
	spew.Config.Indent = "\t"
}

func TestMergeRosterAndSC(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(time.Second)

	r := []*roster.Entry{
		// match the latest entry for foo
		{SCUsername: "foo", Timestamp: t1},
		{SCUsername: "foo", Timestamp: t2},
		// bar has no corresponding Social Club entry
		{SCUsername: "bar", Timestamp: t1},
	}
	sc := []*socialclub.Member{
		{Username: "foo"},
		// baz has no corresponding Roster entry
		{Username: "baz"},
	}

	want := []*anna.Member{
		{
			RosterEntry: &roster.Entry{SCUsername: "foo", Timestamp: t2},
			SCMember:    &socialclub.Member{Username: "foo"},
		},
		{
			RosterEntry: &roster.Entry{SCUsername: "bar", Timestamp: t1},
			SCMember:    nil,
		},
		{
			RosterEntry: nil,
			SCMember:    &socialclub.Member{Username: "baz"},
		},
	}

	got := mergeRosterAndSC(r, sc)
	sort.Sort(bySCUsername(got))
	sort.Sort(bySCUsername(want))
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got:\n%s\n\nwant:%s", spew.Sdump(got), spew.Sdump(want))
	}
}

type bySCUsername []*anna.Member

func (m bySCUsername) Len() int      { return len(m) }
func (m bySCUsername) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

func (m bySCUsername) Less(i, j int) bool {
	var a string
	if m[i].RosterEntry != nil {
		a = m[i].RosterEntry.SCUsername
	} else {
		a = m[i].SCMember.Username
	}

	var b string
	if m[j].RosterEntry != nil {
		b = m[j].RosterEntry.SCUsername
	} else {
		b = m[j].SCMember.Username
	}

	return a < b
}
