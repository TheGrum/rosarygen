package rosarygen

type Structure struct {
	Key       string
	Name      string
	Preamble  []string
	Group     []string
	Mystery   []string
	Postamble []string
}

func NewStructure(key string, name string) *Structure {
	return &Structure{
		Key:       key,
		Name:      name,
		Preamble:  make([]string, 0, 10),
		Group:     make([]string, 0, 10),
		Mystery:   make([]string, 0, 20),
		Postamble: make([]string, 0, 10),
	}
}

func (s *Structure) AddPreamble(preamble string) {
	s.Preamble = append(s.Preamble, preamble)
}

func (s *Structure) AddGroup(group string) {
	s.Group = append(s.Group, group)
}

func (s *Structure) AddMystery(mystery string) {
	s.Mystery = append(s.Mystery, mystery)
}

func (s *Structure) AddPostamble(postamble string) {
	s.Postamble = append(s.Postamble, postamble)
}
