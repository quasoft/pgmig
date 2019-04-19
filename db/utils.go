package db

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func buildConnString(host, port, database, username, password, sslmode string) string {
	type param struct {
		key   string
		value string
	}
	connStr := ""

	if host != "" {
		connStr += "host=" + host + " "
	}
	if port != "" {
		connStr += "port=" + port + " "
	}
	if database != "" {
		connStr += "dbname=" + database + " "
	}
	if username != "" {
		connStr += "user=" + username + " "
	}
	if password != "" {
		connStr += "password=" + password + " "
	}
	if sslmode != "" {
		connStr += "sslmode=" + sslmode + " "
	}

	return connStr
}

func sanitizeIdentifier(identifier string) string {
	isQuoted := len(identifier) > 2 && identifier[0] == '"' && identifier[len(identifier)-1:] == `"`

	var allowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789@$"
	var result string
	for _, c := range identifier {
		if strings.Index(allowedChars, string(c)) > -1 {
			result += string(c)
		}
	}
	if isQuoted {
		result = `"` + result + `"`
	}
	return result
}

func quoteString(value string) string {
	return "'" + strings.Replace(value, "'", "''", -1) + "'"
}

func getPassword(interactive bool) string {
	password := os.Getenv("PG_PASSWORD")
	if password == "" && interactive {
		pwd, err := readPassword("Enter DB password: ")
		if err == nil {
			password = pwd
		}
	}
	return password
}

func readPassword(prompt string, args ...interface{}) (string, error) {
	fmt.Printf(prompt, args...)
	pwd, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("could not read password: %s", err)
	}
	fmt.Println()
	return string(pwd), nil
}
