package rosarygen

import (
	"log"
	"strconv"

	"github.com/pelletier/go-toml"
)

func ParsePrayers(data *toml.TomlTree) map[string]*Prayer {
	prayers := make(map[string]*Prayer)
	//	pbag, _ := data.Query("$.prayers")
	pbag := data.Get("prayers").(*toml.TomlTree)
	for _, p := range pbag.Keys() {
		po := pbag.Get(p).(*toml.TomlTree)
		prayers[p] = ParsePrayer(p, po)
		//fmt.Printf("beep: %v,%v,%v\n", i, p, pbag.Get(p))

	}
	return prayers
}

func ParsePrayer(key string, po *toml.TomlTree) *Prayer {
	np := NewPrayer(key, po.Get("name").(string))
	if po.Has("filename") {
		np.SetFilename(po.Get("filename").(string))
	}
	if po.Has("filenames") {
		for _, f := range po.Get("filenames").([]interface{}) {
			np.AddFilename(f.(string))
		}
	}
	if po.Has("text") {
		np.SetText(po.Get("text").(string))
	}
	if po.Has("desc") {
		np.SetDesc(po.Get("desc").(string))
	}
	if po.Has("options") {
		list := po.Get("options").(*toml.TomlTree)
		for i := 1; i < len(list.Keys()); i++ {
			po2 := list.Get(strconv.Itoa(i)).(*toml.TomlTree)
			np.AddOption(ParsePrayer(key, po2))
		}
	}
	return np
}

func ParseStructures(data *toml.TomlTree) map[string]*Structure {
	structures := make(map[string]*Structure)
	//	pbag, _ := data.Query("$.prayers")
	sbag := data.Get("structure").(*toml.TomlTree)
	for _, s := range sbag.Keys() {
		so := sbag.Get(s).(*toml.TomlTree)
		structures[s] = ParseStructure(s, so)
		//fmt.Printf("beep: %v,%v,%v\n", i, p, pbag.Get(p))

	}
	return structures
}

func ParseStructure(key string, so *toml.TomlTree) *Structure {
	ns := NewStructure(key, so.Get("name").(string))
	if so.Has("preamble") {
		for _, p := range so.Get("preamble").([]interface{}) {
			ns.AddPreamble(p.(string))
		}
	}

	if so.Has("group") {
		for _, p := range so.Get("group").([]interface{}) {
			ns.AddGroup(p.(string))
		}
	}

	if so.Has("mystery") {
		for _, p := range so.Get("mystery").([]interface{}) {
			ns.AddMystery(p.(string))
		}
	}

	if so.Has("postamble") {
		for _, p := range so.Get("postamble").([]interface{}) {
			ns.AddPostamble(p.(string))
		}
	}
	return ns
}

func ParseMysteries(data *toml.TomlTree) map[string]*Mystery {
	mysteries := make(map[string]*Mystery)
	//	pbag, _ := data.Query("$.prayers")
	sbag := data.Get("mystery").(*toml.TomlTree)
	for _, s := range sbag.Keys() {
		so := sbag.Get(s).(*toml.TomlTree)
		num, err := strconv.Atoi(s)
		if err != nil {
			log.Fatalf("Could not parse '%v' as an int.", s)
		}
		mysteries[s] = ParseMystery(num, so)
		//fmt.Printf("beep: %v,%v,%v\n", i, p, pbag.Get(p))

	}
	return mysteries
}

func ParseMystery(num int, so *toml.TomlTree) *Mystery {
	ns := NewMystery(num, so.Get("name").(string))
	if so.Has("desc") {
		ns.SetDesc(so.Get("desc").(string))
	}

	return ns
}

func ParseGroups(data *toml.TomlTree) map[string]*Group {
	groups := make(map[string]*Group)
	//	pbag, _ := data.Query("$.prayers")
	sbag := data.Get("group").(*toml.TomlTree)
	for _, s := range sbag.Keys() {
		so := sbag.Get(s).(*toml.TomlTree)
		groups[s] = ParseGroup(s, so)
		//fmt.Printf("beep: %v,%v,%v\n", i, p, pbag.Get(p))

	}
	return groups
}

func ParseGroup(key string, so *toml.TomlTree) *Group {
	ns := NewGroup(int(so.Get("order").(int64)), key, so.Get("name").(string))
	for _, p := range so.Get("mysteries").([]interface{}) {
		ns.AddMystery(int(p.(int64)))
	}

	return ns
}

func ParseOptions(data *toml.TomlTree) *Options {
	options := NewOptions()
	sbag := data.Get("options").(*toml.TomlTree)
	for _, s := range sbag.Keys() {
		so := sbag.Get(s).(int64)
		options.AddOption(s, int(so))

	}
	return options
}

// Merge functions return a with b's entries added, possibly replacing a's entries
func MergePrayers(a map[string]*Prayer, b map[string]*Prayer) map[string]*Prayer {
	for k, v := range b {
		a[k] = v
	}
	return a
}

func MergeStructures(a map[string]*Structure, b map[string]*Structure) map[string]*Structure {
	for k, v := range b {
		a[k] = v
	}
	return a
}

func MergeMysteries(a map[string]*Mystery, b map[string]*Mystery) map[string]*Mystery {
	for k, v := range b {
		a[k] = v
	}
	return a
}

func MergeGroups(a map[string]*Group, b map[string]*Group) map[string]*Group {
	for k, v := range b {
		a[k] = v
	}
	return a
}
