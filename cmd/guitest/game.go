package main

import (
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"

	"github.com/wmarshpersonal/gogeebee/cartridge"
	"github.com/wmarshpersonal/gogeebee/gb"
	"github.com/wmarshpersonal/gogeebee/ppu"
)

type Game struct {
	gb   *gb.GB
	sync synchronization
}

func (g *Game) Update() error {
	// g.gb.ProcessJoypad(ReadButtons(), ReadDirections())

	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return width, height
}

func initGame(romData []byte) (*Game, error) {
	mbc, err := cartridge.Load(romData)
	if err != nil {
		return nil, err
	}

	g := &Game{
		gb: gb.NewDMG(mbc),
		sync: synchronization{
			cond:         sync.NewCond(&sync.Mutex{}),
			audioSamples: make([]byte, 0, sampleRate),
		},
	}

	ac := audio.NewContext(sampleRate)
	if p, err := ac.NewPlayer(&audioStream{sync: &g.sync}); err != nil {
		panic(err)
	} else {
		p.SetBufferSize(time.Millisecond * 20)
		p.Play()
		// defer p.Close()
	}

	go func() {
		var f float64 = 440
		var phase float64
		var samples []byte
		var frame ppu.PixelBuffer
		for {
			drawn := g.gb.RunFor(gb.TCyclesPerSecond/framesPerSecond, &frame)
			if drawn > 0 {
				g.sync.addFrame(frame)
			}
			samples = samples[:0]
			const samplesPerFrame = sampleRate / framesPerSecond
			for i := 0; i < samplesPerFrame; i++ {
				s := math.Sin(phase)
				phase += f * 2 * math.Pi / sampleRate
				v := int16(s * (math.MaxInt16 - 1))
				samples = append(samples, byte(v), byte(v>>8), byte(v), byte(v>>8))
			}
			g.sync.addSamples(samples)
			//
			// dropped calc is wrong. needs to be checked in synchronization
			dropped := max(0, drawn-1)
			if dropped > 0 {
				slog.Debug("frame drop", slog.Int("dropped", dropped))
			}
		}
	}()

	return g, nil
}

type audioStream struct {
	sync *synchronization
}

func (w *audioStream) Read(p []byte) (n int, err error) {
	n = w.sync.waitConsumeSamples(p)
	return
}
