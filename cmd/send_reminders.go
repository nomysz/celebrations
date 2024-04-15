package cmd

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/slack-go/slack"
	"github.com/spf13/cobra"
)

var send_reminders = &cobra.Command{
	Use:   "send-reminders",
	Short: "Sending remidners using available handlers",
	Long:  "Sending remidners using available handlers",
	Run:   func(cmd *cobra.Command, args []string) { sendReminders() },
}

func sendReminders() {
	c := getConfig()

	log.Println(len(c.People), "people found in config.")

	todaysEventsCh := make(chan Event)

	var wg sync.WaitGroup

	for _, p := range c.People {
		wg.Add(1)
		go GetTodaysEvents(p, todaysEventsCh, &wg)
	}

	go func() {
		wg.Wait()
		close(todaysEventsCh)
	}()

	for event := range todaysEventsCh {
		for event_type, handlers := range getAnniversaryHandlers(c) {
			if event.EventType == event_type {
				for _, handler := range handlers {
					handler(event.Person, c)
				}
			}
		}
	}
}

type EventType uint16

const (
	Anniversary EventType = iota
	Birthday
)

type Event struct {
	Person    Person
	EventType EventType
}

type EventHandler func(Person, *Config)

func SlackAnniversaryChannelHandler(p Person, c *Config) {
	yearsInCompany := time.Now().Year() - p.JoinDate.Year()
	anniversaryWishes := fmt.Sprintf(
		c.Slack.AnniversaryChannelReminder.MessageTemplate,
		p.SlackMemberID,
		yearsInCompany,
	)

	_, _, err := slack.New(c.Slack.BotToken).PostMessage(
		c.Slack.AnniversaryChannelReminder.ChannelName,
		slack.MsgOptionAttachments(slack.Attachment{Pretext: anniversaryWishes}),
	)

	if err != nil {
		log.Println("Error when posting anniversary message", err)
		return
	}

	log.Println("Sent anniversary info to channel for person", p.SlackMemberID)
}

func SlackBirthdayReminderChannelHandler(p Person, c *Config) {
	_, _, err := slack.New(c.Slack.BotToken).PostMessage(
		c.Slack.BirthdaysChannelReminder.ChannelName,
		slack.MsgOptionAttachments(
			slack.Attachment{
				Pretext: fmt.Sprintf(c.Slack.BirthdaysChannelReminder.MessageTemplate, p.SlackMemberID),
			},
		),
	)

	if err != nil {
		log.Println("Error when posting birthday reminder message to channel", err)
		return
	}

	log.Println("Sent birthday reminder to channel", p.SlackMemberID)
}

func SlackBirthdayReminderDirectMessageHandler(p Person, c *Config) {
	api := slack.New(c.Slack.BotToken)

	slack_ch, _, _, err := api.OpenConversation(
		&slack.OpenConversationParameters{
			Users:    []string{*p.LeadSlackMemberID},
			ReturnIM: false,
		},
	)

	if err != nil {
		log.Println("Error when opening Slack conversation with lead", err)
	}

	_, _, err = api.PostMessage(
		slack_ch.ID,
		slack.MsgOptionText(
			fmt.Sprintf(c.Slack.BirthdaysDirectMessageReminder.MessageTemplate, p.SlackMemberID),
			false,
		),
	)

	if err != nil {
		log.Println("Error when sending birthday reminder Slack DM to lead", err)
		return
	}

	log.Println("Sent birthday reminder Slack DM to lead", p.SlackMemberID)
}

func SlackBirthdayPersonalReminderHandler(p Person, c *Config) {
	if p.LeadSlackMemberID == nil {
		return
	}

	_, err := slack.New(c.Slack.UserToken).AddUserReminder(
		*p.LeadSlackMemberID,
		fmt.Sprintf(c.Slack.BirthdaysPersonalReminder.MessageTemplate, p.SlackMemberID),
		c.Slack.BirthdaysPersonalReminder.Time,
	)
	if err != nil {
		log.Println("There was an error when posting Slack reminder", err)
		return
	}

	log.Println("Set birthday Slack reminder for lead", *p.LeadSlackMemberID)
}

func getAnniversaryHandlers(c *Config) map[EventType][]EventHandler {
	anniversaryHandlers := []EventHandler{}
	if c.Slack.AnniversaryChannelReminder.Enabled {
		anniversaryHandlers = append(anniversaryHandlers, SlackAnniversaryChannelHandler)
	}

	birthdayHandlers := []EventHandler{}
	if c.Slack.BirthdaysChannelReminder.Enabled {
		birthdayHandlers = append(birthdayHandlers, SlackBirthdayReminderChannelHandler)
	}
	if c.Slack.BirthdaysPersonalReminder.Enabled {
		birthdayHandlers = append(birthdayHandlers, SlackBirthdayPersonalReminderHandler)
	}
	if c.Slack.BirthdaysDirectMessageReminder.Enabled {
		birthdayHandlers = append(birthdayHandlers, SlackBirthdayReminderDirectMessageHandler)
	}

	return map[EventType][]EventHandler{
		Anniversary: anniversaryHandlers,
		Birthday:    birthdayHandlers,
	}
}

func IsAnniversaryDay(t *time.Time) bool {
	ct := time.Now()
	return ct.Day() == t.Day() && ct.Month() == t.Month()
}

func GetTodaysEvents(
	p Person,
	peopleToRemindCh chan Event,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	if IsAnniversaryDay(&p.BirthDate) {
		peopleToRemindCh <- Event{
			Person:    p,
			EventType: Birthday,
		}
	}
	if IsAnniversaryDay(&p.JoinDate) {
		peopleToRemindCh <- Event{
			Person:    p,
			EventType: Anniversary,
		}
	}
}
