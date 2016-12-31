package rosarygen

type OptionProvider interface {
	GetOption(prayer string) int
}

type Options struct {
	Options map[string]int
}

func NewOptions() *Options {
	return &Options{
		Options: make(map[string]int, 10),
	}
}

func (o *Options) AddOption(prayer string, choice int) {
	o.Options[prayer] = choice
}

func (o *Options) GetOption(prayer string) int {
	x, ok := o.Options[prayer]
	if ok {
		return x
	} else {
		return 1
	}
}
