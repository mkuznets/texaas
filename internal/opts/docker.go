package opts

type Docker struct {
	UID int `long:"uid" env:"UID" required:"true" description:"docker container UID"`
	GID int `long:"gid" env:"GID" required:"true" description:"docker container GID"`
}
