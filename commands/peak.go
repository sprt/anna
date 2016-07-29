package commands

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"github.com/sprt/anna"
)

func Peak(bot *anna.Bot, msg *discordgo.Message, args []string) error {
	if len(args) >= 1 {
		switch args[0] {
		case "set":
			return peakSet(bot, msg, args[1:])
		}
	}

	peak := new(anna.Peak)
	if err := bot.DB.View(func(tx *bolt.Tx) error {
		buk := tx.Bucket([]byte("default"))
		b := buk.Get([]byte("peak"))
		if b == nil {
			return nil
		}
		buf := bytes.NewBuffer(b)
		return gob.NewDecoder(buf).Decode(peak)
	}); err != nil {
		return err
	}
	if peak.Count == 0 {
		return nil
	}

	say := fmt.Sprintf("Last peak: %d people online @ %s", peak.Count, peak.Date.Format("Jan 2 2006 15:04 MST"))
	_, err := bot.Session.ChannelMessageSend(msg.ChannelID, say)
	return err
}

func peakSet(bot *anna.Bot, msg *discordgo.Message, args []string) error {
	if msg.Author.ID != bot.OwnerID || len(args) < 2 {
		return nil
	}

	count, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	date, err := time.Parse("Jan 2 2006 15:04 MST", strings.Join(args[1:], " "))
	if err != nil {
		return err
	}

	if err := bot.DB.Update(func(tx *bolt.Tx) error {
		buk := tx.Bucket([]byte("default"))
		peak := &anna.Peak{Count: count, Date: date}
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(peak); err != nil {
			return err
		}
		return buk.Put([]byte("peak"), buf.Bytes())
	}); err != nil {
		return err
	}

	_, err = bot.Session.ChannelMessageSend(msg.ChannelID, "OK")
	return err
}
