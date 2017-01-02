package rosarygen

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"azul3d.org/audio.v1"
	"azul3d.org/audio/wav.v1"
	//	"github.com/azul3d/engine/audio/flac"
)

const sampleRate = 48000

type FileStack struct {
	OutputFilename string
	Filenames      []string
}

func NewFileStack(filename string) *FileStack {
	return &FileStack{
		OutputFilename: filename,
		Filenames:      []string{},
	}
}

func (f *FileStack) AddFilename(filename string) {
	f.Filenames = append(f.Filenames, filename)
}

//func (f *FileStack) RenderFlac(gap int) {
//
//	encoder, err := flac.NewEncoder(f.OutputFilename)
//}

func (f *FileStack) RenderWav(gap int) {
	if len(f.Filenames) < 1 {
		return
	}
	in, err := os.Open(f.Filenames[0])
	if err != nil {
		panic(err)
	}
	decoder, _, err := audio.NewDecoder(in)
	if err != nil {
		panic(err)
	}

	out, err := os.Create(f.OutputFilename)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := out.Close()
		if err != nil {
			panic(err)
		}
	}()
	config := decoder.Config()
	in.Close()

	encoder, err := wav.NewEncoder(out, config)

	// create buffer
	bufSize := 2 * config.SampleRate * config.Channels
	buf := make(audio.F64Samples, bufSize, bufSize)
	silence := make(audio.F64Samples, sampleRate/10)

	// build audio pipeline
	pipe := make(chan audio.Slice)
	var inChan <-chan audio.Slice = pipe

	// run the loop
	go func() {
		for _, filename := range f.Filenames {
			if strings.HasPrefix(filename, "silence:") {
				filename = strings.Replace(filename, "silence:", "", -1)
				seconds, err := strconv.Atoi(filename)
				if err == nil {
					for i := 0; i < seconds; i++ {
						pipe <- silence
					}
				} else {
					time, err := strconv.ParseFloat(filename, 64)
					if err == nil {
						tempSilence := make(audio.F64Samples, int(sampleRate*time))
						pipe <- tempSilence
					}
				}
			} else {
				in, err = os.Open(filename)
				if err != nil {
					panic(err)
				}
				decoder, _, err = audio.NewDecoder(in)
				if err != nil {
					panic(err)
				}

				for {
					read, err := decoder.Read(buf)
					if read > 0 {
						dst := make(audio.F64Samples, read)
						buf.Slice(0, read-1).CopyTo(dst)

						pipe <- dst
						//v.Volume = v.Volume * 0.5
					}
					if err == audio.EOS {
						break
					}
				}
				in.Close()
				for i := 0; i < gap; i++ {
					pipe <- silence
				}
			}
		}
		close(pipe)
	}()

	// consume the loop
	for outslice := range inChan {
		if outslice != nil {
			pcm := make(audio.PCM16Samples, outslice.Len())
			outslice.CopyTo(pcm)
			_, err := encoder.Write(pcm)
			if err != nil {
				panic(err)
			}
		}
	}
	encoder.Close()
	fmt.Printf("File %v written.\n", f.OutputFilename)
}
