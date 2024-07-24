package cmd

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/nomysz/celebrations/config"
	"github.com/nomysz/celebrations/slack"
	"github.com/spf13/cobra"
)

var SendRemindersCmd = &cobra.Command{
	Use:   "send-reminders",
	Short: "Send remidners via configured handlers",
	Long:  "Send remidners via configured handlers",
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

type SlackClient struct {
	SlackChannelMsgSender       func(channel string, msg string, botToken string) error
	SlackDMSender               func(slackId string, msg string, botToken string) error
	SlackPersonalReminderSetter func(slackId string, time string, msg string, userToken string) error
}

func SendReminders(c *config.Config, sc SlackClient) {
	log.Println(len(c.People), "people found in config.")

	var todaysEvents []Event

	for _, p := range c.People {
		for e := range GetTodaysEventsForPerson(p, c) {
			todaysEvents = append(todaysEvents, e)
		}
	}

	if GetNow().Day() == 1 && c.Slack.MonthlyReport.Enabled {
		todaysEvents = append(todaysEvents, GetMonthlyReportEvent(c))
	}

	for _, e := range todaysEvents {
		for handlerEventType, handlers := range GetEventHandlers(c) {
			if e.EventType == handlerEventType {
				for _, handler := range handlers {
					handler(e, c, sc)
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
	MonthlyReportDay
)

type Event struct {
	EventType EventType

	Person *config.Person

	BirthdaysThisMonth     []config.Person
	AnniversariesThisMonth []config.Person
}

type EventHandler func(Event, *config.Config, SlackClient)

func GetMonthlyReportEvent(c *config.Config) Event {
	var birthdaysThisMonth,
		anniversariesThisMonth []config.Person

	currentMonth := GetNow().Month()

	for _, p := range c.People {
		if p.BirthDate.Month() == currentMonth {
			birthdaysThisMonth = append(birthdaysThisMonth, p)
		}
		if p.JoinDate.Month() == currentMonth {
			anniversariesThisMonth = append(anniversariesThisMonth, p)
		}
	}

	return Event{
		EventType:              MonthlyReportDay,
		BirthdaysThisMonth:     birthdaysThisMonth,
		AnniversariesThisMonth: anniversariesThisMonth,
	}
}

func GetTodaysEventsForPerson(
	p config.Person,
	c *config.Config,
) <-chan Event {
	ch := make(chan Event)
	go func() {
		defer close(ch)
		if DayAndMonthMatch(
			p.BirthDate.Add(
				-time.Hour * 24 * time.Duration(c.Slack.BirthdaysDirectMessageReminder.PreReminderDaysBefore),
			),
		) {
			ch <- Event{
				EventType: UpcomingBirthday,
				Person:    &p,
			}
		}
		if DayAndMonthMatch(p.BirthDate) {
			ch <- Event{
				EventType: Birthday,
				Person:    &p,
			}
		}
		if DayAndMonthMatch(p.JoinDate) {
			ch <- Event{
				Person:    &p,
				EventType: Anniversary,
			}
		}
	}()
	return ch
}

func DayAndMonthMatch(t time.Time) bool {
	ct := GetNow()
	return ct.Day() == t.Day() && ct.Month() == t.Month()
}

func GetEventHandlers(c *config.Config) map[EventType][]EventHandler {
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

	monthlyReportHandlers := []EventHandler{}
	if c.Slack.MonthlyReport.Enabled {
		monthlyReportHandlers = append(monthlyReportHandlers, SlackMonthlyReportHandler)
	}

	return map[EventType][]EventHandler{
		Anniversary:      anniversaryHandlers,
		Birthday:         birthdayHandlers,
		UpcomingBirthday: upcomingBirthdayHandlers,
		MonthlyReportDay: monthlyReportHandlers,
	}
}

func getYearsPassedToCurrentYear(birthday time.Time) int {
	return GetNow().Year() - birthday.Year()
}

func SlackMonthlyReportHandler(e Event, c *config.Config, sc SlackClient) {
	var textBirthdaysThisMonth,
		textAnniversariesThisMonth string

	sort.Slice(e.BirthdaysThisMonth, func(i, j int) bool {
		if e.BirthdaysThisMonth[i].BirthDate.Month() == e.BirthdaysThisMonth[j].BirthDate.Month() {
			return e.BirthdaysThisMonth[i].BirthDate.Day() < e.BirthdaysThisMonth[j].BirthDate.Day()
		}
		return e.BirthdaysThisMonth[i].BirthDate.Month() < e.BirthdaysThisMonth[j].BirthDate.Month()
	})

	sort.Slice(e.AnniversariesThisMonth, func(i, j int) bool {
		if e.AnniversariesThisMonth[i].JoinDate.Month() == e.AnniversariesThisMonth[j].JoinDate.Month() {
			return e.AnniversariesThisMonth[i].JoinDate.Day() < e.AnniversariesThisMonth[j].JoinDate.Day()
		}
		return e.AnniversariesThisMonth[i].JoinDate.Month() < e.AnniversariesThisMonth[j].JoinDate.Month()
	})

	for _, p := range e.BirthdaysThisMonth {
		textBirthdaysThisMonth += fmt.Sprintf(
			"%s, <@%s> %d years old\n",
			p.BirthDate.Format("2 January"),
			p.SlackMemberID,
			getYearsPassedToCurrentYear(p.BirthDate),
		)
	}

	for _, p := range e.AnniversariesThisMonth {
		textAnniversariesThisMonth += fmt.Sprintf(
			"%s, <@%s> %s in company\n",
			p.JoinDate.Format("2 January"),
			p.SlackMemberID,
			getYearsText(p.JoinDate),
		)
	}

	monthlyReport := fmt.Sprintf(
		c.Slack.MonthlyReport.MessageTemplate,
		textBirthdaysThisMonth,
		textAnniversariesThisMonth,
	)
	if err := sc.SlackChannelMsgSender(
		c.Slack.MonthlyReport.ChannelName,
		monthlyReport,
		c.Slack.BotToken,
	); err != nil {
		log.Println("Error when posting monthly report reminder:", err)
		return
	}
	log.Println("Sent monthly report to channel", c.Slack.MonthlyReport.ChannelName)
}

func getYearsText(date time.Time) string {
	yearsInCompany := getYearsPassedToCurrentYear(date)
	if yearsInCompany > 1 {
		return fmt.Sprintf("%d years", yearsInCompany)
	}
	return "1 year"
}

func SlackAnniversaryChannelHandler(e Event, c *config.Config, sc SlackClient) {
	if e.Person == nil {
		return
	}

	anniversaryWishes := fmt.Sprintf(
		c.Slack.AnniversaryChannelReminder.MessageTemplate,
		e.Person.SlackMemberID,
		getYearsText(e.Person.JoinDate),
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
	if e.Person == nil {
		return
	}
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
	if e.Person == nil {
		return
	}

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
			c.Slack.BirthdaysDirectMessageReminder.PreReminderDaysBefore,
		)
	default:
		log.Println("Error when sending DM remidner: Invalid EventType:", e.EventType)
		return
	}

	if err := sc.SlackDMSender(*e.Person.LeadSlackMemberID, msg, c.Slack.BotToken); err != nil {
		log.Println("Error when sending DM remidner:", err)
		return
	}
	for _, slackMemberID := range c.Slack.BirthdaysDirectMessageReminder.AlwaysNotifySlackIds {
		if err := sc.SlackDMSender(slackMemberID, msg, c.Slack.BotToken); err != nil {
			log.Println("Error when sending DM remidner:", err)
			return
		}
	}
	log.Println("Sent birthday reminder Slack DM to lead", e.Person.SlackMemberID)
}

func SlackBirthdayPersonalReminderHandler(e Event, c *config.Config, sc SlackClient) {
	if e.Person == nil {
		return
	}
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
