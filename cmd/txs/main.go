package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Commander interface {
	Init(opts interface{}) error
	Execute(args []string) error
	Close()
}

func main() {
	_ = godotenv.Load()

	log.Logger = zerolog.New(os.Stderr).
		With().
		Timestamp().
		Logger().
		Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "2006-01-02 15:04:05",
		})

	var opts Options
	parser := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)

	parser.CommandHandler = func(command flags.Commander, args []string) error {
		c := command.(Commander)

		if err := c.Init(opts.Common); err != nil {
			log.Fatal().Msg(err.Error())
			return nil
		}
		defer c.Close()

		if err := c.Execute(args); err != nil {
			log.Fatal().Msg(err.Error())
			return nil
		}

		return nil
	}

	if _, err := parser.Parse(); err != nil {
		if e, ok := err.(*flags.Error); ok {
			if e.Type == flags.ErrHelp {
				fmt.Print(e.Message)
				os.Exit(0)
			}
		}
		log.Fatal().Err(err).Msg("parse error")
		os.Exit(1)
	}
}
