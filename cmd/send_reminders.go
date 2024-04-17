package cmd

import (
	"fmt"
	"log"
	"sync"
	"time"

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
					handler(event, c)
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
	Person    Person
	EventType EventType
}

type EventHandler func(Event, *Config)

func SlackAnniversaryChannelHandler(e Event, c *Config) {
	yearsInCompany := time.Now().Year() - e.Person.JoinDate.Year()
	anniversaryWishes := fmt.Sprintf(
		c.Slack.AnniversaryChannelReminder.MessageTemplate,
		e.Person.SlackMemberID,
		yearsInCompany,
	)
	if err := SendSlackChannelMsg(
		c.Slack.AnniversaryChannelReminder.ChannelName,
		anniversaryWishes,
		c,
	); err != nil {
		log.Println("Error when posting anniversary reminder:", err)
		return
	}
	log.Println("Sent anniversary info to channel for person", e.Person.SlackMemberID)
}

func SlackBirthdayReminderChannelHandler(e Event, c *Config) {
	if err := SendSlackChannelMsg(
		c.Slack.BirthdaysChannelReminder.ChannelName,
		fmt.Sprintf(c.Slack.BirthdaysChannelReminder.MessageTemplate, e.Person.SlackMemberID),
		c,
	); err != nil {
		log.Println("Error when posting birthday reminder:", err)
		return
	}
	log.Println("Sent birthday reminder to channel", e.Person.SlackMemberID)
}

func SlackBirthdayReminderDirectMessageHandler(e Event, c *Config) {
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

	if err := SendSlackDM(*e.Person.LeadSlackMemberID, msg, c); err != nil {
		log.Println("Error when sending DM remidner:", err)
		return
	}
	for _, slack_member_id := range c.Slack.BirthdaysDirectMessageReminder.AlwaysNotifySlackIds {
		if err := SendSlackDM(slack_member_id, msg, c); err != nil {
			log.Println("Error when sending DM remidner:", err)
			return
		}
	}
	log.Println("Sent birthday reminder Slack DM to lead", e.Person.SlackMemberID)
}

func SlackBirthdayPersonalReminderHandler(e Event, c *Config) {
	if e.Person.LeadSlackMemberID == nil {
		return
	}
	if err := SetSlackPersonalReminder(
		*e.Person.LeadSlackMemberID,
		c.Slack.BirthdaysPersonalReminder.Time,
		fmt.Sprintf(c.Slack.BirthdaysPersonalReminder.MessageTemplate, e.Person.SlackMemberID),
		c,
	); err != nil {
		log.Println("Error when posting Slack reminder:", err)
		return
	}
	log.Println("Set birthday Slack reminder for lead", *e.Person.LeadSlackMemberID)
}

func GetAnniversaryHandlers(c *Config) map[EventType][]EventHandler {
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
	p Person,
	todaysEvents chan Event,
	wg *sync.WaitGroup,
	c *Config,
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
