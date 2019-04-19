package mig

import (
	"testing"
)

func TestParseFileName(t *testing.T) {
	var tests = []struct {
		fileName string
		want     File
		noError  bool
	}{
		{"1_Initial_db_structure.sql", File{FileName: "1_Initial_db_structure.sql", Ver: 1, Title: "Initial db structure"}, true},
		{"0042_Create_test_table.sql", File{FileName: "0042_Create_test_table.sql", Ver: 42, Title: "Create test table"}, true},
		{"MigrationFileWithNoVersion.sql", File{}, false},
		{"008.sql", File{}, false},
		{"0016MigrationWithNoSeparator.sql", File{}, false},
		{"0000_Migration_with_zero_version.sql", File{FileName: "0000_Migration_with_zero_version.sql", Ver: 0, Title: "Migration with zero version"}, true},
	}

	dir := NewDir()
	for _, tt := range tests {
		got, err := dir.parseFileName(tt.fileName)

		if err != nil {
			if tt.noError {
				t.Fatalf("parseFileName(%s) returned error %v", tt.fileName, err)
			}
			continue
		}
		if err == nil && !tt.noError {
			t.Fatalf("parseFileName(%s) should have returned an error", tt.fileName)
		}
		if err == nil && got == nil {
			t.Fatalf("parseFileName(%s) should have either returned a structure or an error", tt.fileName)
		}

		if got.FileName != tt.want.FileName {
			t.Errorf("parseFileName(%s): got filename=%q, want filename=%q", tt.fileName, got.FileName, tt.want.FileName)
		}
		if got.Ver != tt.want.Ver {
			t.Errorf("parseFileName(%s): got ver=%d, want ver=%d", tt.fileName, got.Ver, tt.want.Ver)
		}
		if got.Title != tt.want.Title {
			t.Errorf("parseFileName(%s): got title=%q, want title=%q", tt.fileName, got.Title, tt.want.Title)
		}
	}
}
