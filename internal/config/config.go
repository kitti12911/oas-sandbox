package config

import (
	"time"

	"github.com/kitti12911/lib-monitor/profiling"
	"github.com/kitti12911/lib-monitor/tracing"
	"github.com/kitti12911/lib-util/v3/logger"
)

type Config struct {
	Service     Service          `mapstructure:"service"      validate:"required"`
	Logging     logger.Config    `mapstructure:"logging"`
	Tracing     tracing.Config   `mapstructure:"tracing"`
	Profiling   profiling.Config `mapstructure:"profiling"`
	UserService ServiceAddr      `mapstructure:"user_service" validate:"required"`
}

type Service struct {
	Name            string        `mapstructure:"name"             env:"SERVICE_NAME"      validate:"required"`
	Port            int           `mapstructure:"port"             env:"PORT"              validate:"required,gte=1,lte=65535"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT"`
}

type ServiceAddr struct {
	Host string `mapstructure:"host" env:"USER_SERVICE_HOST" validate:"required"`
	Port int    `mapstructure:"port" env:"USER_SERVICE_PORT" validate:"required,gte=1,lte=65535"`
}
