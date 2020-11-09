package docker

import (
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

type Docker struct {
	client *client.Client
}

func New() (*Docker, error) {
	c, err := client.NewEnvClient()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	d := &Docker{
		client: c,
	}
	return d, nil
}
