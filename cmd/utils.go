package cmd

import (
	"log"
	"os"

	"github.com/quasoft/pgmig/db"

	"github.com/spf13/cobra"
)

func getFlagOrEnv(cmd *cobra.Command, flagName string, envName string) string {
	// If flag has been set in command line arguments, use that
	if cmd.Flags().Changed(flagName) {
		value, err := cmd.Flags().GetString(flagName)
		if err == nil {
			return value
		}
		log.Printf("could not get value for changed flag %s", flagName)
	}

	// If corresponding environment variable has been set, use that
	if envName != "" {
		value := os.Getenv(envName)
		if value != "" {
			return value
		}
	}

	// Else, use the default value
	value, err := cmd.Flags().GetString(flagName)
	if err == nil {
		return value
	}
	log.Printf("could not get default value for flag %s", flagName)

	return ""
}

func ParseFlagsOrEnv(s *db.Session, cmd *cobra.Command) {
	s.Host = getFlagOrEnv(cmd, "host", "PGHOST")
	s.Port = getFlagOrEnv(cmd, "port", "PGPORT")
	s.Database = getFlagOrEnv(cmd, "database", "PGDATABASE")
	s.Username = getFlagOrEnv(cmd, "username", "PGUSER")
	s.SslMode = getFlagOrEnv(cmd, "ssl-mode", "PGSSLMODE")
	interactive, err := cmd.Flags().GetBool("interactive")
	if err != nil {
		s.Interactive = true
	} else {
		s.Interactive = interactive
	}
}
