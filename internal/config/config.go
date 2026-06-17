package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"

	ginadapter "github.com/Medzoner/gomedz/pkg/http/adapter/gin"

	"github.com/Medzoner/gomedz/pkg/auth"
	"github.com/Medzoner/gomedz/pkg/captcha"
	"github.com/Medzoner/gomedz/pkg/config"
	"github.com/Medzoner/gomedz/pkg/connector"
	"github.com/Medzoner/gomedz/pkg/http/server"
	"github.com/Medzoner/gomedz/pkg/logger"
	"github.com/Medzoner/gomedz/pkg/notifier"
	"github.com/Medzoner/gomedz/pkg/observability"
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

	TestConfig struct {
		Obs    observability.Config `envPrefix:"TELEMETRY_"`
		Engine ginadapter.Config    `envPrefix:"ENGINE_"`
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
