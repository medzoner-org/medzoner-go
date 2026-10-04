package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"

	ginadapter "github.com/medzoner-org/gomedz/pkg/http/adapter/gin"

	"github.com/medzoner-org/gomedz/pkg/auth"
	"github.com/medzoner-org/gomedz/pkg/captcha"
	"github.com/medzoner-org/gomedz/pkg/config"
	"github.com/medzoner-org/gomedz/pkg/connector"
	"github.com/medzoner-org/gomedz/pkg/http/server"
	"github.com/medzoner-org/gomedz/pkg/logger"
	"github.com/medzoner-org/gomedz/pkg/notifier"
	"github.com/medzoner-org/gomedz/pkg/observability"
)

type (
	RootPath string

	Config struct {
		Obs       observability.Config `envPrefix:"TELEMETRY_"`
		Engine    ginadapter.Config    `envPrefix:"ENGINE_"`
		Logger    logger.Config        `envPrefix:"LOGGER_"`
		Auth      auth.Config          `envPrefix:"AUTH_"`
		Server    server.Config        `envPrefix:"SERVER_"`
		Mailer    notifier.Config      `envPrefix:"MAILER_"`
		Database  connector.Config     `envPrefix:"DATABASE_"`
		RootPath  RootPath             `env:"ROOT_PATH"`
		Recaptcha captcha.Config       `envPrefix:"RECAPTCHA_"`
	}
)

func NewConfig() (Config, error) {
	cfg := Config{}

	if err := env.ParseWithOptions(&cfg, env.Options{}); err != nil {
		return cfg, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	config.PrintConfigEnv(cfg, "")

	return cfg, nil
}
