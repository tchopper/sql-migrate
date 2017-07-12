package main

import (
	"flag"
	"strings"

	"github.com/rubenv/sql-migrate"
)

type DownCommand struct {
}

func (c *DownCommand) Help() string {
	helpText := `
Usage: sql-migrate down [options] ...

  Undo a database migration.

Options:

	-config=dbconfig.yml           Configuration file to use.
	-env="development"             Environment.

	-config.dialect="mysql"        Config param for dialect.
  -config.datasource="root:..."  Config param for datasource.
  -config.dir="db/migrations"    Config param for directory.
  -config.table="test_table"     Config param for table_name.
  -limit=1                       Limit the number of migrations (0 = unlimited).
  -dryrun                        Don't apply migrations, just print them.

`
	return strings.TrimSpace(helpText)
}

func (c *DownCommand) Synopsis() string {
	return "Undo a database migration"
}

func (c *DownCommand) Run(args []string) int {
	var limit int
	var dryrun bool

	cmdFlags := flag.NewFlagSet("down", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.IntVar(&limit, "limit", 1, "Max number of migrations to apply.")
	cmdFlags.BoolVar(&dryrun, "dryrun", false, "Don't apply migrations, just print them.")
	ConfigFlags(cmdFlags)

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	err := ApplyMigrations(migrate.Down, dryrun, limit)
	if err != nil {
		ui.Error(err.Error())
		return 1
	}

	return 0
}
