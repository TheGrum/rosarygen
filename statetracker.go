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
	PrayerName    string

	OutputFileNum int
	InputFileNum  int
	GroupNum      int
	MysteryNum    int
	PrayerNum     int
	HailMaryNum   int

	OutputFilenameTemplate string
	LastFilename           string
}

func NewStateTracker(idirs []string, odir string, outputFilename string, format string) *StateTracker {
	return &StateTracker{
		InputDirs: idirs,
		OutputDir: odir,
		Format:    format,

		Group:         "Preamble",
		DecadeNumWord: "",
		Mystery:       "",
		Prayer:        "",
		PrayerName:    "",

		OutputFileNum: 0,
		InputFileNum:  0,
		GroupNum:      1,
		MysteryNum:    0,
		PrayerNum:     0,
		HailMaryNum:   0,

		OutputFilenameTemplate: outputFilename,
		LastFilename:           "",
	}

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

func (s *StateTracker) NumWord(num int) string {
	switch num {
	case 0:
		return ""
	case 1:
		return "First"
	case 2:
		return "Second"
	case 3:
		return "Third"
	case 4:
		return "Fourth"
	case 5:
		return "Fifth"
	case 6:
		return "Sixth"
	case 7:
		return "Seventh"
	case 8:
		return "Eighth"
	case 9:
		return "Ninth"
	case 10:
		return "Tenth"
	default:
		return strconv.Itoa(num)
	}
}

func (s *StateTracker) SetDecadeNumWord(num int) {
	s.DecadeNumWord = s.NumWord(num)
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

func (s *StateTracker) XofGroup() string {
	switch s.Group {
	case "Preamble":
		return "Preamble"
	case "Postamble":
		return "Postamble"
	default:
		return s.DecadeNumWord + " Of " + s.Group
	}
}

func (s *StateTracker) FileNum() int {
	return s.OutputFileNum
}

func (s *StateTracker) ZeroNum(num int, zeroes int) string {
	return fmt.Sprintf("%0"+strconv.Itoa(zeroes)+"d", num)
}

func (s *StateTracker) ZeroFileNum() string {
	return fmt.Sprintf("%03d", s.OutputFileNum)
}

func (s *StateTracker) CDTrack() string {
	return fmt.Sprintf("%02d %s", s.OutputFileNum, s.XthGroupMystery())
}
