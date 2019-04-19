package mig

// File represents an SQL migration file with filename in format "0001_Initial_db_structure.sql"
type File struct {
	Ver      int
	Title    string
	FileName string
}

// NewFile creates a new migration file object
func NewFile(fileName string, ver int) *File {
	return &File{Ver: ver, FileName: fileName}
}
