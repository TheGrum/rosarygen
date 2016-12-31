package rosarygen

import (
	"log"
	"math"
	"math/cmplx"

	"azul3d.org/audio.v1"
)

type (
	Processor interface {
		Process(p audio.Slice)
	}

	FFTProcessor interface {
		Process(p []complex128)
	}

	GenProcessor struct {
		p func(p audio.Slice)
	}
)

func Apply(in <-chan audio.Slice, q func(p audio.Slice)) <-chan audio.Slice {
	out := make(chan audio.Slice)
	go func() {
		defer close(out)
		for n := range in {
			q(n)
			out <- n
		}
	}()
	return out
}

func Split(in <-chan audio.Slice) (<-chan audio.Slice, <-chan audio.Slice) {
	a := make(chan audio.Slice)
	b := make(chan audio.Slice)
	go func() {
		defer close(a)
		defer close(b)
		for n := range in {
			n2 := make(audio.F64Samples, n.Len())
			n.CopyTo(n2)
			a <- n
			b <- n2
		}
	}()
	return a, b
}

func Merge(add ...<-chan audio.Slice) <-chan audio.Slice {
	out := make(chan audio.Slice)
	go func() {
		defer close(out)

		var bufs []*audio.Buffer
		var s audio.Slice
		var numBufs int
		var numOpen int

		// Create as many buffers as we have slices
		for _, ch := range add {
			s = <-ch
			bufs = append(bufs, audio.NewBuffer(s))
			numBufs += 1
		}

		// At this point we have at least one complete set of data
		process := func() {
			if numBufs == 0 {
				return
			}
			var minLen int = bufs[0].Len()

			for _, b := range bufs {
				if minLen > b.Len() {
					minLen = b.Len()
				}
			}
			// Now minLen has the greatest number of bytes
			// we can safely pull from every buffer

			f := make(audio.F64Samples, minLen)
			s := make(audio.F64Samples, minLen)
			for _, b := range bufs {
				_, _ = b.Read(s)
				for i, k := range s {
					f.Set(i, f.At(i)+k)
				}
			}

			for i, _ := range f {
				f.Set(i, f.At(i)/audio.F64(numBufs))
			}

			out <- f
		}

		process()
		// now loop getting new bits from each and building up our buffers
		for {
			numOpen = 0
			for i, ch := range add {
				s, open := <-ch
				if open {
					numOpen += 1
					bufs[i].Write(s)
				}
			}

			// And then empty what we can
			process()

			// if numOpen is 0 then all channels are closed and we are done
			// theoretically, if some channels produce excess data,
			// there might be some left in the pipe unmerged
			// I'm ignoring that for now
			if numOpen == 0 {
				break
			}
		}
	}()
	return out
}

func Transistor(input <-chan audio.Slice, control <-chan audio.Slice, controller func(input audio.F64, control audio.F64) audio.F64) <-chan audio.Slice {
	out := make(chan audio.Slice)
	go func() {
		defer close(out)

		var inBuf *audio.Buffer
		var cBuf *audio.Buffer
		var s audio.Slice
		var numOpen int

		s = <-input
		inBuf = audio.NewBuffer(s)
		s = <-control
		cBuf = audio.NewBuffer(s)

		// At this point we have at least one complete set of data
		process := func() {
			var minLen int = inBuf.Len()
			if minLen > cBuf.Len() {
				minLen = cBuf.Len()
			}

			// Now minLen has the greatest number of bytes
			// we can safely pull from every buffer

			f := make(audio.F64Samples, minLen)
			s := make(audio.F64Samples, minLen)
			c := make(audio.F64Samples, minLen)
			_, _ = inBuf.Read(s)
			_, _ = cBuf.Read(c)

			for i, _ := range f {
				f.Set(i, controller(s.At(i), c.At(i)))
			}

			out <- f
		}

		process()
		// now loop getting new bits from each and building up our buffers
		for {
			numOpen = 0
			s, open := <-input
			if open {
				numOpen += 1
				inBuf.Write(s)
			}
			s, open = <-control
			if open {
				numOpen += 1
				cBuf.Write(s)
			}

			// And then empty what we can
			process()

			// if numOpen is 0 then all channels are closed and we are done
			// theoretically, if some channels produce excess data,
			// there might be some left in the pipe unmerged
			// I'm ignoring that for now
			if numOpen == 0 {
				break
			}
		}
	}()
	return out
}

func (g *GenProcessor) Process(p audio.Slice) {
	if g.p != nil {
		g.p(p)
	}
}

/*
func (g *GenProcessor) Apply(in <-chan audio.Slice) <-chan audio.Slice {
	if g.p != nil {
		return Apply(in, g.p)
	} else {
		return in
	}
}
*/

func NewGenProcessor(f func(p audio.Slice)) *GenProcessor {
	return &GenProcessor{f}
}

type VolumeKnob struct {
	Volume float64
}

func (v *VolumeKnob) Process(p audio.Slice) {
	for i := 0; i < p.Len(); i = i + 1 {
		p.Set(i, audio.F64(float64(p.At(i))*v.Volume))
	}
}

type PureTone struct {
	LeftFreq, RightFreq float64
	SampleRate          float64
	stepL, phaseL       float64
	stepR, phaseR       float64
}

func (a *PureTone) Process(p audio.Slice) {
	for i := 0; i < p.Len(); i = i + 2 {
		p.Set(i, audio.F64(math.Sin(2*math.Pi*a.phaseL)))
		_, a.phaseL = math.Modf(a.phaseL + a.stepL)
	}
	for i := 1; i < p.Len(); i = i + 2 {
		p.Set(i, audio.F64(math.Sin(2*math.Pi*a.phaseR)))
		_, a.phaseR = math.Modf(a.phaseR + a.stepR)
	}
}

func (a *PureTone) SetFreq(LeftFreq, RightFreq float64) {
	a.LeftFreq = LeftFreq
	a.RightFreq = RightFreq
	a.stepL = LeftFreq / a.SampleRate
	a.stepR = RightFreq / a.SampleRate
}

func NewPureTone(LeftFreq, RightFreq, SampleRate float64) *PureTone {
	r := &PureTone{LeftFreq, RightFreq, SampleRate, 0, 0, 0, 0}
	r.stepL = LeftFreq / SampleRate
	r.stepR = RightFreq / SampleRate
	return r
}

type HarmonicTone struct {
	Freq             float64
	SampleRate       float64
	NumHarmonics     int
	HarmonicDistance float64
	VibratoDistance  float64
	VibratoRate      float64

	step, phase   []float64
	volScale      float64
	vibratoStep   float64
	vibratoExtent float64
	vibratoPos    float64
}

// 1  2          3                  4
// 1, 2/3 + 2/6, 4/7 + 4/14 + 4/28, 8/15 + 8/30 + 8/60 + 8/120
func (a *HarmonicTone) Process(p audio.Slice) {
	var F float64
	for i := 0; i < p.Len(); i = i + 2 {
		F = 0.0
		for j := 0; j < a.NumHarmonics; j++ {
			F += math.Sin(2*math.Pi*a.phase[j]) * (a.volScale / float64(j+1))
			_, a.phase[j] = math.Modf(a.phase[j] + a.step[j] + a.vibratoPos)
			a.vibratoPos += a.vibratoStep
			if ((a.vibratoStep > 0) && (a.vibratoPos > a.vibratoExtent)) || ((a.vibratoStep < 0) && (a.vibratoPos < a.vibratoExtent)) {
				a.vibratoStep = 0 - a.vibratoStep
				a.vibratoExtent = 0 - a.vibratoExtent
			}
		}
		p.Set(i, audio.F64(F)*0.9)
		p.Set(i+1, audio.F64(F)*0.9)
	}
}

func (a *HarmonicTone) SetFreq(Freq float64) {
	a.Freq = Freq
	for i := 0; i < a.NumHarmonics; i++ {
		a.step[i] = (Freq * (float64(i+1) * math.Pow(a.HarmonicDistance, float64(i)))) / a.SampleRate
	}
	log.Print(a.step)
}

func (a *HarmonicTone) SetVibrato(VibratoDistance float64, VibratoRate float64) {
	a.VibratoDistance = VibratoDistance
	a.VibratoRate = VibratoRate

	if a.VibratoDistance == 0 {
		a.vibratoPos = 0
		a.vibratoStep = 0
		a.vibratoExtent = 0
		return
	}
	a.vibratoPos = 0
	a.vibratoStep = (VibratoDistance / a.SampleRate) / (VibratoRate * a.SampleRate)
	a.vibratoExtent = (VibratoDistance / a.SampleRate) / 2
}

func NewHarmonicTone(Freq, SampleRate float64, num int, distance float64) *HarmonicTone {
	if num < 1 {
		panic("num cannot be less than 1")
	}
	r := &HarmonicTone{Freq, SampleRate, num, distance, 0, 0, nil, nil, 0, 0, 0, 0}
	r.step = make([]float64, num, num)
	r.phase = make([]float64, num, num)
	r.volScale = math.Pow(2, float64(num-1)) / (math.Pow(2, float64(num)) - 1)
	r.SetFreq(Freq)
	return r
}

type TonePattern struct {
	LeftSteps  []float64
	RightSteps []float64
	leftStep   int
	rightStep  int
	leftValue  float64
	rightValue float64
}

func (t *TonePattern) Process(p audio.Slice) {
	for i := 0; i < p.Len(); i = i + 2 {
		t.leftValue += t.LeftSteps[t.leftStep]
		t.leftStep += 1 //int(math.Mod(float64(t.leftStep), float64(len(t.LeftSteps))))
		if t.leftStep >= len(t.LeftSteps) {
			t.leftStep = 0
		}
		p.Set(i, audio.F64(t.leftValue))
	}
	for i := 1; i < p.Len(); i = i + 2 {
		t.rightValue += t.RightSteps[t.rightStep]
		//t.rightStep = int(math.Mod(float64(t.rightStep), float64(len(t.RightSteps))))
		t.rightStep += 1 //int(math.Mod(float64(t.leftStep), float64(len(t.LeftSteps))))
		if t.rightStep >= len(t.RightSteps) {
			t.rightStep = 0
		}

		p.Set(i, audio.F64(t.rightValue))
	}
}

func NewSymmetricTonePattern(Steps ...float64) *TonePattern {
	return &TonePattern{Steps, Steps, 0, 0, 0, 0}
}

// Give this a list of alternating steps and lengths
// [1, 5, 0, 5]
// and it will make a list of steps
// [1, 1, 1, 1, 1, 0, 0, 0, 0, 0]
func MakeTonePattern(steps ...float64) []float64 {
	out := make([]float64, 0, len(steps))
	for i := 0; i < len(steps); i += 2 {
		for j := 0; j < int(steps[i+1]); j++ {
			out = append(out, steps[i])
		}
	}
	return out
}

func StretchTonePattern(x int, steps ...float64) []float64 {
	out := make([]float64, len(steps)*x, len(steps)*x)
	for k, s := range steps {
		for i := 0; i < x; i++ {
			out[(k*x)+i] = s / float64(x)
		}
	}
	return out
}

// based on https://netwerkt.wordpress.com/2011/08/25/goertzel-filter/
type GoertzelFilter struct {
	Freq       float64
	SampleRate float64

	s_prev     float64
	s_prev2    float64
	totalpower float64
	n          int
	normFreq   float64
	coeff      float64
}

func (g *GoertzelFilter) Zero() {
	g.s_prev = 0
	g.s_prev2 = 0
	g.totalpower = 0
	g.n = 0
}

func (g *GoertzelFilter) Calculate(sample audio.F64) float64 {
	s := float64(sample) + g.coeff*g.s_prev - g.s_prev2
	g.s_prev2 = g.s_prev
	g.s_prev = s
	g.n++
	power := g.s_prev2*g.s_prev2 + g.s_prev*g.s_prev - g.coeff*g.s_prev*g.s_prev2
	g.totalpower += float64(sample) * float64(sample)
	if g.totalpower == 0 {
		g.totalpower = 1
	}
	return power / g.totalpower / float64(g.n)
}

func NewGoertzelFilter(Freq float64, SampleRate float64) *GoertzelFilter {
	g := &GoertzelFilter{Freq: Freq, SampleRate: SampleRate}
	g.normFreq = Freq / SampleRate
	g.coeff = 2 * math.Cos(2*math.Pi*g.normFreq)
	return g
}

type CyclingGoertzelFilter struct {
	Freq       float64
	SampleRate float64
	ResetEvery int

	filters []*GoertzelFilter
	active  int
	num     int
}

func (c *CyclingGoertzelFilter) Calculate(sample audio.F64) float64 {
	var out, temp float64

	c.num++
	for i := 0; i < len(c.filters); i++ {
		temp = c.filters[i].Calculate(sample)
		if i == c.active {
			out = temp
		}
	}
	if c.num >= c.ResetEvery {
		c.filters[c.active].Zero()
		c.num = 0
		c.active++
		if c.active >= len(c.filters) {
			c.active = 0
		}
	}
	return out
}

func NewCyclingGoertzelFilter(Freq float64, SampleRate float64, ResetEvery int, num int) *CyclingGoertzelFilter {
	c := &CyclingGoertzelFilter{Freq, SampleRate, ResetEvery, nil, 0, 0}
	c.filters = make([]*GoertzelFilter, num, num)
	for i := 0; i < num; i++ {
		c.filters[i] = NewGoertzelFilter(Freq, SampleRate)
	}
	return c
}

type GoertzelVolume struct {
	Freq       float64
	SampleRate float64

	Filter *CyclingGoertzelFilter
}

func (g *GoertzelVolume) Process(p audio.Slice) {
	if g.Filter == nil {
		return
	}
	for i := 0; i < p.Len(); i = i + 2 {
		result := g.Filter.Calculate((p.At(i) + p.At(i+1)) / 2)
		p.Set(i, p.At(i)*audio.F64(result))
		p.Set(i+1, p.At(i+1)*audio.F64(result))
	}
}

func NewGoertzelVolume(Freq float64, SampleRate float64, ResetEvery int, num int) *GoertzelVolume {
	return &GoertzelVolume{Freq, SampleRate, NewCyclingGoertzelFilter(Freq, SampleRate, ResetEvery, num)}
}

func FFTIndexFreq(i int, SampleRate float64, numBuckets int) float64 {
	return float64(i) * (SampleRate / float64(numBuckets) / 2)
}

func FFTFreqIndex(Freq, SampleRate float64, numBuckets int) int {
	return int(Freq / (SampleRate / float64(numBuckets) / 2))
}

func ComplexifyRealSlice(Freq []float64) []complex128 {
	F2 := make([]complex128, len(Freq))
	for i, k := range Freq {
		F2[i] = complex(k, 0)
	}
	return F2
}

func ComplexSliceModulus(Freq []complex128) []float64 {
	F2 := make([]float64, len(Freq))
	for i, k := range Freq {
		F2[i] = cmplx.Abs(k)
	}
	return F2
}

func CooleyTukeyDITFFTReal(Freq []float64, Result []complex128, n, s int) {
	F2 := ComplexifyRealSlice(Freq)
	CooleyTukeyDITFFT(F2, Result, n, s)
}

// Cooley-Tukey Decimation In Time Fast Fourier Transform radix-2
func CooleyTukeyDITFFT(Freq []complex128, Result []complex128, n, s int) {
	if n == 1 {
		Result[0] = Freq[0]
		return
	}
	n2 := n / 2
	partial := -2 * math.Pi / float64(n)
	CooleyTukeyDITFFT(Freq, Result, n2, 2*s)
	CooleyTukeyDITFFT(Freq[s:], Result[n2:], n2, 2*s)
	for i := 0; i < n2; i++ {
		twiddleFactor := cmplx.Rect(1, partial*float64(i)) * Result[i+n2]
		Result[i], Result[i+n2] = Result[i]+twiddleFactor, Result[i]-twiddleFactor
	}
}

// Cooley-Tukey Decimation In Time Inverse Fast Fourier Transform radix-2
func CooleyTukeyDITIFFT(Freq []complex128, Result []complex128, n, s int) {
	F2 := make([]complex128, n)
	for i, k := range Freq {
		//		fmt.Print(i, (n-1)-i, k)
		F2[(n-1)-i] = k
	}
	CooleyTukeyDITFFT(F2, Result, n, s)
	for i, k := range Result {
		Result[i] = k / complex(float64(n), 0)
	}
}

type Stack struct {
	Processors []Processor

	buff audio.Slice
}

func NewStack(p Processor) *Stack {
	s := &Stack{make([]Processor, 0, 1), nil}
	s.Processors = append(s.Processors, p)
	return s
}

func (s *Stack) Add(p Processor) {
	s.Processors = append(s.Processors, p)
}

func (s *Stack) Apply(in <-chan audio.Slice) <-chan audio.Slice {
	ch := in
	for _, k := range s.Processors {
		ch = Apply(ch, k.Process)
	}
	return ch
}

/*
func (s *Stack) Read(p audio.Slice) (n int, err error) {
	n, err := Reader.Read(buff)
	if n == 0 {
		return 0, err
	} else {
		if Process != nil {
			(*Process)(p, n)
		}
		return n, err
	}
}*/
