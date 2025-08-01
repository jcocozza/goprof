package goprof

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"
)

type profiler struct {
	start time.Time
	end   time.Time

	// these are the different reports that get written out
	cpu   *os.File
	block *os.File
	trace *os.File
}

func (p *profiler) duration() time.Duration {
	return p.end.Sub(p.start)
}

var p profiler

func setupFiles(name string) error {
	cpu, err := os.Create(fmt.Sprintf("%s.cpu.pprof", name))
	if err != nil {
		return err
	}
	p.cpu = cpu

	block, err := os.Create(fmt.Sprintf("%s.block.prof", name))
	if err != nil {
		return err
	}
	p.block = block

	trace, err := os.Create(fmt.Sprintf("%s.trace.out", name))
	if err != nil {
		return err
	}
	p.trace = trace
	return nil
}

func cleanupFiles() error {
	if err := p.cpu.Close(); err != nil {
		return err
	}
	if err := p.block.Close(); err != nil {
		return err
	}
	if err := p.trace.Close(); err != nil {
		return err
	}
	return nil
}

func Start(name string) error {
	if err := setupFiles(name); err != nil {
		return err
	}

	if err := pprof.StartCPUProfile(p.cpu); err != nil {
		return err
	}

	if err := trace.Start(p.trace); err != nil {
		return err
	}

	runtime.SetBlockProfileRate(1)

	// run this last; we don't want setup to affect total time
	p.start = time.Now()
	return nil
}

func Stop() error {
	// run this first; we don't want teardown to affect total time
	p.end = time.Now()
	pprof.StopCPUProfile()
	trace.Stop()
	if err := pprof.Lookup("block").WriteTo(p.block, 0); err != nil {
		return err
	}

	if err := cleanupFiles(); err != nil {
		return err
	}
	return nil
}

// summary functions

func Summarize() {
	fmt.Println(p.duration())
}
