package slack

import (
	"errors"
	"fmt"

	"github.com/slack-go/slack"
)

type SlackClient struct {
	botToken  string
	userToken string
}

func NewSlackClient(botToken string, userToken string) *SlackClient {
	return &SlackClient{
		botToken:  botToken,
		userToken: userToken,
	}
}

type ChannelMessenger interface {
	SendSlackChannelMsg(channel string, msg string) error
}

type DirectMessenger interface {
	SendSlackDM(slackId string, msg string) error
}

type PersonalReminderSetter interface {
	SetSlackPersonalReminder(slackId string, time string, msg string) error
}

type SlackCommunicator interface {
	ChannelMessenger
	DirectMessenger
	PersonalReminderSetter
}

func (sc *SlackClient) SendSlackChannelMsg(channel string, msg string) error {
	_, _, err := slack.New(sc.botToken).PostMessage(
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

func (sc *SlackClient) SendSlackDM(slackId string, msg string) error {
	api := slack.New(sc.botToken)

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

func (sc *SlackClient) SetSlackPersonalReminder(slackId string, time string, msg string) error {
	_, err := slack.New(sc.userToken).AddUserReminder(
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
