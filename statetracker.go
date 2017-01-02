package rosarygen

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

type StateTracker struct {
	InputDirs []string
	OutputDir string
	Format    string

	Group         string // Preamble/[Group]/Postamble
	DecadeNumWord string
	Mystery       string
	MysteryPhrase string
	Prayer        string

	OutputFileNum int
	InputFileNum  int
	GroupNum      int
	MysteryNum    int
	PrayerNum     int
	HailMaryNum   int

	OutputFilenameTemplate string
	LastFilename           string
}

func (s *StateTracker) UpdateFilename() (filenameChanged bool) {
	temp := filepath.Join(s.OutputDir, s.Apply(s.OutputFilenameTemplate)+"."+s.Format)
	if temp != s.LastFilename {
		s.OutputFileNum += 1
		s.LastFilename = filepath.Join(s.OutputDir, s.Apply(s.OutputFilenameTemplate)+"."+s.Format)
		return true
	}
	return false

}

func (s *StateTracker) Apply(name string) string {
	if strings.Contains(name, "{{") {
		var out bytes.Buffer
		t := template.Must(template.New(".").Parse(name))
		if err := t.Execute(&out, s); err != nil {
			log.Fatalf("Error '%v' parsing template '%v'", err, name)
		}
		return out.String()
	} else {
		return name
	}
}

// Searches input dirs in order specified for the file
func (s *StateTracker) MatchActualFile(filename string) (string, error) {
	fname := filename + "." + s.Format
	for _, p := range s.InputDirs {
		t := filepath.Join(p, fname)
		if _, err := os.Stat(t); err == nil {
			return t, nil
		}
	}
	return fname, errors.New(fmt.Sprintf("File '%v' was not found in any input directory.", fname))
}

func (s *StateTracker) SetDecadeNumWord(num int) {
	switch num {
	case 0:
		s.DecadeNumWord = ""
	case 1:
		s.DecadeNumWord = "First"
	case 2:
		s.DecadeNumWord = "Second"
	case 3:
		s.DecadeNumWord = "Third"
	case 4:
		s.DecadeNumWord = "Fourth"
	case 5:
		s.DecadeNumWord = "Fifth"
	case 6:
		s.DecadeNumWord = "Sixth"
	case 7:
		s.DecadeNumWord = "Seventh"
	case 8:
		s.DecadeNumWord = "Eighth"
	case 9:
		s.DecadeNumWord = "Ninth"
	case 10:
		s.DecadeNumWord = "Tenth"
	default:
		s.DecadeNumWord = strconv.Itoa(num)
	}
}

func (s *StateTracker) XthGroupMystery() string {
	switch s.Group {
	case "Preamble":
		return "Preamble"
	case "Postamble":
		return "Postamble"
	default:
		return s.DecadeNumWord + " " + s.Group + " Mystery"
	}
}

func (s *StateTracker) CDTrack() string {
	return fmt.Sprintf("%02d %s", s.OutputFileNum, s.XthGroupMystery())
}
