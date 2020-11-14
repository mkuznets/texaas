package opts

type outputDir struct {
	Dir string `long:"dir" env:"DIR" required:"true" description:"fileserver directory"`
}

type OutputFileServer struct {
	outputDir
	Addr string `long:"addr" env:"ADDR" required:"true" description:"fileserver addr"`
}

type Output struct {
	outputDir
	URL string `long:"url" env:"URL" required:"true" description:"external URL of fileserver"`
	UID int    `long:"uid" env:"UID" required:"true" description:"fileserver UID"`
	GID int    `long:"gid" env:"GID" required:"true" description:"fileserver GID"`
}
