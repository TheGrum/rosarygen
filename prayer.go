package rosarygen

import "fmt"

type (
	Prayer struct {
		Key       string
		Name      string
		Desc      string
		Text      string
		Filename  string
		Filenames []string
		Options   []*Prayer
	}
)

func NewPrayer(key string, name string) *Prayer {
	return &Prayer{
		Key:       key,
		Name:      name,
		Filenames: make([]string, 0, 1),
		Options:   make([]*Prayer, 0, 1),
	}
}

func (p *Prayer) SetName(name string) {
	p.Name = name
}

func (p *Prayer) SetText(text string) {
	p.Text = text
}

func (p *Prayer) SetDesc(desc string) {
	p.Desc = desc
}

func (p *Prayer) SetFilename(filename string) {
	p.Filename = filename
}

func (p *Prayer) AddOption(option *Prayer) {
	p.Options = append(p.Options, option)
}

func (p *Prayer) AddFilename(filename string) {
	p.Filenames = append(p.Filenames, filename)
}

func (p *Prayer) GetChosenFilenames(o OptionProvider) []string {
	r := make([]string, 0, 1)
	i := o.GetOption(p.Key)
	if len(p.Options) > 0 {
		if i <= len(p.Options) {
			po := p.Options[i-1]
			r = append(r, po.GetChosenFilenames(o)...)
		} else {
			r = append(r, p.Filename)
			return r
		}
	} else {
		// Either no options, or we are already in an option leaf
		if len(p.Filenames) > 0 {
			return p.Filenames
		} else {
			r = append(r, p.Filename)
		}
	}
	return r
}

func (p *Prayer) ForEachFile(o OptionProvider, s *StateTracker, f func(filename string, p *Prayer, s *StateTracker)) {
	s.PrayerNum += 1
	r := p.GetChosenFilenames(o)
	for _, file := range r {
		s.InputFileNum += 1
		if p.Key == "hailmary" {
			s.HailMaryNum += 1
		}
		s.Prayer = p.Key
		ofile := s.Apply(file)
		f(ofile, p, s)
	}
}

// Utility functions to pass to ForEachFile
// Functions should handle being passed "", nil, nil as the last entry
func PrintFilename(filename string, p *Prayer, s *StateTracker) {
	if filename != "" {
		fmt.Println(filename)
	}
}

func GetFiles(ch chan *FileStack) func(filename string, p *Prayer, s *StateTracker) {
	var stack *FileStack
	return func(filename string, p *Prayer, s *StateTracker) {
		if stack == nil {
			s.UpdateFilename()
			stack = NewFileStack(s.LastFilename)
		} else {
			if filename == "" {
				// Last file
				ch <- stack
				close(ch)
				return
			}
			if s.UpdateFilename() {
				ch <- stack
				stack = NewFileStack(s.LastFilename)
			}
		}
		actual, err := s.MatchActualFile(filename)
		if err == nil {
			stack.AddFilename(actual)
		}
	}
}

func GetBadFilenamesFunc(f func(filename string, err error)) func(filename string, p *Prayer, s *StateTracker) {
	m := map[string]error{}
	return func(filename string, p *Prayer, s *StateTracker) {
		if filename == "" {
			return
		}
		actual, err := s.MatchActualFile(filename)
		if err != nil {
			_, ok := m[filename]
			if !ok {
				m[filename] = err
				f(actual, err)
			}
		}
	}
}
