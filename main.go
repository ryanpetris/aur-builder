package main

import (
	"flag"
	"fmt"
	"github.com/ryanpetris/aur-builder/cli"
	"github.com/ryanpetris/aur-builder/config"
	"log/slog"
	"os"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cmdConfig := flag.String("config", "", "path to configuration file")
	flag.Parse()

	if *cmdConfig != "" {
		cfg := config.GetGlobalConfig()

		if err := cfg.Load(*cmdConfig); err != nil {
			panic(err)
		}
	}

	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("invalid command")
		os.Exit(1)
	}

	switch args[0] {
	case "import":
		cli.ImportMain(args)

	case "needs-build":
		cli.NeedsBuildMain(args)

	case "prepare":
		cli.PrepareMain(args)

	case "update":
		cli.UpdateMain(args)

	case "update-vcs":
		cli.UpdateVcsMain(args)

	case "formatconfig":
		cli.FormatConfigMain(args)

	case "bump-pkgrel":
		cli.BumpPkgrel(args)

	default:
		fmt.Println("invalid command")
		os.Exit(1)
	}
}
