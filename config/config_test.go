package config

import (
	"io"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadingEnvVars(t *testing.T) {
	log.SetOutput(io.Discard)

	InitConfig("test_config")
	c := GetConfig()

	assert.True(t, c.Slack.BotToken == "test-bot-token")
	assert.True(t, c.Slack.UserToken == "test-user-token")
}
