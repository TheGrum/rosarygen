package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/TheGrum/rosarygen"
	"github.com/pelletier/go-toml"
	"github.com/vharitonsky/iniflags"
)

var (
	idirs           = flag.String("idirs", "data", "Comma separated list of audio data folders, searched in order given")
	odir            = flag.String("odir", "output", "output folder")
	ofilename       = flag.String("ofilename", "{{.GroupNum}} {{.Group}} Mysteries", "Output filename template. Available fields: GroupNum, Mystery, Prayer")
	mysteryGroups   = flag.String("groups", "All", "Mystery decade groupings to generate. Possible values: All, Old (All excluding Luminous), Joyful, Luminous, Sorrowful, Glorious, and Custom (specify list of mysteries with mysteries)")
	customMysteries = flag.String("mysteries", "", "List of mysteries to use in place of group. Use ListMysteries to see options.")
	structure       = flag.String("structure", "basic", "Rosary structure to use. Use ListStructures to see options.")
	format          = flag.String("format", "wav", "wav or flac")
	gap             = flag.Int("gapLength", 5, "seconds of silence to add between prayers")
)

func main() {
	iniflags.Parse()

	var (
		prayers    map[string]*rosarygen.Prayer
		mysteries  map[string]*rosarygen.Mystery
		groups     map[string]*rosarygen.Group
		structures map[string]*rosarygen.Structure
		options    *rosarygen.Options
	)

	prayerconfig, err := toml.LoadFile("prayers.toml")
	if err != nil {
		fmt.Println("Error reading prayers.toml: ", err.Error())
		return
	} else {
		prayers = rosarygen.ParsePrayers(prayerconfig)
		mysteries = rosarygen.ParseMysteries(prayerconfig)
		groups = rosarygen.ParseGroups(prayerconfig)
	}
	structureconfig, err := toml.LoadFile("structures.toml")
	if err != nil {
		fmt.Println("Error reading structures.toml: ", err.Error())
		return
	} else {
		structures = rosarygen.ParseStructures(structureconfig)
	}
	optionconfig, err := toml.LoadFile("options.toml")
	if err != nil {
		fmt.Println("Error reading structures.toml: ", err.Error())
		return
	} else {
		options = rosarygen.ParseOptions(optionconfig)
	}

	if (flag.NArg()) > 0 {
		switch flag.Arg(0) {
		case "PrintOptions":
			for k, v := range options.Options {
				fmt.Printf("%v: %v\n", k, v)
			}
		case "ListPrayers":
			keys := make([]string, len(prayers), len(prayers))
			i := 0
			for k, _ := range prayers {
				keys[i] = k
				i += 1
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("%v: %v '%v'\n", k, prayers[k].Name, prayers[k].Desc)
			}
			return
		case "ListGroups":
			keys := make([]string, len(groups), len(groups))
			i := 0
			for k, _ := range groups {
				keys[i] = k
				i += 1
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("%v: %v\n", k, groups[k].Name)
			}
			return
		case "ListMysteries":
			for i := 1; i <= len(mysteries); i++ {
				fmt.Printf("%v: %v\n", i, mysteries[strconv.Itoa(i)].Name)
			}
			return
		case "ListStructures":
			keys := make([]string, len(structures), len(structures))
			i := 0
			for k, _ := range structures {
				keys[i] = k
				i += 1
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("%v: %v\n", k, structures[k].Name)
			}
			return
		}

		// If we get here, they want something from a calculated rosary, so prepare it
		var r *rosarygen.Rosary
		switch *mysteryGroups {
		case "All":
			r = rosarygen.NewRosary(structures[*structure], []*rosarygen.Group{groups["joyful"], groups["luminous"], groups["sorrowful"], groups["glorious"]}, mysteries, prayers)
		case "Old":
			r = rosarygen.NewRosary(structures[*structure], []*rosarygen.Group{groups["joyful"], groups["sorrowful"], groups["glorious"]}, mysteries, prayers)
		case "Custom":
			log.Fatal("Not Implemented Yet.")
		default:
			r = rosarygen.NewRosary(structures[*structure], []*rosarygen.Group{groups[strings.ToLower(*mysteryGroups)]}, mysteries, prayers)
		}
		if r == nil {
			log.Fatal("No Rosary generated.")
		}
		inputdirs := strings.Split(*idirs, ",")
		switch flag.Arg(0) {
		case "Prayers":
			for _, p := range r.GetPrayers() {
				fmt.Printf("%v: %v\n", p.Name, p.Filename)
			}
		case "MissingFiles":
			onBadFileFunc := func(filename string, err error) {
				fmt.Printf("%v: %v\n", filename, err)
			}
			r.ForEachFile(inputdirs, *odir, *ofilename, *format, options, rosarygen.GetBadFilenamesFunc(onBadFileFunc))
		case "Render":
			ch := make(chan *rosarygen.FileStack)
			go r.ForEachFile(inputdirs, *odir, *ofilename, *format, options, rosarygen.GetFiles(ch))
			switch *format {
			case "wav":
				for f := range ch {
					f.RenderWav(*gap)
					//fmt.Println(f)
				}
			case "flac":
				//				for f := range ch {
				//					f.RenderFlac(*gap)
				//					fmt.Println(f)
				//				}
			}
		}
	}
}
