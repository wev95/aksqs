package aksqs

import (
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	resolveLog()
}

type SeverityHook struct{}

func (h SeverityHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Str("logName", "aksqs")
}

var logger zerolog.Logger = log.Hook(SeverityHook{})

func resolveLog() {

	level, _ := zerolog.ParseLevel(zerolog.LevelErrorValue)

	zerolog.TimestampFieldName = "@timestamp"
	zerolog.CallerFieldName = "logger_name"

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}

	logger = log.With().Caller().Logger()

	zerolog.SetGlobalLevel(level)
}
