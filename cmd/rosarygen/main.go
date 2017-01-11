package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/TheGrum/rosarygen"
	"github.com/vharitonsky/iniflags"
)

var (
	idirs           = flag.String("idirs", "data", "Comma separated list of audio data folders, searched in order given")
	odir            = flag.String("odir", "output", "output folder")
	ofilename       = flag.String("ofilename", "{{.GroupNum}} {{.Group}} Mysteries", "Output filename template. Available fields: Group, GroupNum, Mystery, MysteryNum, Prayer, PrayerNum, OutputFileNum, XthGroupMystery")
	mysteryGroups   = flag.String("groups", "All", "Mystery decade groupings to generate. Possible values: All, Old (All excluding Luminous), Joyful, Luminous, Sorrowful, Glorious, and Custom (specify list of mysteries with mysteries)")
	customMysteries = flag.String("mysteries", "", "List of mysteries to use in place of group. Use ListMysteries to see options.")
	structure       = flag.String("structure", "basic", "Rosary structure to use. Use ListStructures to see options.")
	format          = flag.String("format", "wav", "wav or flac")
	gap             = flag.Int("gapLength", 5, "tenths of seconds of silence to add between prayers")
)

func main() {
	iniflags.Parse()

	g := rosarygen.NewGenerator()

	if (flag.NArg()) > 0 {
		switch flag.Arg(0) {
		case "PrintOptions":
			for k, v := range g.Options.Options {
				fmt.Printf("%v: %v\n", k, v)
			}
		case "ListPrayers":
			keys := make([]string, len(g.Prayers), len(g.Prayers))
			i := 0
			for k, _ := range g.Prayers {
				keys[i] = k
				i += 1
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("%v: %v '%v'\n", k, g.Prayers[k].Name, g.Prayers[k].Desc)
			}
			return
		case "ListGroups":
			keys := make([]string, len(g.Groups), len(g.Groups))
			i := 0
			for k, _ := range g.Groups {
				keys[i] = k
				i += 1
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("%v: %v\n", k, g.Groups[k].Name)
			}
			return
		case "ListMysteries":
			for i := 1; i <= len(g.Mysteries); i++ {
				fmt.Printf("%v: %v\n", i, g.Mysteries[strconv.Itoa(i)].Name)
			}
			return
		case "ListStructures":
			keys := make([]string, len(g.Structures), len(g.Structures))
			i := 0
			for k, _ := range g.Structures {
				keys[i] = k
				i += 1
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("%v: %v\n", k, g.Structures[k].Name)
			}
			return
		}

		if flag.Arg(0) == "RenderList" {
			// Here we are rendering a whole
			// series of rosaries/prayers
			var stream io.Reader
			var err error
			if flag.NArg() > 1 {
				if flag.Arg(1) == "-" {
					stream = os.Stdin
				} else {
					stream, err = os.Open(flag.Arg(1))
					if err != nil {
						log.Fatal(err)
					}
				}
			} else {
				stream = os.Stdin
			}
			g.RenderList(stream)
			return
		}

		// If we get here, they want something from a calculated rosary, so prepare it
		var r *rosarygen.Rosary
		r = g.NewRosary(*structure, g.GroupsForRosary(*mysteryGroups, *customMysteries)...)
		//		switch strings.ToLower(*mysteryGroups) {
		//		case "all":
		//			r = g.NewRosary(*structure, "joyful", "luminous", "sorrowful", "glorious")
		//		case "old":
		//			r = g.NewRosary(*structure, "joyful", "sorrowful", "glorious")
		//		case "custom":
		//			r = g.NewRosary(*structure, strings.Split(strings.ToLower(*customMysteries), ",")...)
		//		default:
		//			r = g.NewRosary(*structure, strings.ToLower(*mysteryGroups))
		//		}
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
			r.ForEachFile(inputdirs, *odir, *ofilename, *format, g.Options, rosarygen.GetBadFilenamesFunc(onBadFileFunc), nil)
		case "ActualFiles":
			r.ForEachFile(inputdirs, *odir, *ofilename, *format, g.Options, rosarygen.PrintActualFilename, nil)
		case "Render":
			if *format == "flac" {
				log.Fatal("Not implemented yet.")
			}
			r.RenderToFiles(inputdirs, *odir, *ofilename, *format, g.Options, *gap, nil)

		}
	}
}
