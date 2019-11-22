package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"

	"github.com/quasoft/pgmig/mig"

	// Import postgres DB driver
	_ "github.com/lib/pq"
)

// Session represents a user session to a specific PostgreSQL database
type Session struct {
	Host          string
	Port          string
	Database      string
	Username      string
	SslMode       string
	ChangelogName string
	Interactive   bool
	db            *sql.DB
}

// NewSession creates a new database session object
func NewSession() *Session {
	return &Session{}
}

// Connect creates a new connection to the database and makes sure it is responding by pinging it.
func (s *Session) Connect() error {
	// Build connection string
	password := getPassword(s.Interactive)
	connStr := buildConnString(s.Host, s.Port, s.Database, s.Username, password, s.SslMode)

	// Open connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("could not open DB connection: %v", err)
	}

	// Test DB connection (ping)
	var dummy string
	err = db.QueryRow("SELECT 1;").Scan(&dummy)
	if err != nil || dummy != "1" {
		db.Close()
		return fmt.Errorf("could not ping DB: %v", err)
	}
	s.db = db

	return nil
}

// Disconnect closes the database connection
func (s *Session) Disconnect() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// EnsureChangelogExists creates the changelog table if it does not exist
func (s *Session) EnsureChangelogExists() error {
	// TODO: Remove unused fields from table structure
	sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (
		id serial,
		version integer NOT NULL,
		file_name varchar(2048) NOT NULL,
		applied_by varchar(100) NOT NULL DEFAULT CURRENT_USER,
		date_time timestamp without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
		state bool NOT NULL DEFAULT false,
		CONSTRAINT "%s_pkey" PRIMARY KEY(id),
		CONSTRAINT "%s_version_unique" UNIQUE(version)
		)`,
		sanitizeIdentifier(s.ChangelogName),
		sanitizeIdentifier(s.ChangelogName),
		sanitizeIdentifier(s.ChangelogName),
	)
	_, err := s.db.Exec(sql)
	return err
}

// InsertLog records the migration in the changelog
func (s *Session) InsertLog(m mig.File) error {
	sql := fmt.Sprintf(
		`INSERT INTO %s(
			version,
			file_name
		)
		VALUES(
			$1,
			$2
		)`,
		sanitizeIdentifier(s.ChangelogName),
	)
	_, err := s.db.Exec(sql, m.Ver, m.Path)
	return err
}

// lastMigratedVer returns the version of the last migration file that was applied successfully,
// according to the changelog table
func (s *Session) lastMigratedVer() (int, error) {
	query := fmt.Sprintf(
		`SELECT COALESCE(MAX(version), 0) FROM "%s" WHERE state = true`,
		sanitizeIdentifier(s.ChangelogName),
	)
	var migVer int
	err := s.db.QueryRow(query).Scan(&migVer)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("could not get version of last migration from changelog table: %v", err)
	}

	return migVer, nil
}

// wasApplied checks if the specified migration was applied to DB
func (s *Session) wasApplied(migVer int) (bool, error) {
	sql := fmt.Sprintf(
		`SELECT COUNT(*) FROM "%s" WHERE state = true AND version = $1`,
		sanitizeIdentifier(s.ChangelogName),
	)
	var cnt int
	err := s.db.QueryRow(sql, migVer).Scan(&cnt)
	if err != nil {
		return false, fmt.Errorf("could not check in changelog %s if migration #%d was applied: %v", s.ChangelogName, migVer, err)
	}

	return cnt > 0, nil
}

func (s *Session) Apply(m mig.File) error {
	bytes, err := ioutil.ReadFile(m.Path)
	if err != nil {
		return fmt.Errorf("could not read migration file %s: %v", m.FileName, err)
	}

	sql := string(bytes)
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("could not open transaction: %v", err)
	}

	_, err = s.db.Exec(sql)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("could not execute migration #%d from file %s: %v", m.Ver, m.FileName, err)
	}

	sql = fmt.Sprintf(
		`UPDATE %s SET state = $1 WHERE version = $2`,
		sanitizeIdentifier(s.ChangelogName),
	)
	_, err = s.db.Exec(sql, true, m.Ver)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("could not mark migration #%d for file %s as completed in DB, rolling back...: %v", m.Ver, m.FileName, err)
	}

	return tx.Commit()
}

// PendingMigrations returns a list of migration files that have not been applied yet, according to the changelog
func (s *Session) PendingMigrations(dir *mig.Dir) ([]mig.File, error) {
	// TODO: Use version of last applied migration and only check later migrations
	// Get version of last applied migration
	lastVer, err := s.lastMigratedVer()
	if err != nil {
		return nil, fmt.Errorf("could not determine version of last migration: %v", err)
	}

	allMigrations, err := dir.Migrations()
	if err != nil {
		return nil, err
	}

	var pending []mig.File
	for _, m := range allMigrations {
		if m.Ver <= lastVer {
			continue
		}
		// Make sure the specific migration was not applied
		applied, err := s.wasApplied(m.Ver)
		if err != nil {
			return pending, err
		}
		if !applied {
			pending = append(pending, m)
		}
	}
	return pending, nil
}
