package rosarygen

import (
	"fmt"
	"log"
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
		log.Fatalf("Structure %s not found.\n")
	}
	return NewRosary(g.Structures[structure], actualGroups, g.Mysteries, g.Prayers)
}
