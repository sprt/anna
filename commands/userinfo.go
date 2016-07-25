package commands

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"github.com/sprt/anna"
)

// set by UserInfo before listing members' local times
var utcNow time.Time

func UserInfo(bot *anna.Bot, msg *discordgo.Message, args []string) error {
	members := make([]*anna.Member, 0)

	if len(args) > 1 {
		return nil
	}

	if err := bot.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("default"))
		bMembers := b.Get([]byte("members"))
		if bMembers == nil {
			return nil
		}
		return gob.NewDecoder(bytes.NewBuffer(bMembers)).Decode(&members)
	}); err != nil {
		return err
	}

	utcNow = time.Now().UTC()

	switch len(args) {
	case 0:
		// TODO
	case 1:
		q := args[0]
		matching := searchMembers(members, q)
		num := len(matching)

		results := make([]string, 0, len(matching))
		for _, m := range matching {
			results = append(results, memberAsResult(m, q))
		}

		var say string

		switch num {
		case 0:
			say = fmt.Sprintf("No match for %q", q)
		case 1:
			say = results[0]
		case 2, 3:
			s := make([]string, 1, 1+num)
			s[0] = fmt.Sprintf("%d matches:", num)
			for _, res := range results {
				s = append(s, fmt.Sprintf("⦁ %s", res))
			}
			say = strings.Join(s, "\n")
		default:
			s := make([]string, 1, 4)
			s[0] = fmt.Sprintf("%d matches, showing first 3:", num)
			for _, res := range results[:3] {
				s = append(s, fmt.Sprintf("⦁ %s", res))
			}
			say = strings.Join(s, "\n")
		}

		_, err := bot.Session.ChannelMessageSend(msg.ChannelID, say)
		if err != nil {
			return err
		}
	}

	return nil
}

func memberAsResult(m *anna.Member, q string) string {
	fields := make([]string, 0, 6)
	if m.RosterEntry != nil {
		fields = append(fields, fmt.Sprintf("PSN: %s", highlightSubstr(m.RosterEntry.PSNUsername, q)))
		fields = append(fields, fmt.Sprintf("Reddit: %s", highlightSubstr(m.RosterEntry.RedditUsername, q)))
	}
	if m.RosterEntry != nil {
		fields = append(fields, fmt.Sprintf("SC: %s", highlightSubstr(m.RosterEntry.SCUsername, q)))
	} else {
		fields = append(fields, fmt.Sprintf("SC: %s", highlightSubstr(m.SCMember.Username, q)))
	}
	if m.RosterEntry != nil {
		fields = append(fields, fmt.Sprintf("Nick: %s", highlightSubstr(m.RosterEntry.Nickname, q)))
	}
	if m.SCMember != nil {
		fields = append(fields, fmt.Sprintf("Joined: %s", m.SCMember.JoinDate.Format("Jan 2 2006")))
	}
	if m.RosterEntry != nil {
		fields = append(fields, fmt.Sprintf("Timezone: %s (%s)", m.RosterEntry.Timezone(),
			m.RosterEntry.LocalTimeAt(utcNow).Format("15:04")))
	}

	var buf bytes.Buffer
	buf.WriteString(strings.Join(fields, ", "))
	if m.SCMember == nil {
		buf.WriteString(" `not in crew`")
	}
	if m.RosterEntry == nil {
		buf.WriteString(" **`not on roster`**")
	}

	return buf.String()
}

func highlightSubstr(s, substr string) string {
	pos := strings.Index(strings.ToLower(s), strings.ToLower(substr))
	if pos == -1 {
		return escape(s)
	}
	start := s[:pos]
	mid := s[pos : pos+len(substr)]
	end := s[pos+len(substr):]
	return fmt.Sprintf("%s**%s**%s", escape(start), escape(mid), escape(end))
}

func escape(s string) string {
	for _, c := range []string{"_", "*", "`", "~"} {
		s = strings.Replace(s, c, "\\"+c, -1)
	}
	return s
}

func searchMembers(members []*anna.Member, q string) (matches []*anna.Member) {
	// TODO: remove all those strings.ToLower calls
	for _, m := range members {
		var ok bool
		if m.SCMember != nil && icontains(m.SCMember.Username, q) {
			ok = true
		}
		if m.RosterEntry != nil {
			switch {
			case
				icontains(m.RosterEntry.PSNUsername, q),
				icontains(m.RosterEntry.RedditUsername, q),
				icontains(m.RosterEntry.SCUsername, q),
				icontains(m.RosterEntry.Nickname, q):
				ok = true
			}
		}
		if ok {
			matches = append(matches, m)
		}
	}
	return
}

func icontains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
