package rosarygen

import (
	"fmt"
	"strconv"
)

type Rosary struct {
	Preamble  []*Prayer
	Decades   []*Decade
	Postamble []*Prayer
}

func NewRosary(structure *Structure, groups []*Group, mysteries map[string]*Mystery, prayers map[string]*Prayer) *Rosary {
	pergroup := []*Prayer{}
	r := &Rosary{
		Preamble:  []*Prayer{},
		Decades:   []*Decade{},
		Postamble: []*Prayer{},
	}
	for _, s := range structure.Preamble {
		r.Preamble = append(r.Preamble, prayers[s])
	}
	for _, s := range structure.Postamble {
		r.Postamble = append(r.Postamble, prayers[s])
	}
	for _, s := range structure.Group {
		pergroup = append(pergroup, prayers[s])
	}
	for _, g := range groups {
		nms := []*Mystery{}
		for _, mi := range g.Mysteries {
			m := mysteries[strconv.Itoa(mi)]
			nm := NewMystery(m.Num, m.Name)
			for _, s := range structure.Mystery {
				nm.Prayers = append(nm.Prayers, prayers[s])
			}
			nms = append(nms, nm)
		}
		r.Decades = append(r.Decades, NewDecade(g.Name, pergroup, nms...))

	}
	return r
}

func (r *Rosary) GetPrayers() []*Prayer {
	s := []*Prayer{}
	for _, p := range r.Preamble {
		s = append(s, p)
	}
	for _, d := range r.Decades {
		s = append(s, d.GetPrayers()...)
	}
	for _, p := range r.Postamble {
		s = append(s, p)
	}
	return s
}

func (r *Rosary) ForEachFile(idirs []string, odir string, outputFilename string, format string, o OptionProvider, f func(filename string, p *Prayer, s *StateTracker)) {
	s := &StateTracker{
		InputDirs: idirs,
		OutputDir: odir,
		Format:    format,

		Group:         "Preamble",
		DecadeNumWord: "",
		Mystery:       "",
		Prayer:        "",

		OutputFileNum: 0,
		InputFileNum:  0,
		GroupNum:      1,
		MysteryNum:    0,
		PrayerNum:     0,
		HailMaryNum:   0,

		OutputFilenameTemplate: outputFilename,
		LastFilename:           "",
	}

	for _, p := range r.Preamble {
		p.ForEachFile(o, s, f)
	}
	for _, d := range r.Decades {
		d.ForEachFile(o, s, f)
	}
	s.Group = "Postamble"
	s.GroupNum += 1
	s.MysteryNum = 0
	s.Mystery = ""
	for _, p := range r.Postamble {
		p.ForEachFile(o, s, f)
	}
	f("", nil, nil)
}

type Decade struct {
	Name      string
	Group     []*Prayer
	Mysteries []*Mystery
}

func NewDecade(name string, group []*Prayer, mysteries ...*Mystery) *Decade {
	return &Decade{
		Name:      name,
		Group:     group,
		Mysteries: mysteries,
	}
}

func (d *Decade) GetPrayers() []*Prayer {
	s := []*Prayer{}
	for _, p := range d.Group {
		s = append(s, p)
	}
	for _, m := range d.Mysteries {
		s = append(s, m.GetPrayers()...)
	}
	return s
}

func (d *Decade) ForEachFile(o OptionProvider, s *StateTracker, f func(filename string, p *Prayer, s *StateTracker)) {
	s.GroupNum += 1
	s.Group = d.Name
	s.HailMaryNum = 0
	i := 1
	s.SetDecadeNumWord(1)
	if len(d.Mysteries) > 0 {
		s.MysteryNum = d.Mysteries[0].Num

		s.Mystery = d.Mysteries[0].Name
		s.MysteryPhrase = s.Mystery + " Mystery"
	} else {
		s.MysteryNum = 0
		s.Mystery = ""
		s.MysteryPhrase = ""
	}
	for _, p := range d.Group {
		p.ForEachFile(o, s, f)
	}
	for _, m := range d.Mysteries {
		s.SetDecadeNumWord(i)
		m.ForEachFile(o, s, f)
		i += 1
	}
	s.MysteryNum = 0
	s.Mystery = ""
	s.MysteryPhrase = ""
	s.SetDecadeNumWord(0)
}

type Mystery struct {
	Num     int
	Name    string
	Desc    string
	Prayers []*Prayer
}

func NewMystery(num int, name string) *Mystery {
	return &Mystery{
		Num:     num,
		Name:    name,
		Prayers: make([]*Prayer, 0, 20),
	}
}

func (m *Mystery) SetDesc(desc string) {
	m.Desc = desc
}

func (m *Mystery) GetPrayers() []*Prayer {
	s := []*Prayer{}
	for _, p := range m.Prayers {
		s = append(s, p)
	}
	return s
}

func (m *Mystery) ForEachFile(o OptionProvider, s *StateTracker, f func(filename string, p *Prayer, s *StateTracker)) {
	s.MysteryNum = m.Num
	s.Mystery = m.Name
	s.MysteryPhrase = s.Mystery + " Mystery"
	s.HailMaryNum = 0
	for _, p := range m.Prayers {
		p.ForEachFile(o, s, f)
	}
	s.HailMaryNum = 0
}

type Group struct {
	Order     int
	Key       string
	Name      string
	Mysteries []int
}

func NewGroup(order int, key string, name string) *Group {
	return &Group{
		Order:     order,
		Key:       key,
		Name:      name,
		Mysteries: make([]int, 0, 5),
	}
}

func (g *Group) AddMystery(mystery int) {
	g.Mysteries = append(g.Mysteries, mystery)
}

func (g *Group) String() string {
	return fmt.Sprintf("%s: %s", g.Key, g.Name)
}
