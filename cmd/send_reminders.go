package cmd

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nomysz/celebrations/config"
	"github.com/nomysz/celebrations/slack"
	"github.com/spf13/cobra"
)

var SendRemindersCmd = &cobra.Command{
	Use:   "send-reminders",
	Short: "Sending remidners using available handlers",
	Long:  "Sending remidners using available handlers",
	Run: func(cmd *cobra.Command, args []string) {
		SendReminders(
			config.GetConfig(),
			SlackClient{
				SlackChannelMsgSender:       slack.SendSlackChannelMsg,
				SlackDMSender:               slack.SendSlackDM,
				SlackPersonalReminderSetter: slack.SetSlackPersonalReminder,
			},
		)
	},
}

type SlackChannelMsgSender func(channel string, msg string, botToken string) error
type SlackDMSender func(slackId string, msg string, botToken string) error
type SlackPersonalReminderSetter func(slackId string, time string, msg string, userToken string) error

type SlackClient struct {
	SlackChannelMsgSender
	SlackDMSender
	SlackPersonalReminderSetter
}

func SendReminders(c *config.Config, sc SlackClient) {
	log.Println(len(c.People), "people found in config.")

	todaysEventsCh := make(chan Event)

	var wg sync.WaitGroup

	for _, p := range c.People {
		wg.Add(1)
		go GetTodaysEvents(p, todaysEventsCh, &wg, c)
	}

	go func() {
		wg.Wait()
		close(todaysEventsCh)
	}()

	for event := range todaysEventsCh {
		for event_type, handlers := range GetAnniversaryHandlers(c) {
			if event.EventType == event_type {
				for _, handler := range handlers {
					handler(event, c, sc)
				}
			}
		}
	}
}

type EventType uint16

const (
	Anniversary EventType = iota
	Birthday
	UpcomingBirthday
)

type Event struct {
	Person    config.Person
	EventType EventType
}

type EventHandler func(Event, *config.Config, SlackClient)

func SlackAnniversaryChannelHandler(e Event, c *config.Config, sc SlackClient) {
	yearsInCompany := time.Now().Year() - e.Person.JoinDate.Year()
	anniversaryWishes := fmt.Sprintf(
		c.Slack.AnniversaryChannelReminder.MessageTemplate,
		e.Person.SlackMemberID,
		yearsInCompany,
	)
	if err := sc.SlackChannelMsgSender(
		c.Slack.AnniversaryChannelReminder.ChannelName,
		anniversaryWishes,
		c.Slack.BotToken,
	); err != nil {
		log.Println("Error when posting anniversary reminder:", err)
		return
	}
	log.Println("Sent anniversary info to channel for person", e.Person.SlackMemberID)
}

func SlackBirthdayReminderChannelHandler(e Event, c *config.Config, sc SlackClient) {
	if err := sc.SlackChannelMsgSender(
		c.Slack.BirthdaysChannelReminder.ChannelName,
		fmt.Sprintf(c.Slack.BirthdaysChannelReminder.MessageTemplate, e.Person.SlackMemberID),
		c.Slack.BotToken,
	); err != nil {
		log.Println("Error when posting birthday reminder:", err)
		return
	}
	log.Println("Sent birthday reminder to channel", e.Person.SlackMemberID)
}

func SlackBirthdayReminderDirectMessageHandler(e Event, c *config.Config, sc SlackClient) {
	var msg string
	switch e.EventType {
	case Birthday:
		msg = fmt.Sprintf(
			c.Slack.BirthdaysDirectMessageReminder.MessageTemplate,
			e.Person.SlackMemberID,
		)
	case UpcomingBirthday:
		msg = fmt.Sprintf(
			c.Slack.BirthdaysDirectMessageReminder.PreRemidnerMessageTemplate,
			e.Person.SlackMemberID,
		)
	default:
		log.Println("Error when sending DM remidner: Invalid EventType:", e.EventType)
		return
	}

	if err := sc.SlackDMSender(*e.Person.LeadSlackMemberID, msg, c.Slack.BotToken); err != nil {
		log.Println("Error when sending DM remidner:", err)
		return
	}
	for _, slack_member_id := range c.Slack.BirthdaysDirectMessageReminder.AlwaysNotifySlackIds {
		if err := sc.SlackDMSender(slack_member_id, msg, c.Slack.BotToken); err != nil {
			log.Println("Error when sending DM remidner:", err)
			return
		}
	}
	log.Println("Sent birthday reminder Slack DM to lead", e.Person.SlackMemberID)
}

func SlackBirthdayPersonalReminderHandler(e Event, c *config.Config, sc SlackClient) {
	if e.Person.LeadSlackMemberID == nil {
		return
	}
	if err := sc.SlackPersonalReminderSetter(
		*e.Person.LeadSlackMemberID,
		c.Slack.BirthdaysPersonalReminder.Time,
		fmt.Sprintf(c.Slack.BirthdaysPersonalReminder.MessageTemplate, e.Person.SlackMemberID),
		c.Slack.UserToken,
	); err != nil {
		log.Println("Error when posting Slack reminder:", err)
		return
	}
	log.Println("Set birthday Slack reminder for lead", *e.Person.LeadSlackMemberID)
}

func GetAnniversaryHandlers(c *config.Config) map[EventType][]EventHandler {
	anniversaryHandlers := []EventHandler{}
	if c.Slack.AnniversaryChannelReminder.Enabled {
		anniversaryHandlers = append(anniversaryHandlers, SlackAnniversaryChannelHandler)
	}

	birthdayHandlers := []EventHandler{}
	upcomingBirthdayHandlers := []EventHandler{}
	if c.Slack.BirthdaysChannelReminder.Enabled {
		birthdayHandlers = append(birthdayHandlers, SlackBirthdayReminderChannelHandler)
	}
	if c.Slack.BirthdaysPersonalReminder.Enabled {
		birthdayHandlers = append(birthdayHandlers, SlackBirthdayPersonalReminderHandler)
	}
	if c.Slack.BirthdaysDirectMessageReminder.Enabled {
		birthdayHandlers = append(birthdayHandlers, SlackBirthdayReminderDirectMessageHandler)
		upcomingBirthdayHandlers = append(upcomingBirthdayHandlers, SlackBirthdayReminderDirectMessageHandler)
	}

	return map[EventType][]EventHandler{
		Anniversary:      anniversaryHandlers,
		Birthday:         birthdayHandlers,
		UpcomingBirthday: upcomingBirthdayHandlers,
	}
}

func IsTodayAnAnniversary(t time.Time) bool {
	ct := time.Now()
	return ct.Day() == t.Day() && ct.Month() == t.Month()
}

func GetTodaysEvents(
	p config.Person,
	todaysEvents chan Event,
	wg *sync.WaitGroup,
	c *config.Config,
) {
	defer wg.Done()
	if IsTodayAnAnniversary(
		p.BirthDate.Add(
			time.Hour * 24 * time.Duration(c.Slack.BirthdaysDirectMessageReminder.PreReminderDaysBefore),
		),
	) {
		todaysEvents <- Event{
			Person:    p,
			EventType: UpcomingBirthday,
		}
	}
	if IsTodayAnAnniversary(p.BirthDate) {
		todaysEvents <- Event{
			Person:    p,
			EventType: Birthday,
		}
	}
	if IsTodayAnAnniversary(p.JoinDate) {
		todaysEvents <- Event{
			Person:    p,
			EventType: Anniversary,
		}
	}
}
