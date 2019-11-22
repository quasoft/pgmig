package cmd

import (
	"fmt"
	"os"

	"github.com/quasoft/pgmig/db"
	"github.com/quasoft/pgmig/mig"

	"github.com/spf13/cobra"
)

var applySession = db.NewSession()
var applyDir = mig.NewDir()
var createChangelog bool

func init() {
	applyCmd.Flags().SortFlags = false
	applyCmd.Flags().StringVarP(&applyDir.Path, "dir", "D", "", "Local directory with migration scripts (default: current dir)")
	applyCmd.Flags().StringP("host", "", "localhost", "Hostname or IP address of PostgreSQL server")
	applyCmd.Flags().StringP("port", "p", "5432", "The port of the DB instance")
	applyCmd.Flags().StringP("database", "d", "localhost", "Hostname or IP address of PostgreSQL server")
	applyCmd.Flags().StringP("username", "U", "", "The username of a superuser")
	applyCmd.Flags().StringP("ssl-mode", "s", "disable", "SSL mode (disable | allow | prefer | require | verify-ca | validate-full)")
	applyCmd.Flags().BoolVarP(&createChangelog, "create-changelog", "c", false, "Automatically create changelog table if it does not exist")
	applyCmd.Flags().StringVarP(&applySession.ChangelogName, "changelog-name", "n", "changelog", "Name of table to write change logs to")
	applyCmd.Flags().BoolP("interactive", "i", true, "Ask for password if not provided in PG_PASSWORD environment variable or the PGPASSFILE")
	rootCmd.AddCommand(applyCmd)
}

var applyCmd = &cobra.Command{
	Use:   "pgmig apply [--dir <path>] [--host <string>] [--port <int>] [--database <string>] [--username <string>] [--ssl-mode <string>] [--create-changelog <bool>] [--changelog-name <string>] [--interactive]",
	Short: "Applies migration SQL files from a directory to a specified PostgreSQL database",
	Example: `  Apply pending migrations:
  pgmig apply

  Automatically create a custom changelog table and immediately apply pending migrations:
  pgmig apply -c -n myproj_changelog

  Set migrations directory and database
  pgmig apply -D ~/proj/db/migrations --host 10.0.0.1 -d testdb -U postgres

  Apply pending migrations and log to an existing changelog table:
  pgmig apply -n myproj_changelog
`,
	Run: func(cmd *cobra.Command, args []string) {
		ParseFlagsOrEnv(applySession, cmd)

		// Connect to DB
		fmt.Printf("Connecting to %s:%s\n", applySession.Host, applySession.Port)
		err := applySession.Connect()
		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}
		defer applySession.Disconnect()

		// Create changelog table if it does not exist
		if createChangelog {
			exists, err := applySession.EnsureChangelogExists()
			if err != nil || !exists {
				fmt.Println("Error: changelog table does not exists and could not be created!")
				applySession.Disconnect()
				os.Exit(1)
			}
		}

		// Scan specified directory for migration files that have not been applied (with ID > lastID)
		migrations, err := applySession.PendingMigrations(applyDir)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			applySession.Disconnect()
			os.Exit(1)
		}

		if len(migrations) == 0 {
			fmt.Println("There are no pending migrations to apply.")
			applySession.Disconnect()
			os.Exit(0)
		}

		// Apply each file sequentially
		for _, m := range migrations {
			fmt.Printf("Applying migration #%d from file %s.\r\n", m.Ver, m.FileName)
			err := applySession.Apply(m)
			if err != nil {
				fmt.Println("Error: " + err.Error())
				applySession.Disconnect()
				os.Exit(1)
			}
			fmt.Printf("Migration #%d applied successfully.\r\n", m.Ver)
		}
		fmt.Printf("Successfully applied %d migrations.\r\n", len(migrations))
	},
}
