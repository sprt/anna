package tasks

import (
	"bytes"
	"encoding/gob"

	"github.com/boltdb/bolt"

	"github.com/sprt/anna"
)

func FetchOnlineFriends(bot *anna.Bot) error {
	online, err := bot.PSN.OnlineFriends()
	if err != nil {
		return err
	}

	if err := bot.DB.Update(func(tx *bolt.Tx) error {
		var b bytes.Buffer
		if err := gob.NewEncoder(&b).Encode(online); err != nil {
			return err
		}
		buk := tx.Bucket([]byte("default"))
		return buk.Put([]byte("online"), b.Bytes())
	}); err != nil {
		return err
	}

	return nil
}
