package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/jessevdk/go-flags"
	"qubert/internal/config"
	"qubert/internal/installer"
	"qubert/internal/logger"

	"qubert/application"
)

var (
	Version string
	Commit  string
)

type options struct {
	ConfigFile  string `short:"c" long:"config" description:"Config file"`
	Daemon      bool   `short:"d" long:"daemon" description:"Run as daemon"`
	ShowVersion bool   `short:"v" long:"version" description:"Show version and exit"`
	Install     bool   `short:"i" long:"install" description:"Install and run"`
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

	p, err := os.StartProcess(args[0], append([]string{"qubert"}, args[1:]...), attr)
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

	flagParser := flags.NewParser(opt, flags.HelpFlag|flags.PassDoubleDash)

	_, err := flagParser.ParseArgs(args)
	if err != nil {
		return err
	}

	if opt.Install && opt.Daemon {
		return fmt.Errorf("using 'install' and 'daemon' is not supported at the same time")
	}

	if opt.Daemon {
		err = RunInBackground(args)
		if err != nil {
			return err
		}
	}

	cfg, err := config.Load(opt.ConfigFile)
	if err != nil {
		return err
	}

	if opt.Install {
		log := logger.CreateLogger(false)

		err = installer.InstallCurrentVersion(cfg, log)
		if err != nil {
			return err
		}

		log.Info("successfully installed")

		os.Exit(0)
	}

	if opt.ShowVersion {
		message := `Qubert

Version: %s
Git commit: %s
`
		_, _ = fmt.Fprintf(os.Stderr, message, Version, Commit)
		return nil
	}

	ctx := context.Background()

	app := application.NewApplication(cfg, Version, Commit)

	err = app.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
