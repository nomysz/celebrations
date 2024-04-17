package cmd

import (
	"errors"
	"fmt"

	"github.com/slack-go/slack"
)

func SendSlackChannelMsg(channel string, msg string, c *Config) error {
	_, _, err := slack.New(c.Slack.BotToken).PostMessage(
		channel,
		slack.MsgOptionAttachments(
			slack.Attachment{
				Pretext: msg,
			},
		),
	)

	if err != nil {
		return errors.New(
			fmt.Sprintf(
				"Error sending message to Slack channel %s: %s",
				channel,
				err,
			),
		)
	}
	return nil
}

func SendSlackDM(slackId string, msg string, c *Config) error {
	api := slack.New(c.Slack.BotToken)

	slack_ch, _, _, err := api.OpenConversation(
		&slack.OpenConversationParameters{
			Users:    []string{slackId},
			ReturnIM: false,
		},
	)

	if err != nil {
		return errors.New(
			fmt.Sprintf(
				"Error when opening Slack conversation with Slack ID %s: %s",
				slackId,
				err,
			),
		)
	}

	_, _, err = api.PostMessage(slack_ch.ID, slack.MsgOptionText(msg, false))

	if err != nil {
		return errors.New(
			fmt.Sprintf(
				"Error sending DM to person with Slack ID %s: %s",
				slackId,
				err,
			),
		)
	}
	return nil
}

func SetSlackPersonalReminder(slackId string, time string, msg string, c *Config) error {
	_, err := slack.New(c.Slack.UserToken).AddUserReminder(
		slackId,
		msg,
		time,
	)
	if err != nil {
		return errors.New(
			fmt.Sprintf(
				"Error when posting Slack reminder to person with Slack ID %s: %s",
				slackId,
				err,
			),
		)
	}
	return nil
}
