package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/rubenv/sql-migrate"
	"gopkg.in/gorp.v1"
	"gopkg.in/yaml.v2"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var dialects = map[string]gorp.Dialect{
	"sqlite3":  gorp.SqliteDialect{},
	"postgres": gorp.PostgresDialect{},
	"mysql":    gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"},
}

var ConfigFile string
var ConfigEnvironment string
var Dialect string
var DataSource string
var Dir string
var Table string

func ConfigFlags(f *flag.FlagSet) {
	f.StringVar(&ConfigFile, "config", "dbconfig.yml", "Configuration file to use.")
	f.StringVar(&ConfigEnvironment, "env", "development", "Environment to use.")
	f.StringVar(&Dialect, "config.dialect", "", "Dialect to use.")
	f.StringVar(&DataSource, "config.data_source", "", "Data source to use.")
	f.StringVar(&Dir, "config.dir", "db/migrations", "Directory to use")
	f.StringVar(&Table, "config.table", "", "Table to use")
}

type Environment struct {
	Dialect    string `yaml:"dialect"`
	DataSource string `yaml:"datasource"`
	Dir        string `yaml:"dir"`
	TableName  string `yaml:"table"`
	SchemaName string `yaml:"schema"`
}

func ReadConfig() (map[string]*Environment, error) {
	file, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}

	config := make(map[string]*Environment)
	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func GetConfig() (string, map[string]*Environment, error) {
	if CheckConfigParams() {
		config := make(map[string]*Environment)
		config[ConfigEnvironment] = &Environment{
			Dialect:    Dialect,
			DataSource: DataSource,
			Dir:        Dir,
			TableName:  Table,
		}

		return "config.X flags", config, nil
	} else {
		config, err := ReadConfig()
		if err != nil {
			return "yaml file", nil, err
		}
		return "yaml file", config, nil
	}
}

func CheckConfigParams() bool {
	if Dialect != "" || DataSource != "" || Table != "" {
		return true
	}
	return false
}

func GetEnvironment() (*Environment, error) {
	origin, config, err := GetConfig()
	if err != nil {
		return nil, err
	}

	env := config[ConfigEnvironment]
	if env == nil {
		return nil, fmt.Errorf("%s: No Environment: %s", origin, ConfigEnvironment)
	}

	if env.Dialect == "" {
		return nil, fmt.Errorf("%s: No dialect specified", origin)
	}

	if env.DataSource == "" {
		return nil, fmt.Errorf("%s: No data source specified", origin)
	}
	env.DataSource = os.ExpandEnv(env.DataSource)

	if env.Dir == "" {
		env.Dir = "migrations"
	}

	if env.TableName != "" {
		migrate.SetTable(env.TableName)
	}

	if env.SchemaName != "" {
		migrate.SetSchema(env.SchemaName)
	}

	return env, nil
}

func GetConnection(env *Environment) (*sql.DB, string, error) {
	db, err := sql.Open(env.Dialect, env.DataSource)
	if err != nil {
		return nil, "", fmt.Errorf("Cannot connect to database: %s", err)
	}

	// Make sure we only accept dialects that were compiled in.
	_, exists := dialects[env.Dialect]
	if !exists {
		return nil, "", fmt.Errorf("Unsupported dialect: %s", env.Dialect)
	}

	return db, env.Dialect, nil
}
