package anna

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/time/rate"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"

	"github.com/sprt/anna/services/psn"
	"github.com/sprt/anna/services/roster"
	"github.com/sprt/anna/services/socialclub"
)

type Config struct {
	CommandPrefix string

	OwnerID          string
	OTChannelID      string
	FriendRequestMsg string

	DiscordEmail, DiscordPassword, DiscordToken string

	SocialClubCrewID    int
	GoogleDriveRosterID string

	PSNClientID, PSNClientSecret       string
	PSNDuid                            string
	PSNUsername, PSNEmail, PSNPassword string
}

type Bot struct {
	email, password, token string
	cmdPrefix              string
	DB                     *bolt.DB
	Session                *discordgo.Session
	user                   *discordgo.User
	commands               []*command
	tasks                  []*task

	OwnerID          string
	otChannelID      string
	FriendRequestMsg string

	PSN        *psn.Client
	Roster     *roster.Client
	SocialClub *socialclub.Client
}

func NewBot(config *Config, db *bolt.DB) *Bot {
	psnConfig := &psn.Config{
		Username:     config.PSNUsername,
		Email:        config.PSNEmail,
		Password:     config.PSNPassword,
		ClientID:     config.PSNClientID,
		ClientSecret: config.PSNClientSecret,
		DUID:         config.PSNDuid,
	}
	psnTS := oauth2.ReuseTokenSource(nil, psn.NewTokenSource(psnConfig, nil))
	scRL := rate.NewLimiter(rate.Every(30*time.Second), 1)
	return &Bot{
		email:     config.DiscordEmail,
		password:  config.DiscordPassword,
		token:     config.DiscordToken,
		cmdPrefix: config.CommandPrefix,
		DB:        db,

		OwnerID:          config.OwnerID,
		otChannelID:      config.OTChannelID,
		FriendRequestMsg: config.FriendRequestMsg,

		PSN:        psn.NewClient(psnConfig, oauth2.NewClient(oauth2.NoContext, psnTS), nil),
		Roster:     roster.NewClient(config.GoogleDriveRosterID, nil, nil),
		SocialClub: socialclub.NewClient(&socialclub.Config{CrewID: config.SocialClubCrewID}, nil, scRL),
	}
}

func (b *Bot) Start() error {
	if err := b.DB.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("default")); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte("friended")); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	session, err := discordgo.New(b.email, b.password, b.token)
	if err != nil {
		return err
	}

	user, err := session.User("@me")
	if err != nil {
		return err
	}

	err = session.Open()
	if err != nil {
		return err
	}

	b.Session = session
	b.user = user

	b.Session.AddHandler(b.onReady)
	b.Session.AddHandler(b.onMessageCreate)
	b.Session.AddHandler(b.onGuildMemberAdd)
	b.Session.AddHandler(b.onPresenceUpdate)

	return nil
}

func (b *Bot) RegisterCommand(name string, fn func(*Bot, *discordgo.Message, []string) error) {
	// FIXME: panic if a task with the same name already exists
	b.commands = append(b.commands, &command{
		name: name,
		fn:   fn,
	})
}

func (b *Bot) RegisterTask(fn func(*Bot) error, sleep time.Duration) {
	b.tasks = append(b.tasks, &task{
		fn:    fn,
		sleep: sleep,
	})
}

func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	for _, task := range b.tasks {
		task := task
		go func() {
			for {
				time.Sleep(task.sleep)
				log.Print("Starting task...")
				if err := task.fn(b); err != nil {
					log.Printf("ERROR: task: %s", err)
					continue
				}
				log.Printf("Task done, next run in %s", task.sleep)
			}
		}()
	}
}

func (b *Bot) onGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	say := fmt.Sprintf("Welcome <@!%s>! Don't forget to set your nickname to your PSN.", m.User.ID)
	_, err := s.ChannelMessageSend(b.otChannelID, say)
	if err != nil {
		log.Println("ERROR onGuildMemberAdd:", err)
	}
}

func (bot *Bot) onPresenceUpdate(s *discordgo.Session, m *discordgo.PresenceUpdate) {
	// TODO: trigger peak check on ready
	if m.Status == "" || m.Status == "offline" {
		return
	}

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		log.Println(err)
		return
	}
	var current int
	for _, pres := range guild.Presences {
		if pres.Status != "offline" {
			current++
		}
	}

	var newPeak *Peak
	var lastPeak *Peak
	if err := bot.DB.Update(func(tx *bolt.Tx) error {
		buk := tx.Bucket([]byte("default"))
		b := buk.Get([]byte("peak"))
		peak := new(Peak)
		if b != nil {
			lastPeak = peak
			buf := bytes.NewBuffer(b)
			if err := gob.NewDecoder(buf).Decode(peak); err != nil {
				return err
			}
		}
		if current > peak.Count {
			newPeak = &Peak{Count: current, Date: time.Now().UTC()}
			var b bytes.Buffer
			if err := gob.NewEncoder(&b).Encode(newPeak); err != nil {
				return err
			}
			return buk.Put([]byte("peak"), b.Bytes())
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	if newPeak != nil {
		say := fmt.Sprintf("New peak: %d people online!", newPeak.Count)
		if lastPeak != nil {
			say += fmt.Sprintf(" (last peak: %d ppl @ %s)", lastPeak.Count,
				lastPeak.Date.Format("Jan 2 2006 15:04 MST"))
		}
		_, err := s.ChannelMessageSend(bot.otChannelID, say)
		if err != nil {
			log.Println(err)
		}
	}
}

func (b *Bot) onMessageCreate(s *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.Author.ID == b.user.ID {
		return
	}

	// XXX: don't allow spaces between prefix and command name
	if strings.HasPrefix(mc.Content, b.cmdPrefix) {
		args := strings.Fields(mc.Content[len(b.cmdPrefix):])
		b.onCommand(mc.Message, args[0], args[1:])
		return
	}
}

func (b *Bot) onCommand(message *discordgo.Message, name string, args []string) {
	for _, cmd := range b.commands {
		if strings.EqualFold(cmd.name, name) {
			err := cmd.fn(b, message, args)
			if err != nil {
				log.Printf("ERROR: %s: %s", cmd.name, err)
			}
			break
		}
	}
}

type command struct {
	name string
	fn   func(*Bot, *discordgo.Message, []string) error
}

type task struct {
	fn    func(*Bot) error
	sleep time.Duration
}

type Peak struct {
	Count int
	Date  time.Time
}
