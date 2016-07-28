package tasks

import (
	"bytes"
	"encoding/gob"
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/sprt/anna"
)

func AcceptFriendRequests(bot *anna.Bot) error {
	var members []*anna.Member
	if err := bot.DB.View(func(tx *bolt.Tx) error {
		buk := tx.Bucket([]byte("default"))
		b := buk.Get([]byte("members"))
		if b == nil {
			return nil
		}
		buf := bytes.NewBuffer(b)
		return gob.NewDecoder(buf).Decode(&members)
	}); err != nil {
		return err
	}
	acceptMembers := make(map[string]struct{})
	for _, m := range members {
		if m.SCMember != nil && m.RosterEntry != nil {
			acceptMembers[m.RosterEntry.PSNUsername] = struct{}{}
		}
	}

	reqs, err := bot.PSN.FriendRequests()
	if err != nil {
		return err
	}

	var accept []string
	for _, u := range reqs {
		if _, ok := acceptMembers[u.Username]; ok {
			accept = append(accept, u.Username)
		}
	}

	log.Printf("%d eligible for friending", len(accept))
	if os.Getenv("ANNA_PROD") == "1" {
		for _, username := range accept {
			if err := bot.PSN.AcceptFriendRequest(username); err != nil {
				log.Printf("ERROR accepting friend request from %s: %v", username, err)
				continue
			}
			if err := bot.DB.Update(func(tx *bolt.Tx) error {
				buk := tx.Bucket([]byte("friended"))
				return buk.Put([]byte(username), []byte(""))
			}); err != nil {
				log.Printf("WARNING could not store in DB that we friended %s", username)
				continue
			}
			log.Printf("Accepted friend request from %s", username)
		}
		log.Printf("Accepted %d friend requests", len(accept))
	} else {
		log.Printf("Dev env, not accepting friend requests")
	}

	return nil
}
