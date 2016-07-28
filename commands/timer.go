package commands

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sprt/anna"
)

func Timer(bot *anna.Bot, msg *discordgo.Message, args []string) error {
	if len(args) != 1 {
		return nil
	}

	n, err := strconv.Atoi(args[0])
	if n > 60 || err != nil {
		return nil
	}

	m, err := bot.Session.ChannelMessageSend(msg.ChannelID, "Ready…")
	if err != nil {
		return nil
	}
	time.Sleep(5 * time.Second)

	tick := time.NewTicker(time.Second).C
	done := time.NewTimer(time.Duration(n) * time.Second).C
loop:
	for {
		select {
		case <-done:
			m, err = bot.Session.ChannelMessageEdit(m.ChannelID, m.ID, "GO!")
			if err != nil {
				return err
			}
			break loop
		case <-tick:
			m, err = bot.Session.ChannelMessageEdit(m.ChannelID, m.ID, strconv.Itoa(n))
			if err != nil {
				return err
			}
			n--
		}
	}

	return nil
}
