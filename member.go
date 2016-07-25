package anna

import (
	"github.com/sprt/anna/services/roster"
	"github.com/sprt/anna/services/socialclub"
)

type Member struct {
	// either fields can be nil (exclusively)
	RosterEntry *roster.Entry
	SCMember    *socialclub.Member
}
