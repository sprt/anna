package tasks

import (
	"bytes"
	"encoding/gob"
	"log"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/sprt/anna"
	"github.com/sprt/anna/services/roster"
	"github.com/sprt/anna/services/socialclub"
)

func FetchMembers(b *anna.Bot) error {
	roster, err := b.Roster.Entries()
	if err != nil {
		return err
	}
	sc, err := b.SocialClub.Members()
	if err != nil {
		return err
	}
	members := mergeRosterAndSC(roster, sc)

	if err := b.DB.Update(func(tx *bolt.Tx) error {
		var bMembers bytes.Buffer
		if err := gob.NewEncoder(&bMembers).Encode(members); err != nil {
			return err
		}

		bu := tx.Bucket([]byte("default"))
		return bu.Put([]byte("members"), bMembers.Bytes())
	}); err != nil {
		return err
	}

	log.Printf("Fetched %d members", len(members))

	return nil
}

func mergeRosterAndSC(r []*roster.Entry, sc []*socialclub.Member) []*anna.Member {
	memberMap := make(map[string]*anna.Member)
	for _, entry := range r {
		memberMap[strings.ToLower(entry.SCUsername)] = &anna.Member{RosterEntry: entry}
	}
	for _, mem := range sc {
		key := strings.ToLower(mem.Username)
		if memberMap[key] == nil {
			memberMap[key] = new(anna.Member)
		}
		memberMap[key].SCMember = mem
	}

	members := make([]*anna.Member, 0, len(memberMap))
	for _, member := range memberMap {
		members = append(members, member)
	}

	return members
}
