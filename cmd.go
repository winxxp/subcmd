package subcmd

import (
	"bytes"
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"os/exec"
	"syscall"
)

type Options struct {
	args       []string
	hideWidows bool

	quit   func()
	logfun func(string)
}

type Option func(o *Options)

func HideWindows() Option {
	return func(o *Options) {
		o.hideWidows = true
	}
}

func Args(args ...string) Option {
	return func(o *Options) {
		o.args = make([]string, len(args))
		copy(o.args, args)
	}
}

func QuitHandle(h func()) Option {
	return func(o *Options) {
		o.quit = h
	}
}

func LogHandle(h func(string)) Option {
	return func(o *Options) {
		o.logfun = h
	}
}

type Cmd struct {
	// 应用程序名称
	binName string

	opts Options
}

func New(binName string, opts ...Option) *Cmd {
	c := &Cmd{
		binName: binName,
		opts:    Options{},
	}

	for _, opt := range opts {
		opt(&c.opts)
	}

	return c
}

func (c *Cmd) Run(ctx context.Context) error {
	var (
		g, ctx2 = errgroup.WithContext(ctx)
	)

	defer func() {
		if c.opts.quit != nil {
			c.opts.quit()
		}
	}()

	r, w, err := os.Pipe()
	if err != nil {
		return err
	}

	g.Go(func() error {
		cmd := exec.CommandContext(ctx2, c.binName, c.opts.args...)

		if c.opts.hideWidows {
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		}

		cmd.Stderr = w
		cmd.Stdout = w

		defer w.Close()

		if err := cmd.Run(); err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		buf := make([]byte, 1024)

		for {
			n, err := r.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			} else {
				lines := bytes.Split(buf[:n], []byte("\n"))
				for _, line := range lines {
					if c.opts.logfun != nil {
						c.opts.logfun(string(line) + "\n")
					}
				}
			}
		}
		return nil
	})

	return g.Wait()
}
