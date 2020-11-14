package run

import (
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Options struct {
	Autoremove    bool
	Image         string
	Command       []string
	Containter    string
	Binds         []string
	EnableNetwork bool
	Stdout        io.Writer
	Stderr        io.Writer
	WorkingDir    string
	Env           []string
	User          string
	Memory        int64
}

func New(options []Option) (*Options, error) {
	opts := &Options{
		EnableNetwork: true,
		WorkingDir:    "/",
	}
	for _, opt := range options {
		opt(opts)
	}

	if opts.Image == "" || len(opts.Command) == 0 {
		return nil, errors.Errorf("image and command required")
	}

	if opts.Containter == "" {
		u, err := uuid.NewUUID()
		if err != nil {
			return nil, err
		}
		opts.Containter = u.String()
	}

	return opts, nil
}

type Option = func(*Options)

func Autoremove(v bool) Option {
	return func(r *Options) {
		r.Autoremove = v
	}
}

func Image(repo, version string) Option {
	return func(r *Options) {
		r.Image = fmt.Sprintf("%s:%s", repo, version)
	}
}

func Cmd(args ...string) Option {
	return func(r *Options) {
		r.Command = args
	}
}

func EnableNetwork(v bool) Option {
	return func(r *Options) {
		r.EnableNetwork = v
	}
}

func WorkingDir(v string) Option {
	return func(r *Options) {
		r.WorkingDir = v
	}
}

func Mount(src, dst, opts string) Option {
	return func(r *Options) {
		arg := fmt.Sprintf("%s:%s", src, dst)
		if opts != "" {
			arg = fmt.Sprintf("%s:%s", arg, opts)
		}
		r.Binds = append(r.Binds, arg)
	}
}

func Env(name, value string) Option {
	return func(r *Options) {
		r.Env = append(r.Env, fmt.Sprintf("%s=%s", name, value))
	}
}

func CombinedOutput(w io.Writer) Option {
	return func(r *Options) {
		r.Stdout = w
		r.Stderr = w
	}
}

func Stderr(w io.Writer) Option {
	return func(r *Options) {
		r.Stderr = w
	}
}

func Stdout(w io.Writer) Option {
	return func(r *Options) {
		r.Stdout = w
	}
}

func UID(uid, gid int) Option {
	return func(r *Options) {
		user := fmt.Sprintf("%d", uid)
		if gid != -1 {
			user = fmt.Sprintf("%s:%d", user, gid)
		}
		r.User = user
	}
}

func Memory(v int64) Option {
	return func(r *Options) {
		r.Memory = v
	}
}
