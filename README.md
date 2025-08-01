# goprof

goprof is a convenience wrapper around go's pprof library.
If you need more control when profiling, don't use this.

You should only call `Start()` or `Run()` once in your code.
Calling more then once will return an error.
Calling `Stop()` before `Start()` or `Run()` will also produce an error.

There are three main ways to use this package:

1. Targets to a specific bit of code

```go
goprof.Run("<name>", func() {
	// <your arbitrary code>
})
```

2. Target a chunk of code. Anywhere in your code call the following:

```go
if err := goprof.Start("<name>"); err != nil {
	// handle error
}

// <your code here>

if err := goprof.End(); err != nil {
	// handle error
}
```

3. Profile all your code. At the beginning of your main method, just call:

```go
goprof.Start("<name>")
defer goprof.End()
```
