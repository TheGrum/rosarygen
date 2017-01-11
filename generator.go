package rosarygen

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml"
)

type Generator struct {
	Prayers    map[string]*Prayer
	Mysteries  map[string]*Mystery
	Groups     map[string]*Group
	Structures map[string]*Structure
	Options    *Options
}

// NewGenerator creates and loads a Generator
// from .toml files
func NewGenerator() *Generator {
	g := &Generator{}
	g.Init()
	return g
}

// Init loads Generator with data from .toml files
// prayers.toml
// structures.toml
// options.toml
func (g *Generator) Init() {
	prayerconfig, err := toml.LoadFile("prayers.toml")
	if err != nil {
		fmt.Println("Error reading prayers.toml: ", err.Error())
		return
	} else {
		g.Prayers = ParsePrayers(prayerconfig)
		g.Mysteries = ParseMysteries(prayerconfig)
		g.Groups = ParseGroups(prayerconfig)
	}
	structureconfig, err := toml.LoadFile("structures.toml")
	if err != nil {
		fmt.Println("Error reading structures.toml: ", err.Error())
		return
	} else {
		g.Structures = ParseStructures(structureconfig)
	}
	optionconfig, err := toml.LoadFile("options.toml")
	if err != nil {
		fmt.Println("Error reading options.toml: ", err.Error())
		g.Options = &Options{}
	} else {
		g.Options = ParseOptions(optionconfig)
		// now mixin any local redefinitions, extra prayers, etc:
		g.Prayers = MergePrayers(g.Prayers, ParsePrayers(optionconfig))
		g.Mysteries = MergeMysteries(g.Mysteries, ParseMysteries(optionconfig))
		g.Groups = MergeGroups(g.Groups, ParseGroups(optionconfig))
		g.Structures = MergeStructures(g.Structures, ParseStructures(optionconfig))
	}

}

func (g *Generator) FindMystery(mystery string) *Mystery {
	for _, k := range g.Mysteries {
		if strings.ToLower(k.Name) == mystery {
			return k
		}
	}
	return nil
}

func (g *Generator) GroupsForRosary(groups string, mysteries string) []string {
	switch strings.ToLower(groups) {
	case "all":
		return []string{"joyful", "luminous", "sorrowful", "glorious"}
	case "old":
		return []string{"joyful", "sorrowful", "glorious"}
	case "custom":
		return strings.Split(strings.ToLower(mysteries), ",")
	default:
		return []string{strings.ToLower(groups)}
	}
}

func (g *Generator) NewRosary(structure string, groups ...string) *Rosary {
	actualGroups := []*Group{}
	for _, v := range groups {
		group, ok := g.Groups[v]
		if ok {
			actualGroups = append(actualGroups, group)
		} else {
			mystery := g.FindMystery(v)
			if mystery != nil {
				newgroup := NewGroup(-1, "Custom", "Custom")
				newgroup.AddMystery(mystery.Num)
				actualGroups = append(actualGroups, newgroup)
			}
		}
	}
	_, ok := g.Structures[structure]
	if !ok {
		_, ok := g.Prayers[structure]
		if !ok {
			if strings.Contains(structure, "[") {
				// Getting silly now - we are trying to define a structure on the fly? Oh well, handle it if we can.
				structure = strings.Replace(structure, "[", "", -1)
				structure = strings.Replace(structure, "]", "", -1)
				structure = strings.Replace(structure, "\"", "", -1)
				return NewRosary(StructureForPrayers(structure), actualGroups, g.Mysteries, g.Prayers)
			} else {
				log.Fatalf("Structure %s not found.\n", structure)
			}
		}
		// So, it's a prayer. One lonely single prayer. Let's mock up a structure and run with it.
		return NewRosary(StructureForPrayer(structure), actualGroups, g.Mysteries, g.Prayers)

	}
	return NewRosary(g.Structures[structure], actualGroups, g.Mysteries, g.Prayers)
}

func (g *Generator) RenderList(reader io.Reader) {
	var params map[string]string
	var s *StateTracker
	var pieces []string
	var pair []string

	params = map[string]string{
		"idirs":     "data",
		"odir":      "output",
		"ofilename": "{{.GroupNum}} {{.Group}} Mysteries",
		"groups":    "All",
		"mysteries": "",
		"structure": "basic",
		"format":    "wav",
		"gap":       "5",
	}

	s = NewStateTracker(nil, "", "", "")

	render := false
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "=") {
			// If there are equal signs, treat it like a command line, expect ALL items as parameter=value pairs, separated by spaces
			line = strings.Replace(line, "\\ ", "|", -1)
			pieces = strings.Split(line, " ")
			for _, piece := range pieces {
				piece = strings.Replace(piece, "|", " ", -1)
				pair = strings.SplitN(piece, "=", 2)
				if len(pair) > 1 {
					params[pair[0]] = pair[1]
					if pair[0] == "structure" {
						// we set the structure, so render it
						// otherwise we'll render only if we see Render command
						render = true
					} else if pair[0] == "outputfilenum" || pair[0] == "filenum" {
						// resetting the filenumber
						fnum, err := strconv.Atoi(pair[1])
						if err == nil {
							s.OutputFileNum = fnum - 1
						}
					}
				} else if strings.ToLower(pair[0]) == "render" {
					render = true
				}

			}
		} else if strings.Contains(line, "[") {
			// They are passing a list of prayers
			params["structure"] = line
			render = true
		} else {
			// If there are no equal signs, assume we are taking
			// everything from previous lines, and this is just a structure name, or a structure,groups pair
			pair = strings.Split(line, ",")
			params["structure"] = pair[0]
			if len(pair) > 1 {
				params["groups"] = pair[1]
			}
			render = true
		}
		if render {
			inputdirs := strings.Split(params["idirs"], ",")

			s.InputDirs = inputdirs
			s.OutputDir = params["odir"]
			s.OutputFilenameTemplate = params["ofilename"]
			s.Format = params["format"]
			s.GroupNum = 0
			s.MysteryNum = 0
			r := g.NewRosary(params["structure"], g.GroupsForRosary(params["groups"], params["mysteries"])...)
			gap, err := strconv.Atoi(params["gap"])
			if err != nil {
				gap = 5
			}
			// By preparing and sending s here
			// OutputFileNums will increment across the entire list of renders
			r.RenderToFiles(inputdirs, params["odir"], params["ofilename"], params["format"], g.Options, gap, s)

		}
		render = false
	}
}
