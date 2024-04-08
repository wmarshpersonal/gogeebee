package main

import (
	"sync"

	"github.com/wmarshpersonal/gogeebee/ppu"
)

// TODO: rewrite all of this
type synchronization struct {
	cond *sync.Cond

	audioSamples []byte
	frame        *ppu.PixelBuffer
}

func (s *synchronization) addSamples(samples []byte) {
	s.cond.L.Lock()
	max := len(samples)
	for len(s.audioSamples) > max {
		s.cond.Wait()
	}
	s.audioSamples = append(s.audioSamples, samples...)
	s.cond.Broadcast()
	s.cond.L.Unlock()
}

func (s *synchronization) addFrame(f ppu.PixelBuffer) {
	s.cond.L.Lock()
	s.frame = &f
	s.cond.Broadcast()
	s.cond.L.Unlock()
}

func (s *synchronization) tryConsumeFrame() (f ppu.PixelBuffer, ok bool) {
	s.cond.L.Lock()
	ok = s.frame != nil
	if ok {
		f = *s.frame
		s.frame = nil
	}
	s.cond.Broadcast()
	s.cond.L.Unlock()
	return
}

func (s *synchronization) waitConsumeSamples(p []byte) (n int) {
	s.cond.L.Lock()
	for len(s.audioSamples) == 0 {
		s.cond.Wait()
	}
	n = min(len(p), len(s.audioSamples))
	copy(p, s.audioSamples[:n])
	s.audioSamples = s.audioSamples[n:]
	s.cond.Broadcast()
	s.cond.L.Unlock()

	return
}
