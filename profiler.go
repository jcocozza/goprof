/*
goprof is a convenience wrapper around go's pprof library.
If you need more control when profiling, don't use this.
There are three main ways to use this package.

1. Targets to a specific bit of code

	goprof.Run("<name>", func() {
		<your arbitrary code>
	})

2. Target a chunk of code. Anywhere in your code call the following:

	if err := goprof.Start("<name>"); err != nil {
		// handle error
	}

	<your code here>

	if err := goprof.End(); err != nil {
		// handle error
	}

3. Profile all your code. At the beginning of your main method, just call:

	goprof.Start("<name>")
	defer goprof.End()
*/
package goprof

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"
)

func cpuName(name string) string {
	return fmt.Sprintf("%s.cpu.pprof", name)
}
func blockName(name string) string {
	return fmt.Sprintf("%s.block.prof", name)
}
func traceName(name string) string {
	return fmt.Sprintf("%s.trace.out", name)
}
func heapName(name string) string {
	return fmt.Sprintf("%s.heap.prof", name)
}

type profiler struct {
	start time.Time
	end   time.Time

	// these are the different reports that get written out
	cpu   *os.File
	block *os.File
	trace *os.File
	heap  *os.File
}

var ErrAlreadyStarted = errors.New("profiler already started")
var ErrNotStarted = errors.New("profiler has not been started")

// true if started
func (p *profiler) started() bool {
	return !p.start.IsZero() && !p.end.IsZero()
}

func (p *profiler) duration() time.Duration {
	return p.end.Sub(p.start)
}

var p profiler

func setupFiles(name string) error {
	cpu, err := os.Create(cpuName(name))
	if err != nil {
		return err
	}
	p.cpu = cpu

	block, err := os.Create(blockName(name))
	if err != nil {
		return err
	}
	p.block = block

	trace, err := os.Create(traceName(name))
	if err != nil {
		return err
	}
	p.trace = trace

	heap, err := os.Create(heapName(name))
	if err != nil {
		return err
	}
	p.heap = heap
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
	if err := p.heap.Close(); err != nil {
		return err
	}
	return nil
}

// name is optional;
// if name is an empty string, will populate with a time stamp
func Start(name string) error {
	if p.started() {
		return ErrAlreadyStarted
	}

	if name == "" {
		name = fmt.Sprintf("goprof-%d", time.Now().UnixNano())
	}

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
	if !p.started() {
		return ErrNotStarted
	}
	// run this first; we don't want tear down to affect total time
	p.end = time.Now()
	pprof.StopCPUProfile()
	trace.Stop()
	if err := pprof.Lookup("block").WriteTo(p.block, 0); err != nil {
		return err
	}
	if err := pprof.WriteHeapProfile(p.heap); err != nil {
		return err
	}

	if err := cleanupFiles(); err != nil {
		return err
	}
	return nil
}

// convenience wrapper to profile an arbitrary function
func Run(name string, f func()) error {
	if err := Start(name); err != nil {
		return err
	}
	f()
	return Stop()
}

// summary functions

func Summarize() {
	fmt.Println(p.duration())
}

// print the commands to call for pprof
func Commands(name string) {
	fmt.Printf("go tool pprof %s\n", cpuName(name))
	fmt.Printf("go tool pprof -http=:6060 %s\n", cpuName(name))
	fmt.Printf("go tool trace %s\n", traceName(name))
	fmt.Printf("go tool pprof %s\n", blockName(name))
	fmt.Printf("go tool pprof %s\n", heapName(name))
}
