package tx

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Options is a group of common options for all subcommands.
type Options struct {
	ConfigPath string `short:"c" long:"config" description:"Config path" env:"TX_CONFIG"`
	Debug      bool   `long:"debug" description:"Enable debug logging" env:"TX_DEBUG"`
}

// Command is a common part of all subcommands.
type Command struct {
}

func (cmd *Command) Init(opts interface{}) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
	})

	return nil
}

func (cmd *Command) Close() {

}
