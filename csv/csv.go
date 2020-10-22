package csv

import (
	"log"
	"os"
	"strings"
)

// File is ...
type File struct {
	FileName string
	FilePath string
	Location string
	OsFile   *os.File
}

// Add returns nil
func (f *File) Add(items []string) {
	line := strings.Join(items, ";")
	f.write(line)
}

func (f *File) write(line string) {
	if _, err := f.OsFile.WriteString(line + "\n"); err != nil {
		log.Println(err)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// GetCSV returns nil
func GetCSV(path string, filename string) *File {
	path = strings.Replace(path, "\\", "/", -1)

	if !strings.Contains(filename, ".csv") {
		filename = filename + ".csv"
	}

	if path[len(path)-1:] != "/" {
		path = path + "/"
	}

	location := path + filename

	if fileExists(location) {
		os.Remove(location)
	}

	f, err := os.OpenFile(location, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Println(err)
	}

	// defer f.Close()

	csvFile := File{FileName: filename, FilePath: path, Location: location, OsFile: f}

	return &csvFile
}
