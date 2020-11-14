package docker

import (
	"context"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pkg/errors"
	"mkuznets.com/go/texaas/internal/docker/run"
)

var (
	ExitError = errors.New("non-zero exit code")
)

func (d *Docker) Run(ctx context.Context, opts ...run.Option) error {

	r, err := run.New(opts)
	if err != nil {
		return err
	}

	contConf := &container.Config{
		Image:           r.Image,
		Cmd:             r.Command,
		Tty:             false,
		NetworkDisabled: !r.EnableNetwork,
		AttachStdout:    r.Stdout != nil,
		AttachStderr:    r.Stderr != nil,
		WorkingDir:      r.WorkingDir,
		Env:             r.Env,
		User:            r.User,
	}

	hostConf := &container.HostConfig{
		AutoRemove: r.Autoremove,
		Binds:      r.Binds,
		Resources: container.Resources{
			Memory: r.Memory,
		},
	}

	createResp, err := d.client.ContainerCreate(ctx, contConf, hostConf, nil, r.Containter)
	if err != nil {
		return err
	}

	var errCh chan error

	if r.Stderr != nil || r.Stdout != nil {
		options := types.ContainerAttachOptions{
			Stream: true,
			Stdin:  false,
			Stdout: r.Stdout != nil,
			Stderr: r.Stderr != nil,
		}

		attachResp, attachErr := d.client.ContainerAttach(ctx, createResp.ID, options)
		if attachErr != nil {
			return attachErr
		}
		defer attachResp.Close()

		errCh = promise.Go(func() error {
			if r.Stderr == nil {
				r.Stderr = ioutil.Discard
			}
			if r.Stdout == nil {
				r.Stdout = ioutil.Discard
			}
			if _, err := stdcopy.StdCopy(r.Stdout, r.Stderr, attachResp.Reader); err != nil {
				return errors.WithMessage(err, "output error")
			}
			return nil
		})
	}

	if err := d.client.ContainerStart(ctx, createResp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	exitCode, err := d.client.ContainerWait(ctx, createResp.ID)
	if err != nil {
		return err
	}

	if errCh != nil {
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
			break
		case <-ctx.Done():
			break
		}
	}

	if exitCode != 0 {
		return errors.Wrapf(ExitError, "exit code: %d", exitCode)
	}

	return nil
}
