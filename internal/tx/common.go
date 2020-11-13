package tx

// Options is a group of common options for all subcommands.
type Options struct {
	Debug bool `long:"debug" description:"Enable debug logging" env:"TX_DEBUG"`
}

// Command is a common part of all subcommands.
type Command struct {
}

func (cmd *Command) Init(opts interface{}) error {
	return nil
}

func (cmd *Command) Close() {

}
