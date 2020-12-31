# High Performance Go

Notes taken while running through the workshop slides.

1. [Benchmarking](01-Benchmarking.md)
1. [Performance measurement and profiling](02-Performance-measurement-and-profiling.md)
1. [Compiler optimisations](03-Compiler-optimisations.md)
1. [Execution Tracer](04-Execution-Tracer.md)
1. [Memory and Garbage Collector](05-Memory-and-Garbage-Collector.md)
1. [Tips and trips](06-Tips-and-trips.md)

----

## TOC - High Performance Go Workshop

```
Dave Cheney dave@cheney.net version 379996b, 2019-07-24 

1. Benchmarking

    1.1. Benchmarking ground rules
    1.2. Using the testing package for benchmarking
    1.3. Comparing benchmarks with benchstat
    1.4. Avoiding benchmarking start up costs
    1.5. Benchmarking allocations
    1.6. Watch out for compiler optimisations
    1.7. Benchmark mistakes
    1.8. Benchmarking math/rand
    1.9. Profiling benchmarks
    1.10 Discussion

2. Performance measurement and profiling

    2.1. pprof
    2.2. Types of profiles
    2.3. One profile at at time
    2.4. Collecting a profile
    2.5. Analysing a profile with pprof

3. Compiler optimisations

    3.1. History of the Go compiler
    3.2. Escape analysis
    3.3. Inlining
    3.4. Dead code elimination
    3.5. Prove pass
    3.6. Compiler intrinsics
    3.7. Bounds check elimination
    3.8. Compiler flags exercises

4. Execution Tracer

    4.1. What is the execution tracer, why do we need it?
    4.2. Generating the profile
    4.3. Generating a profile with runtime/pprof
    4.4. Tracing vs Profiling
    4.5. Using more than one CPU
    4.6. Batching up work
    4.7. Using workers
    4.8. Using buffered channels
    4.9. Mandelbrot microservice

5. Memory and Garbage Collector

    5.1. Garbage collector world view
    5.2. Garbage collector design
    5.3. Minimise allocations
    5.4. Using sync.Pool
    5.5. Rearrange fields for better packing
    5.6. Exercises

6. Tips and trips

    6.1. Goroutines
    6.2. Go uses efficient network polling for some requests
    6.3. Watch out for IO multipliers in your application
    6.4. Use streaming IO interfaces
    6.5. Timeouts, timeouts, timeouts
    6.6. Defer is expensive, or is it?
    6.7. Make the fast path inlinable
    6.8. Range
    6.9. Avoid Finalisers
    6.10. Minimise cgo
    6.11. Always use the latest released version of Go
    6.12. Performance Mantras
```

### Versions

- `379996b, 2019-07-24` [Sections 1-6](https://dave.cheney.net/high-performance-go-workshop/gophercon-2019.html) (GopherCon San Diego)
- `g660848, 2019-04-26` [Sections 1-7](https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html) (dotGo Paris)

----

## Credit

You can view current version this presentation [here](https://bit.ly/dotgo2019)

----

## License and Materials

This presentation is licensed under the [Creative Commons Attribution-ShareAlike 4.0 International](https://creativecommons.org/licenses/by-sa/4.0/) licence.

You are encouraged to remix, transform, or build upon the material, providing you give appropriate credit and distribute your contributions under the same license.
