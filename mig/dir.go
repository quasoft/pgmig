package mig

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Dir represents an abstraction for listing migration files in a directory
type Dir struct {
	Path string
}

// NewDir creates a new object for listing migration files in a directory
func NewDir() *Dir {
	return &Dir{}
}

// parseFileName parses file names in format "NNN_Title_with_underscores.sql" and
// returns a mig.File structure with results.
func (d *Dir) parseFileName(fileName string) (*File, error) {
	parts := strings.SplitN(fileName, "_", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("filename %s is not in expected format", fileName)
	}
	ver, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("filename %s is not in expected format", fileName)
	}
	m := NewFile(fileName, ver)
	m.Title = strings.Replace(parts[1], "_", " ", -1)
	m.Title = strings.TrimSuffix(m.Title, filepath.Ext(m.Title))
	return m, nil
}

// files returns names of all files found in the specified directory
func (d *Dir) files() ([]string, error) {
	entries, err := ioutil.ReadDir(d.Path)
	if err != nil {
		return nil, fmt.Errorf("could not list files %s: %v", d.Path, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		files = append(files, e.Name())
	}

	return files, nil
}

// Migrations returns a list of all migration files found in the specified directory,
// sorted by the migration version.
func (d *Dir) Migrations() ([]File, error) {
	files, err := d.files()
	if err != nil {
		return nil, err
	}

	var migrations []File
	for _, f := range files {
		m, err := d.parseFileName(f)
		if err != nil {
			return nil, err
		}
		// Check for migrations with duplicated version number
		for _, mm := range migrations {
			if mm.Ver == m.Ver {
				return nil, fmt.Errorf("found migrations with the same version #%d:\r\n- %s\r\n- %s", m.Ver, mm.FileName, m.FileName)
			}
		}
		m.Path = filepath.Join(d.Path, m.FileName)
		migrations = append(migrations, *m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Ver < migrations[j].Ver
	})
	return migrations, nil
}
