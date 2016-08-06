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

	online, err := online(bot)
	if err != nil {
		return err
	}
	var players *gamePlayers
	for _, gp := range online {
		if gp.name == gtav {
			players = gp
		}
	}

	var content string
	if players == nil || len(players.players) == 0 {
		content = "No one on GTA :( Try `!players all`."
	} else {
		content = strings.Join(players.list(), ", ")
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
		buf.WriteString(fmt.Sprintf("⦁ **%s** — %s\n", gp.name, strings.Join(gp.list(), ", ")))
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
	for _, p := range players {
		for _, pres := range p.Presences {
			if pres.Platform == "PS4" && pres.GameName != "" {
				groupped[pres.GameName] = append(groupped[pres.GameName], &player{name: p.Username, presence: pres})
			}
		}
	}
	var gp []*gamePlayers
	for game, players := range groupped {
		gp = append(gp, &gamePlayers{name: game, players: players})
	}
	sort.Sort(byName(gp))
	return gp
}

type player struct {
	name     string
	presence *psn.Presence
}

func (p *player) online() bool {
	switch p.presence.GameName {
	case gtav:
		return strings.Contains(p.presence.GameStatus, "Online") || strings.Contains(p.presence.GameStatus, "Heist")
	default:
		return true
	}
}

func (p *player) String() string {
	if !p.online() {
		return fmt.Sprintf("*%s*", escape(p.name))
	}
	return escape(p.name)
}

type gamePlayers struct {
	name    string
	players []*player
}

func (gp *gamePlayers) list() (names []string) {
	var sp, mp []*player
	for _, p := range gp.players {
		if p.online() {
			mp = append(mp, p)
		} else {
			sp = append(sp, p)
		}
	}
	for _, p := range mp {
		names = append(names, p.String())
	}
	for _, p := range sp {
		names = append(names, p.String())
	}
	return
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
