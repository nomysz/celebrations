package cmd

import (
	"log"
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
		cfg := config.GetConfig()
		SendReminders(
			cfg,
			slack.NewClient(cfg.Slack.BotToken, cfg.Slack.UserToken),
		)
	},
}

type EventType uint16

const (
	Anniversary EventType = iota
	Birthday
	UpcomingBirthday
	MonthlyReportDay
)

type Event interface {
	GetType() EventType
}

type PersonalEvent struct {
	Type   EventType
	Person config.Person
}

func (e PersonalEvent) GetType() EventType {
	return e.Type
}

type MonthlyReportEvent struct {
	Type          EventType
	Birthdays     []config.Person
	Anniversaries []config.Person
}

func (e MonthlyReportEvent) GetType() EventType {
	return e.Type
}

func SendReminders(c *config.Config, sc slack.SlackCommunicator) {
	log.Println(len(c.People), "people found in config.")

	var todaysEvents []Event

	for _, p := range c.People {
		for e := range GetTodaysEventsForPerson(p, c) {
			todaysEvents = append(todaysEvents, e)
		}
	}

	if GetNow().Day() == 1 && c.Slack.MonthlyReport.Enabled {
		todaysEvents = append(todaysEvents, GetMonthlyReportEvent(c.People))
	}

	for _, e := range todaysEvents {
		if pe, ok := e.(PersonalEvent); ok {
			switch e.GetType() {
			case Anniversary:
				if c.Slack.AnniversaryChannelReminder.Enabled {
					SlackAnniversaryChannelHandler(pe, c, sc)
				}
			case Birthday:
				if c.Slack.BirthdaysChannelReminder.Enabled {
					SlackBirthdayReminderChannelHandler(pe, c, sc)
				}
				if c.Slack.BirthdaysDirectMessageReminder.Enabled {
					SlackBirthdayReminderDirectMessageHandler(pe, c, sc)
				}
				if c.Slack.BirthdaysPersonalReminder.Enabled {
					SlackBirthdayPersonalReminderHandler(pe, c, sc)
				}
			case UpcomingBirthday:
				if c.Slack.BirthdaysDirectMessageReminder.Enabled {
					SlackBirthdayReminderDirectMessageHandler(pe, c, sc)
				}
			}
		} else if me, ok := e.(MonthlyReportEvent); ok {
			if c.Slack.MonthlyReport.Enabled {
				SlackMonthlyReportHandler(me, c, sc)
			}
		} else {
			panic("Unknown type of event to handle")
		}
	}
}

func GetMonthlyReportEvent(p []config.Person) MonthlyReportEvent {
	var birthdaysThisMonth,
		anniversariesThisMonth []config.Person

	currentMonth := GetNow().Month()

	for _, p := range p {
		if p.BirthDate.Month() == currentMonth {
			birthdaysThisMonth = append(birthdaysThisMonth, p)
		}
		if p.JoinDate.Month() == currentMonth {
			anniversariesThisMonth = append(anniversariesThisMonth, p)
		}
	}

	return MonthlyReportEvent{
		Type:          MonthlyReportDay,
		Birthdays:     birthdaysThisMonth,
		Anniversaries: anniversariesThisMonth,
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
			ch <- PersonalEvent{
				Type:   UpcomingBirthday,
				Person: p,
			}
		}
		if DayAndMonthMatch(p.BirthDate) {
			ch <- PersonalEvent{
				Type:   Birthday,
				Person: p,
			}
		}
		if DayAndMonthMatch(p.JoinDate) {
			ch <- PersonalEvent{
				Type:   Anniversary,
				Person: p,
			}
		}
	}()
	return ch
}

func DayAndMonthMatch(t time.Time) bool {
	ct := GetNow()
	return ct.Day() == t.Day() && ct.Month() == t.Month()
}
