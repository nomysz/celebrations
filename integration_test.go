package main

import (
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"github.com/nomysz/celebrations/cmd"
	"github.com/nomysz/celebrations/config"
	"github.com/stretchr/testify/assert"
)

func getTestConfig() *config.Config {
	personWithTodaysBirthdayDateSlackID := "birthday-slack-id"
	personWithTodaysAnniversaryDateSlackID := "anniversary-slack-id"
	personWithAllDatesMoreThanMonthAgo := "more-than-month-ago-slack-id"
	leaderSlackID := "leader-slack-id"
	alwaysInformedLeaderSlackID := "leader-always-informed-slack-id"
	var daysBefore int64 = 3

	return &config.Config{
		Slack: config.Slack{
			BotToken:  "bot-token",
			UserToken: "user-token",
			AnniversaryChannelReminder: config.AnniversaryChannelReminder{
				Enabled:         true,
				ChannelName:     "celebrations",
				MessageTemplate: "Happy anniversary <@%s>! %d years in Company!",
			},
			BirthdaysChannelReminder: config.BirthdaysChannelReminder{
				Enabled:         true,
				ChannelName:     "leaders",
				MessageTemplate: "<@%s> is having birthday!",
			},
			BirthdaysPersonalReminder: config.BirthdaysPersonalReminder{
				Enabled:         true,
				Time:            "15pm",
				MessageTemplate: "<@%s> is having birthday!",
			},
			BirthdaysDirectMessageReminder: config.BirthdaysDirectMessageReminder{
				Enabled:                    true,
				MessageTemplate:            "<@%s> is having birthday!",
				PreReminderDaysBefore:      daysBefore,
				PreRemidnerMessageTemplate: "<@%s> is having birthday in " + fmt.Sprint(daysBefore) + " days!",
				AlwaysNotifySlackIds: []string{
					alwaysInformedLeaderSlackID,
				},
			},
			DownloadingUsers: config.DownloadingUsers{
				BirthdayCustomFieldName: "dummy-birthdate-field-name",
				JoinDateCustomFieldName: "dummy-joindate-field-name",
			},
		},
		People: []config.Person{
			{
				SlackMemberID:     personWithTodaysBirthdayDateSlackID,
				BirthDate:         time.Now(),
				JoinDate:          time.Now().AddDate(0, 0, -4),
				LeadSlackMemberID: &leaderSlackID,
			},
			{
				SlackMemberID:     personWithTodaysAnniversaryDateSlackID,
				BirthDate:         time.Now().AddDate(0, 0, -6),
				JoinDate:          time.Now().AddDate(-2, 0, 0),
				LeadSlackMemberID: &alwaysInformedLeaderSlackID,
			},
			{
				SlackMemberID:     personWithAllDatesMoreThanMonthAgo,
				BirthDate:         time.Now().AddDate(0, -2, 0),
				JoinDate:          time.Now().AddDate(0, -2, 0),
				LeadSlackMemberID: &leaderSlackID,
			},
		},
	}
}

func TestSendReminders(t *testing.T) {
	log.SetOutput(io.Discard)

	messages := []string{}

	cmd.SendReminders(
		getTestConfig(),
		cmd.SlackClient{
			SlackChannelMsgSender: func(channel string, msg string, botToken string) error {
				messages = append(
					messages,
					fmt.Sprintf("SENDING '%s' TO CHANNEL '%s' USING TOKEN %s", msg, channel, botToken),
				)
				return nil
			},
			SlackDMSender: func(slackId string, msg string, botToken string) error {
				messages = append(
					messages,
					fmt.Sprintf("SENDING DM '%s' TO '%s' USING TOKEN %s", msg, slackId, botToken),
				)
				return nil
			},
			SlackPersonalReminderSetter: func(slackId string, time string, msg string, userToken string) error {
				messages = append(
					messages,
					fmt.Sprintf("SETTING REMINDER '%s' AT '%s' TO '%s' USING TOKEN %s", msg, slackId, time, userToken),
				)
				return nil
			},
		},
	)

	assert.NotEmpty(t, messages)

	assert.Contains(t, messages,
		"SENDING '<@birthday-slack-id> is having birthday!' TO CHANNEL 'leaders' USING TOKEN bot-token")

	assert.Contains(t, messages,
		"SETTING REMINDER '<@birthday-slack-id> is having birthday!' AT 'leader-slack-id' TO '15pm' USING TOKEN user-token")

	assert.Contains(t, messages,
		"SENDING 'Happy anniversary <@anniversary-slack-id>! 2 years in Company!' TO CHANNEL 'celebrations' USING TOKEN bot-token")

	assert.Contains(t, messages,
		"SENDING DM '<@birthday-slack-id> is having birthday!' TO 'leader-slack-id' USING TOKEN bot-token")
	assert.Contains(t, messages,
		"SENDING DM '<@birthday-slack-id> is having birthday!' TO 'leader-always-informed-slack-id' USING TOKEN bot-token")
}
