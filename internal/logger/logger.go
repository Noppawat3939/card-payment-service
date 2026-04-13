package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Level  string
	Pretty bool
}

func Init(cfg Config) {
	l, e := zerolog.ParseLevel(cfg.Level)
	if e != nil {
		l = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(l)
	zerolog.TimeFieldFormat = time.RFC3339

	var writer io.Writer = os.Stdout
	if cfg.Pretty {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}
	}

	log.Logger = zerolog.New(writer).With().Timestamp().Caller().Logger()
}

func With(service string) zerolog.Logger {
	return log.Logger.With().Str("service", service).Logger()
}
