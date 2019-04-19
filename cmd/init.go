package cmd

import (
	"fmt"
	"os"

	"github.com/quasoft/pgmig/db"

	"github.com/spf13/cobra"
)

var initSession *db.Session

func init() {
	initSession := db.NewSession()
	initCmd.Flags().SortFlags = false
	initCmd.Flags().StringP("host", "", "localhost", "Hostname or IP address of PostgreSQL server")
	initCmd.Flags().StringP("port", "p", "5432", "The port of the DB instance")
	initCmd.Flags().StringP("database", "d", "localhost", "Hostname or IP address of PostgreSQL server")
	initCmd.Flags().StringP("username", "U", "", "The username of a superuser")
	initCmd.Flags().StringP("ssl-mode", "s", "disable", "SSL mode (disable | allow | prefer | require | verify-ca | validate-full)")
	initCmd.Flags().StringVarP(&initSession.ChangelogName, "changelog-name", "n", "changelog", "Name of table to write change logs to")
	initCmd.Flags().BoolP("interactive", "i", true, "Ask for password if not provided in PG_PASSWORD environment variable or the PGPASSFILE")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "pgmig init [--host <string>] [--port <int>] [--database <string>] [--username <string>] [--ssl-mode <string>] [--changelog-name <string>] [--interactive]",
	Short: "Create changelog table in specified PostgreSQL database",
	Example: `  Specify database with PG environment variables:
  pgmig init

  Specify database and changelog name with command arguments:
  pgmig init --host 10.0.0.1 -d testdb -U postgres -n myproj_changelog

  Provide password interactively:
  pgmig init --host 10.0.0.1 -d testdb -U postgres -i
`,
	Run: func(cmd *cobra.Command, args []string) {
		ParseFlagsOrEnv(initSession, cmd)

		// Connect to DB
		fmt.Printf("Connecting to %s:%s\n", initSession.Host, initSession.Port)
		err := initSession.Connect()
		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}
		defer initSession.Disconnect()

		// Create changelog table if it does not exist
		exists, err := initSession.EnsureChangelogExists()
		if err != nil || !exists {
			fmt.Println("Error: changelog table does not exists and could not be created!")
			initSession.Disconnect()
			os.Exit(1)
		}

		fmt.Println("Changelog table created.")
	},
}
