package tasks

import (
	"bytes"
	"encoding/gob"
	"log"
	"os"

	"github.com/boltdb/bolt"

	"github.com/sprt/anna"
)

func SendFriendRequests(bot *anna.Bot) error {
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

	var send []string
	if err := bot.DB.View(func(tx *bolt.Tx) error {
		buk := tx.Bucket([]byte("friended"))
		for _, m := range members {
			if m.SCMember != nil && (m.RosterEntry != nil && buk.Get([]byte(m.RosterEntry.PSNUsername)) == nil) {
				send = append(send, m.RosterEntry.PSNUsername)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	log.Printf("%d eligible for a friend request", len(send))
	if os.Getenv("ANNA_PROD") == "1" {
		for _, username := range send {
			err := bot.PSN.SendFriendRequest(username, bot.FriendRequestMsg)
			if err != nil {
				log.Printf("ERROR sending friend request to %s: %v", username, err)
				continue
			}
			if err := bot.DB.Update(func(tx *bolt.Tx) error {
				buk := tx.Bucket([]byte("friended"))
				return buk.Put([]byte(username), []byte(""))
			}); err != nil {
				log.Printf("WARNING could not store in DB that we friended %s", username)
				continue
			}
			log.Printf("Sent a friend request to %s", username)
		}
		log.Printf("%d friend requests sent", len(send))
	} else {
		log.Printf("Dev env, not sending friend requests")
	}

	return nil
}
