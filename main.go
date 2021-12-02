package main

import (
	"context"
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
	"syscall"

	"qubert/application"
	"qubert/logger"
)

type options struct {
	ConfigFile string `short:"c" long:"config" description:"Config file"`
	Background bool   `short:"b" description:"Run in background"`
}

func main() {
	err := mainFunc(os.Args)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		os.Exit(1)
	}
}

func RunInBackground(args []string) error {
	for i, a := range args {
		if a == "-b" {
			args = append(args[:i], args[i+1:]...)
		}
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}

	attr := &os.ProcAttr{
		Dir: path,
		Env: os.Environ(),
		Files: []*os.File{
			os.Stdin,
			nil,
			nil,
		},
		Sys: &syscall.SysProcAttr{
			Noctty: true,
			//Credential: &syscall.Credential{
			//	Uid: uint32(os.Getuid()),
			//	Gid: uint32(os.Getuid()),
			//},
		},
	}

	p, err := os.StartProcess(args[0], append([]string{"qbert"}, args[1:]...), attr)
	if err != nil {
		return err
	}

	err = p.Release()
	if err != nil {
		return err
	}

	logger.CreateLogger(false).Info("Background process successfully started.")

	os.Exit(0)

	return nil
}

func mainFunc(args []string) error {
	opt := &options{
		ConfigFile: "config.json",
	}

	_, err := flags.ParseArgs(opt, args)
	if err != nil {
		return err
	}

	if opt.Background {
		err = RunInBackground(args)
		if err != nil {
			return err
		}
	}

	ctx := context.Background()

	cfg, err := getConfig(opt.ConfigFile)
	if err != nil {
		return err
	}

	app := application.NewApplication(cfg)

	err = app.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
