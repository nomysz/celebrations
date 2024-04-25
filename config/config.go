package config

import (
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Person struct {
	SlackMemberID     string    `mapstructure:"slack_member_id" validate:"required"`
	BirthDate         time.Time `mapstructure:"birth_date" validate:"required"`
	JoinDate          time.Time `mapstructure:"join_date" validate:"required"`
	LeadSlackMemberID *string   `mapstructure:"lead_slack_member_id" validate:"required"`
}

type DownloadingUsers struct {
	BirthdayCustomFieldName string `mapstructure:"birthday_custom_field_name" validate:"required"`
	JoinDateCustomFieldName string `mapstructure:"join_date_custom_field_name" validate:"required"`
}

type AnniversaryChannelReminder struct {
	Enabled         bool   `mapstructure:"enabled"`
	ChannelName     string `mapstructure:"channel_name" validate:"required"`
	MessageTemplate string `mapstructure:"message_template" validate:"required"`
}

type BirthdaysChannelReminder struct {
	Enabled         bool   `mapstructure:"enabled"`
	ChannelName     string `mapstructure:"channel_name" validate:"required"`
	MessageTemplate string `mapstructure:"message_template" validate:"required"`
}

type BirthdaysPersonalReminder struct {
	Enabled         bool   `mapstructure:"enabled"`
	Time            string `mapstructure:"time" validate:"required"`
	MessageTemplate string `mapstructure:"message_template" validate:"required"`
}

type BirthdaysDirectMessageReminder struct {
	Enabled                    bool     `mapstructure:"enabled"`
	MessageTemplate            string   `mapstructure:"message_template" validate:"required"`
	PreReminderDaysBefore      int64    `mapstructure:"pre_reminder_days_before" validate:"required"`
	PreRemidnerMessageTemplate string   `mapstructure:"pre_remidner_message_template" validate:"required"`
	AlwaysNotifySlackIds       []string `mapstructure:"always_notify_slack_ids" validate:"required"`
}

type Slack struct {
	BotToken                       string                         `mapstructure:"bot_token" validate:"required"`
	UserToken                      string                         `mapstructure:"user_token" validate:"required"`
	AnniversaryChannelReminder     AnniversaryChannelReminder     `mapstructure:"anniversary_channel_reminder" validate:"required"`
	BirthdaysChannelReminder       BirthdaysChannelReminder       `mapstructure:"birthdays_channel_reminder" validate:"required"`
	BirthdaysPersonalReminder      BirthdaysPersonalReminder      `mapstructure:"birthdays_personal_reminder" validate:"required"`
	BirthdaysDirectMessageReminder BirthdaysDirectMessageReminder `mapstructure:"birthdays_direct_message_reminder" validate:"required"`
	DownloadingUsers               DownloadingUsers               `mapstructure:"downloading_users" validate:"required"`
}

type Config struct {
	Slack  Slack    `mapstructure:"slack" validate:"required"`
	People []Person `mapstructure:"people" validate:"required"`
}

func GetConfig() *Config {
	var c Config
	if err := viper.Unmarshal(&c, viper.DecodeHook(
		mapstructure.StringToTimeHookFunc(time.DateOnly),
	)); err != nil {
		log.Fatalln("Error marshalling file:" + err.Error())
	}
	return &c
}

func InitConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("Error reading config file:" + err.Error())
	}

	c := GetConfig()

	if err := validator.New(
		validator.WithRequiredStructEnabled(),
	).Struct(c); err != nil {
		log.Fatalln("Missing required config attributes:" + err.Error())
	}

	// For some reason it's not done properly by validator
	for _, p := range c.People {
		if p.SlackMemberID == "" {
			log.Println(p.SlackMemberID)
			log.Fatalln("Missing slack_member_id")
		}
		if p.BirthDate.IsZero() {
			log.Fatalln("Missing birth date for slack_member_id: " + p.SlackMemberID)
		}
		if p.JoinDate.IsZero() {
			log.Fatalln("Missing join date for slack_member_id: " + p.SlackMemberID)
		}
	}
}
