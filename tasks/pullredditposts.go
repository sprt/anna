package tasks

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jzelinskie/geddit"
	"github.com/sprt/anna"
)

var lastPost time.Time

func PullRedditPosts(bot *anna.Bot) error {
	r := geddit.NewSession("")
	posts, err := r.SubredditSubmissions(bot.Subreddit, geddit.NewSubmissions, geddit.ListingOptions{})
	if err != nil {
		return err
	}
mainloop:
	for i := len(posts) - 1; i >= 0; i-- {
		p := posts[i]
		if time.Now().Sub(time.Unix(int64(p.DateCreated), 0)) >= time.Minute {
			continue
		}
		for _, domain := range bot.DomainBlacklist {
			// XXX: check dot boundaries
			if strings.HasSuffix(p.Domain, domain) {
				continue mainloop
			}
		}
		for _, user := range bot.UserBlacklist {
			if p.Author == user {
				continue mainloop
			}
		}
		say := fmt.Sprintf("/r/GTAA: <%s> %s https://redd.it/%s",
			escape(p.Author), escape(p.Title), p.ID)
		if !p.IsSelf {
			say += fmt.Sprintf(" [%s]", p.URL)
		}
		_, err := bot.Session.ChannelMessageSend(bot.OTChannelID, say)
		if err != nil {
			log.Println(err)
			continue
		}
	}
	return nil
}

func escape(s string) string {
	for _, c := range []string{"_", "*", "`", "~"} {
		s = strings.Replace(s, c, "\\"+c, -1)
	}
	return s
}
