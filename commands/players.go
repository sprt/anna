package commands

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"
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

	var sp, mp []string
	online, err := online(bot)
	if err != nil {
		return err
	}
	for _, gp := range online {
		if gp.name == gtav {
			for _, p := range gp.players {
				escaped := escape(p.name)
				if p.online {
					mp = append(mp, escaped)
				} else {
					sp = append(sp, fmt.Sprintf("*%s*", escaped))
				}
			}
		}
	}
	players := append(mp, sp...)

	var content string
	if len(players) == 0 {
		content = "No one on GTA :( Try `!players all`."
	} else {
		content = strings.Join(players, ", ")
	}
	_, err = bot.Session.ChannelMessageSend(msg.ChannelID, content)
	return err
}

func playersAll(bot *anna.Bot, msg *discordgo.Message, args []string) error {
	online, err := online(bot)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	for _, gp := range online {
		var players []string
		for _, p := range gp.players {
			players = append(players, p.name)
		}
		buf.WriteString(fmt.Sprintf("⦁ **%s** — %s\n", gp.name, strings.Join(players, ", ")))
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

func online(bot *anna.Bot) ([]*gamePlayers, error) {
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
	return groupByGame(online), nil
}

func groupByGame(players []*psn.User) []*gamePlayers {
	groupped := make(map[string][]*player)
	for _, u := range players {
		for _, pres := range u.Presences {
			var online bool
			switch pres.GameName {
			case "":
				continue
			case gtav:
				online = strings.Contains(pres.GameStatus, "Online") || strings.Contains(pres.GameStatus, "Heist")
			}
			groupped[pres.GameName] = append(groupped[pres.GameName], &player{name: u.Username, online: online})
		}
	}

	var list []*gamePlayers
	for game, players := range groupped {
		list = append(list, &gamePlayers{name: game, players: players})
	}

	byGameName := byName(list)
	sort.Sort(byGameName)
	return byGameName
}

type player struct {
	name   string
	online bool
}

type gamePlayers struct {
	name    string
	players []*player
}

type byName []*gamePlayers

func (p byName) Len() int      { return len(p) }
func (p byName) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p byName) Less(i, j int) bool {
	switch {
	case p[i].name == gtav:
		return true
	case p[j].name == gtav:
		return false
	default:
		return p[i].name < p[j].name
	}
}

const gtav = "Grand Theft Auto V"
