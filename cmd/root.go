package cmd

import (
	"fmt"
	"os"

	"github.com/quasoft/pgmig/db"
	"github.com/quasoft/pgmig/mig"

	"github.com/spf13/cobra"
)

var rootSession = db.NewSession()
var rootDir = mig.NewDir()

func init() {
	rootCmd.Flags().SortFlags = false
	rootCmd.Flags().StringVarP(&rootDir.Path, "dir", "D", "", "Local directory with migration scripts (default: current dir)")
	rootCmd.Flags().StringP("host", "", "localhost", "Hostname or IP address of PostgreSQL server")
	rootCmd.Flags().StringP("port", "p", "5432", "The port of the DB instance")
	rootCmd.Flags().StringP("database", "d", "localhost", "Hostname or IP address of PostgreSQL server")
	rootCmd.Flags().StringP("username", "U", "", "The username of a superuser")
	rootCmd.Flags().StringP("ssl-mode", "s", "disable", "SSL mode (disable | allow | prefer | require | verify-ca | validate-full)")
	rootCmd.Flags().StringVarP(&rootSession.ChangelogName, "changelog-name", "n", "changelog", "Name of table to write change logs to")
	rootCmd.Flags().BoolP("interactive", "i", true, "Ask for password if not provided in PG_PASSWORD environment variable or the PGPASSFILE")
}

var rootCmd = &cobra.Command{
	Use:   "pgmig [--dir <path>] [--host <string>] [--port <int>] [--database <string>] [--username <string>] [--ssl-mode <string>] [--changelog-name <string>] [--interactive]",
	Short: "Check if directory contains migration files, which have not been applied yet",
	Example: `  Checks current directory for migration files that have not been applied to the database specified by PG environment variables:
  pgmig

  Checks the directory and database specified with command arguments:
  pgmig -D ~/proj/db/migrations --host 10.0.0.1 -d testdb -U postgres
`,
	Run: func(cmd *cobra.Command, args []string) {
		ParseFlagsOrEnv(rootSession, cmd)

		// Connect to DB
		fmt.Printf("Connecting to %s:%s\n", rootSession.Host, rootSession.Port)
		err := rootSession.Connect()
		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}
		defer rootSession.Disconnect()

		// Scan specified directory for migration files that have not been applied (with ID > lastID)
		migrations, err := rootSession.PendingMigrations(rootDir)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			rootSession.Disconnect()
			os.Exit(1)
		}

		if len(migrations) == 0 {
			fmt.Println("There are no pending migrations to apply.")
			rootSession.Disconnect()
			os.Exit(0)
		}

		// Print information about each pending migration
		fmt.Println("List of pending migrations:")
		fmt.Println("---------------------------")
		for _, m := range migrations {
			fmt.Printf("#%d, %s (file %s)\r\n", m.Ver, m.Title, m.FileName)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
