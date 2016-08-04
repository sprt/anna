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
	if len(args) > 0 {
		switch args[0] {
		case "all":
			return playersAll(bot, msg, args[1:])
		}
	}

	online, err := online(bot)
	if err != nil {
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
				if strings.Contains(pres.GameStatus, "Online") || strings.Contains(pres.GameStatus, "Heist") {
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
		content = "No one on GTA :( Try `!players all`."
	} else {
		content = fmt.Sprintf("**GTA**: %s", strings.Join(players, ", "))
	}
	_, err = bot.Session.ChannelMessageSend(msg.ChannelID, content)
	return err
}

func playersAll(bot *anna.Bot, msg *discordgo.Message, args []string) error {
	online, err := online(bot)
	if err != nil {
		return err
	}

	byGame := make(map[string][]string)
	for _, u := range online {
		for _, pres := range u.Presences {
			if pres.GameName != "" {
				byGame[pres.GameName] = append(byGame[pres.GameName], u.Username)
			}
		}
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%d players:\n", len(online)))
	for game, users := range byGame {
		buf.WriteString(fmt.Sprintf("⦁ **%s** — %s\n", game, strings.Join(users, ", ")))
	}

	var say string
	if len(online) == 0 {
		say = "No one on :("
	} else {
		say = buf.String()
	}
	_, err = bot.Session.ChannelMessageSend(msg.ChannelID, say)
	return err
}

func online(bot *anna.Bot) ([]*psn.User, error) {
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
		return nil, err
	}
	return online, nil
}
