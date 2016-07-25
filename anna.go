package anna

import (
	"log"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"

	"github.com/sprt/anna/services/roster"
	"github.com/sprt/anna/services/socialclub"
)

type Member struct {
	// either fields can be nil (exclusively)
	RosterEntry *roster.Entry
	SCMember    *socialclub.Member
}

type Bot struct {
	email, password, token string
	cmdPrefix              string
	DB                     *bolt.DB
	Session                *discordgo.Session
	user, owner            *discordgo.User
	commands               []*command
	tasks                  []*task

	Roster     *roster.Client
	SocialClub *socialclub.Client
}

func NewBot(db *bolt.DB) *Bot {
	return &Bot{
		email:     Config.DiscordEmail,
		password:  Config.DiscordPassword,
		token:     Config.DiscordToken,
		cmdPrefix: Config.CommandPrefix,
		DB:        db,

		Roster:     roster.NewClient(Config.GoogleDriveRosterID, nil, nil),
		SocialClub: socialclub.NewClient(&socialclub.Config{CrewID: Config.SocialClubCrewID}, nil, nil),
	}
}

func (b *Bot) Start() error {
	if err := b.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("default"))
		if err != nil {
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
	// TODO: initialize owner

	b.Session.AddHandler(b.onReady)
	b.Session.AddHandler(b.onMessageCreate)

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
		go func() {
			for {
				time.Sleep(task.sleep)
				err := task.fn(b)
				if err != nil {
					log.Printf("ERROR: task: %s", err)
				}
			}
		}()
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
