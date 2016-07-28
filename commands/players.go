package commands

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"github.com/sprt/anna"
	"github.com/sprt/anna/services/psn"
)

func Players(bot *anna.Bot, msg *discordgo.Message, args []string) error {
	var online []*psn.User
	if err := bot.DB.View(func(tx *bolt.Tx) error {
		buk := tx.Bucket([]byte("default"))
		b := buk.Get([]byte("online"))
		if b == nil {
			return nil
		}
		buf := bytes.NewBuffer(b)
		return gob.NewDecoder(buf).Decode(&online)
	}); err != nil {
		return err
	}

	var sp, mp []string
	for _, u := range online {
		for _, pres := range u.Presences {
			// TODO: use GameID (region-dependent)
			// EU: CUSA00411_00
			// US: CUSA00419_00
			if pres.GameName == "Grand Theft Auto V" {
				escaped := escape(u.Username)
				if strings.Contains(pres.GameStatus, "Online") {
					mp = append(mp, escaped)
				} else {
					sp = append(sp, "*"+escaped+"*")
				}
			}
		}
	}
	players := append(mp, sp...)

	var content string
	if len(players) == 0 {
		content = "No one on :("
	} else {
		content = fmt.Sprintf("%d players: %s", len(players), strings.Join(players, ", "))
	}
	_, err := bot.Session.ChannelMessageSend(msg.ChannelID, content)
	return err
}
