package cmd

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/nomysz/celebrations/config"
	"github.com/nomysz/celebrations/slack"
)

func SlackMonthlyReportHandler(e MonthlyReportEvent, c *config.Config, s slack.ChannelMessenger) {
	var textBirthdaysThisMonth,
		textAnniversariesThisMonth string

	sort.Slice(e.Birthdays, func(i, j int) bool {
		if e.Birthdays[i].BirthDate.Month() == e.Birthdays[j].BirthDate.Month() {
			return e.Birthdays[i].BirthDate.Day() < e.Birthdays[j].BirthDate.Day()
		}
		return e.Birthdays[i].BirthDate.Month() < e.Birthdays[j].BirthDate.Month()
	})

	sort.Slice(e.Anniversaries, func(i, j int) bool {
		if e.Anniversaries[i].JoinDate.Month() == e.Anniversaries[j].JoinDate.Month() {
			return e.Anniversaries[i].JoinDate.Day() < e.Anniversaries[j].JoinDate.Day()
		}
		return e.Anniversaries[i].JoinDate.Month() < e.Anniversaries[j].JoinDate.Month()
	})

	for _, p := range e.Birthdays {
		textBirthdaysThisMonth += fmt.Sprintf(
			"%s, <@%s> %d years old\n",
			p.BirthDate.Format("2 January"),
			p.SlackMemberID,
			getYearsPassedToCurrentYear(p.BirthDate),
		)
	}

	for _, p := range e.Anniversaries {
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
	if err := s.SendSlackChannelMsg(
		c.Slack.MonthlyReport.ChannelName,
		monthlyReport,
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

func getYearsPassedToCurrentYear(birthday time.Time) int {
	return GetNow().Year() - birthday.Year()
}

func SlackAnniversaryChannelHandler(e PersonalEvent, c *config.Config, s slack.ChannelMessenger) {
	if e.Person == nil {
		return
	}
	anniversaryWishes := fmt.Sprintf(
		c.Slack.AnniversaryChannelReminder.MessageTemplate,
		e.Person.SlackMemberID,
		getYearsText(e.Person.JoinDate),
	)
	if err := s.SendSlackChannelMsg(
		c.Slack.AnniversaryChannelReminder.ChannelName,
		anniversaryWishes,
	); err != nil {
		log.Println("Error when posting anniversary reminder:", err)
		return
	}
	log.Println("Sent anniversary info to channel for person", e.Person.SlackMemberID)
}

func SlackBirthdayReminderChannelHandler(e PersonalEvent, c *config.Config, s slack.ChannelMessenger) {
	if e.Person == nil {
		return
	}
	if err := s.SendSlackChannelMsg(
		c.Slack.BirthdaysChannelReminder.ChannelName,
		fmt.Sprintf(c.Slack.BirthdaysChannelReminder.MessageTemplate, e.Person.SlackMemberID),
	); err != nil {
		log.Println("Error when posting birthday reminder:", err)
		return
	}
	log.Println("Sent birthday reminder to channel", e.Person.SlackMemberID)
}

func SlackBirthdayReminderDirectMessageHandler(e PersonalEvent, c *config.Config, s slack.DirectMessenger) {
	if e.Person == nil {
		return
	}
	var msg string
	switch e.GetType() {
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
		log.Println("Error when sending DM remidner: Invalid EventType:", e.GetType())
		return
	}

	if err := s.SendSlackDM(*e.Person.LeadSlackMemberID, msg); err != nil {
		log.Println("Error when sending DM remidner:", err)
		return
	}
	for _, slackMemberID := range c.Slack.BirthdaysDirectMessageReminder.AlwaysNotifySlackIds {
		if err := s.SendSlackDM(slackMemberID, msg); err != nil {
			log.Println("Error when sending DM remidner:", err)
			return
		}
	}
	log.Println("Sent birthday reminder Slack DM to lead", e.Person.SlackMemberID)
}

func SlackBirthdayPersonalReminderHandler(e PersonalEvent, c *config.Config, s slack.PersonalReminderSetter) {
	if e.Person == nil {
		return
	}
	if e.Person.LeadSlackMemberID == nil {
		return
	}
	if err := s.SetSlackPersonalReminder(
		*e.Person.LeadSlackMemberID,
		c.Slack.BirthdaysPersonalReminder.Time,
		fmt.Sprintf(c.Slack.BirthdaysPersonalReminder.MessageTemplate, e.Person.SlackMemberID),
	); err != nil {
		log.Println("Error when posting Slack reminder:", err)
		return
	}
	log.Println("Set birthday Slack reminder for lead", *e.Person.LeadSlackMemberID)
}
