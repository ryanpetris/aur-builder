package main

import (
	"flag"
	"fmt"
	"github.com/ryanpetris/aur-builder/cli"
	"github.com/ryanpetris/aur-builder/config"
	"os"
)

func main() {
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

	case "prepare":
		cli.PrepareMain(args)

	case "update":
		cli.UpdateMain(args)

	default:
		fmt.Println("invalid command")
		os.Exit(1)
	}
}
