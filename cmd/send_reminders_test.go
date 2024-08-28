package cmd

import (
	"fmt"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/nomysz/celebrations/config"
	"github.com/stretchr/testify/assert"
)

// Returns true if lookup string was found as partial within input []strings
func partialContains(strs []string, lookup string) bool {
	for _, x := range strs {
		if strings.Contains(x, lookup) {
			return true
		}
	}
	return false
}

func getOffsetNowDate(years int, months int, days int) time.Time {
	return GetNow().AddDate(years, months, days)
}

func getTestConfig() *config.Config {
	personWithTodaysBirthdayDateSlackID := "birthday-slack-id"
	personWithTodaysAnniversaryDateSlackID := "anniversary-slack-id"
	personWithThisMonthBirthdayDateSlackID := "monthly-report-birthday-slack-id"
	personWithThisMonthAnniversaryDateSlackID := "monthly-report-anniversary-slack-id"
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
				MessageTemplate: "Happy anniversary <@%s>! %s in Company!",
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
			MonthlyReport: config.MonthlyReport{
				Enabled:         true,
				ChannelName:     "leaders",
				MessageTemplate: "Birthdays:\n%s\nAnniversaries:\n%s",
			},
			DownloadingUsers: config.DownloadingUsers{
				BirthdayCustomFieldName: "dummy-birthdate-field-name",
				JoinDateCustomFieldName: "dummy-joindate-field-name",
			},
		},
		People: []config.Person{
			{
				SlackMemberID:     personWithTodaysBirthdayDateSlackID,
				BirthDate:         getOffsetNowDate(-22, 0, 0),
				JoinDate:          getOffsetNowDate(-5, 0, +4),
				LeadSlackMemberID: &leaderSlackID,
			},
			{
				SlackMemberID:     personWithTodaysAnniversaryDateSlackID,
				BirthDate:         getOffsetNowDate(-20, -3, 0),
				JoinDate:          getOffsetNowDate(-2, 0, 0),
				LeadSlackMemberID: &alwaysInformedLeaderSlackID,
			},
			{
				SlackMemberID:     personWithThisMonthBirthdayDateSlackID,
				BirthDate:         getOffsetNowDate(-30, 0, +10),
				JoinDate:          getOffsetNowDate(-1, -1, 0),
				LeadSlackMemberID: &leaderSlackID,
			},
			{
				SlackMemberID:     personWithThisMonthAnniversaryDateSlackID,
				BirthDate:         getOffsetNowDate(-24, -2, 0),
				JoinDate:          getOffsetNowDate(-1, 0, +20),
				LeadSlackMemberID: &leaderSlackID,
			},
			{
				SlackMemberID:     personWithAllDatesMoreThanMonthAgo,
				BirthDate:         getOffsetNowDate(-36, 0, -40),
				JoinDate:          getOffsetNowDate(-1, -2, 0),
				LeadSlackMemberID: &leaderSlackID,
			},
		},
	}
}

type TestSlackClient struct {
	botToken  string
	userToken string
	messages  []string
}

func (sc *TestSlackClient) SendSlackChannelMsg(channel string, msg string) error {
	sc.messages = append(
		sc.messages,
		fmt.Sprintf("SENDING '%s' TO CHANNEL '%s' USING TOKEN %s", msg, channel, sc.botToken),
	)
	return nil
}

func (sc *TestSlackClient) SendSlackDM(slackId string, msg string) error {
	sc.messages = append(
		sc.messages,
		fmt.Sprintf("SENDING DM '%s' TO '%s' USING TOKEN %s", msg, slackId, sc.botToken),
	)
	return nil
}

func (sc *TestSlackClient) SetSlackPersonalReminder(slackId string, time string, msg string) error {
	sc.messages = append(
		sc.messages,
		fmt.Sprintf("SETTING REMINDER '%s' AT '%s' TO '%s' USING TOKEN %s", msg, slackId, time, sc.userToken),
	)
	return nil
}

func TestSendReminders(t *testing.T) {
	log.SetOutput(io.Discard)

	GetNow = func() time.Time {
		return time.Date(2016, time.June, 1, 0, 0, 0, 0, time.UTC)
	}

	sc := TestSlackClient{
		botToken:  getTestConfig().Slack.BotToken,
		userToken: getTestConfig().Slack.UserToken,
		messages:  []string{},
	}

	SendReminders(
		getTestConfig(),
		&sc,
	)

	assert.NotEmpty(t, sc.messages)

	assert.Contains(t, sc.messages,
		"SENDING '<@birthday-slack-id> is having birthday!' TO CHANNEL 'leaders' USING TOKEN bot-token",
		"Error in birthday channel msg")

	assert.Contains(t, sc.messages,
		"SETTING REMINDER '<@birthday-slack-id> is having birthday!' AT 'leader-slack-id' TO '15pm' USING TOKEN user-token",
		"Error in personal reminder")

	assert.Contains(t, sc.messages,
		"SENDING 'Happy anniversary <@anniversary-slack-id>! 2 years in Company!' TO CHANNEL 'celebrations' USING TOKEN bot-token",
		"Error in anniversary channel msg")

	assert.Contains(t, sc.messages,
		"SENDING DM '<@birthday-slack-id> is having birthday!' TO 'leader-slack-id' USING TOKEN bot-token",
		"Error in DM")
	assert.Contains(t, sc.messages,
		"SENDING DM '<@birthday-slack-id> is having birthday!' TO 'leader-always-informed-slack-id' USING TOKEN bot-token",
		"Error in DM")

	assert.True(t, partialContains(sc.messages, "Birthdays:"), "Error in monthly report")
	assert.True(t, partialContains(sc.messages, "1 June, <@birthday-slack-id> 22 years old"),
		"Error in monthly report")
	assert.True(t, partialContains(sc.messages, "11 June, <@monthly-report-birthday-slack-id> 30 years old"),
		"Error in monthly report")
	assert.True(t, partialContains(sc.messages, "Anniversaries:"), "Error in monthly report")
	assert.True(t, partialContains(sc.messages, "5 June, <@birthday-slack-id> 5 years in company"),
		"Error in monthly report")
	assert.True(t, partialContains(sc.messages, "1 June, <@anniversary-slack-id> 2 years in company"),
		"Error in monthly report")
	assert.True(t, partialContains(sc.messages, "21 June, <@monthly-report-anniversary-slack-id> 1 year in company"),
		"Error in monthly report")
	assert.True(t, partialContains(sc.messages, "TO CHANNEL 'leaders' USING TOKEN bot-token"),
		"Error in monthly report")
	assert.True(t, partialContains(sc.messages, "1 June, <@birthday-slack-id> 22 years old\n11 June, <@monthly-report-birthday-slack-id> 30 years old"),
		"Error in monthly report birthdays sorting")
	assert.True(t, partialContains(sc.messages, "1 June, <@anniversary-slack-id> 2 years in company\n5 June, <@birthday-slack-id> 5 years in company\n21 June, <@monthly-report-anniversary-slack-id> 1 year in company"),
		"Error in monthly report anniversaries sorting")
	assert.Contains(t, sc.messages,
		"SENDING 'Birthdays:\n1 June, <@birthday-slack-id> 22 years old\n11 June, <@monthly-report-birthday-slack-id> 30 years old\n\nAnniversaries:\n1 June, <@anniversary-slack-id> 2 years in company\n5 June, <@birthday-slack-id> 5 years in company\n21 June, <@monthly-report-anniversary-slack-id> 1 year in company\n' TO CHANNEL 'leaders' USING TOKEN bot-token",
		"Error in monthly report")
}
