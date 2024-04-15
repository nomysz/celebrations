package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	filename       = "people.yml"
	limit          int
	download_users = &cobra.Command{
		Use:   "download-users",
		Short: fmt.Sprintf("Get users from Slack and save as %s", filename),
		Long:  fmt.Sprintf("Get users from Slack and save as %s. Filters out users marked as bots and deleted.", filename),
		Run:   func(cmd *cobra.Command, args []string) { downloadUserFromSlack() },
	}
)

func init() {
	download_users.Flags().IntVarP(&limit, "limit", "l", 10, "Limit the number of users being downloaded")
}

type SlackUser struct {
	Name              string `yaml:"name"`
	SlackMemberID     string `yaml:"slack_member_id"`
	BirthDate         string `yaml:"birth_date"`
	JoinDate          string `yaml:"join_date"`
	LeadSlackMemberID string `yaml:"lead_slack_member_id"`
}

func downloadUserFromSlack() {
	c := getConfig()

	api := slack.New(c.Slack.BotToken)

	users, err := api.GetUsers()

	if err != nil {
		log.Fatal("Error downloading users from Slack", err)
	}

	var SlackUsers []SlackUser

	var local_limit int = 0

	for _, u := range users {
		if u.IsBot || u.Deleted {
			continue
		} else {
			local_limit++
		}

		if local_limit >= limit {
			break
		}

		userProfile, err := api.GetUserProfile(
			&slack.GetUserProfileParameters{UserID: u.ID, IncludeLabels: false},
		)

		if err != nil {
			log.Fatal("Error downloading user profile from Slack", err)
		}

		userProfileMap := userProfile.Fields.ToMap()

		p := SlackUser{
			Name:          u.Profile.DisplayName,
			SlackMemberID: u.ID,
			BirthDate:     userProfileMap[c.Slack.DownloadingUsers.BirthdayCustomFieldName].Value,
			JoinDate:      userProfileMap[c.Slack.DownloadingUsers.JoinDateCustomFieldName].Value,
		}

		SlackUsers = append(SlackUsers, p)
	}

	bytes, err := yaml.Marshal(SlackUsers)
	if err != nil {
		log.Fatal("Error marshalling results into yaml", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Fatal("Error creating file", err)
	}
	defer file.Close()

	_, err = file.Write(bytes)
	if err != nil {
		log.Fatal("Error writing to file", err)
	}

	log.Println(
		fmt.Sprintf(
			"%d users downloaded and persisted to file %s",
			len(SlackUsers),
			filename,
		),
	)
}
