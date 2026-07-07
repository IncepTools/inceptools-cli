package main

import (
	"flag"
	"fmt"
	"github.com/IncepTools/inceptools-cli/src/cmd"
	"github.com/IncepTools/inceptools-cli/src/ui"
	"os"
)

var Version = "dev"

func main() {
	// Global version flags
	versionFlagShort := flag.Bool("v", false, "Print version and exit")
	versionFlagLong := flag.Bool("version", false, "Print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: inceptools <command> [arguments]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  init      Initialize a new inceptools project\n")
		fmt.Fprintf(os.Stderr, "  create    Create a new migration file\n")
		fmt.Fprintf(os.Stderr, "  migrate   Run pending migrations\n")
		fmt.Fprintf(os.Stderr, "  down      Roll back migrations\n")
		fmt.Fprintf(os.Stderr, "  update    Update inceptools to a specific or latest version\n")
		fmt.Fprintf(os.Stderr, "  version   Print version information\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *versionFlagShort || *versionFlagLong {
		fmt.Printf("inceptools version %s\n", Version)
		return
	}

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	command := flag.Arg(0)
	switch command {
	case "init":
		cmd.HandleInit()
	case "update":
		cmd.HandleUpdate(Version)
	case "version":
		fmt.Printf("inceptools version %s\n", Version)
	case "create":
		createCmd := flag.NewFlagSet("create", flag.ExitOnError)
		dbName := createCmd.String("db", "", "Named database from icpt.json (when multiple are configured)")
		createCmd.Parse(flag.Args()[1:])
		cmd.HandleCreate(*dbName)
	case "migrate":
		migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
		dbName := migrateCmd.String("db", "", "Named database from icpt.json (default: all configured databases)")
		migrateCmd.Parse(flag.Args()[1:])
		dbURL := ""
		if migrateCmd.NArg() >= 1 {
			dbURL = migrateCmd.Arg(0)
		}
		cmd.HandleMigrate(dbURL, *dbName)
	case "down":
		downCmd := flag.NewFlagSet("down", flag.ExitOnError)
		steps := downCmd.Int("steps", 0, "Number of migrations to roll back (0 = all)")
		dbName := downCmd.String("db", "", "Named database from icpt.json (required when multiple are configured)")
		downCmd.Parse(flag.Args()[1:])
		dbURL := ""
		if downCmd.NArg() >= 1 {
			dbURL = downCmd.Arg(0)
		}
		cmd.HandleDown(dbURL, *dbName, *steps)
	default:
		ui.Error("Unknown command: %s", command)
		flag.Usage()
		os.Exit(1)
	}
}
