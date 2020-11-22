package config

import (
	"github.com/oars-sigs/oars-cloud/core"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

func Load(path string) (*core.Config, error) {
	if path == "" {
		path = ".env"
	}
	logrus.SetReportCaller(true)
	godotenv.Load(path)
	cfg := core.Config{}
	err := envconfig.Process("", &cfg)
	return &cfg, err
}
