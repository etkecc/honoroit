# cache [![Go Report Card](https://goreportcard.com/badge/git.sr.ht/~xn/cache)](https://goreportcard.com/report/git.sr.ht/~xn/cache)[![Go Reference](https://pkg.go.dev/badge/git.sr.ht/~xn/cache.svg)](https://pkg.go.dev/git.sr.ht/~xn/cache)[![Go Documentation](https://godocs.io/git.sr.ht/~xn/cache?status.svg)](https://godocs.io/git.sr.ht/~xn/cache)

Thread-safe implementation of different cache algorithms in golang.
Blazing fast `Get()`s, quite slow `Set()`s.

[issues](https://todo.sr.ht/~xn/cache) | [lists](https://sr.ht/~xn/cache/lists)

```bash
go get git.sr.ht/~xn/cache/v2
```

```go
// example: LRU

lru := cache.NewLRU[int](1000)
lru.Set("key", 10)
```

If you need only specific cache algorithm, you can import it without additional dependencies:

```bash
go get git.sr.ht/~xn/cache/v2/lfu
```

```go
lfuCache := lfu.New[int](1000)
lfuCache.Set("key", 10)
```

# status

### implemented

* `Null` - empty/stub cache client, when you want to disable cache, but your code require something that implements `cache.Cache` interface - `cache.NewNull()`
* [Least Frequently Used](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least-frequently_used_(LFU)) - `cache.NewLFU()`
* [Least Recently Used](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least_recently_used_(LRU)) - `cache.NewLRU()`
* [Time aware Least Recently Used](https://en.wikipedia.org/wiki/Cache_replacement_policies#Time_aware_least_recently_used_(TLRU)) - `cache.NewTLRU()`

### tests

```bash
total:						(statements)	100.0%
```

### benchmarks

```bash
goos: linux
goarch: amd64
pkg: git.sr.ht/~xn/cache/v2/lfu
cpu: Intel(R) Core(TM) i7-8565U CPU @ 1.80GHz
BenchmarkSet-8      	6875901	      205.2 ns/op	     24 B/op	      1 allocs/op
BenchmarkSetX2-8    	  38306	   122613 ns/op	     27 B/op	      1 allocs/op
BenchmarkGet-8      	8738271	      171.2 ns/op	      0 B/op	      0 allocs/op
BenchmarkHas-8      	14548699	      122.9 ns/op	      0 B/op	      0 allocs/op
BenchmarkRemove-8   	9629253	      165.6 ns/op	      0 B/op	      0 allocs/op
BenchmarkPurge-8    	  10000	   194397 ns/op	 458826 B/op	      3 allocs/op
```

```bash
goos: linux
goarch: amd64
pkg: git.sr.ht/~xn/cache/v2/lru
cpu: Intel(R) Core(TM) i7-8565U CPU @ 1.80GHz
BenchmarkSet-8      	5875112	      210.3 ns/op	     25 B/op	      1 allocs/op
BenchmarkSetX2-8    	  39282	   148461 ns/op	     26 B/op	      1 allocs/op
BenchmarkGet-8      	5668399	      229.2 ns/op	      0 B/op	      0 allocs/op
BenchmarkHas-8      	14897650	       95.25 ns/op	      0 B/op	      0 allocs/op
BenchmarkRemove-8   	9606914	      147.4 ns/op	      0 B/op	      0 allocs/op
BenchmarkPurge-8    	  10000	   174681 ns/op	 458827 B/op	      3 allocs/op
```

```bash
goos: linux
goarch: amd64
pkg: git.sr.ht/~xn/cache/v2/tlru
cpu: Intel(R) Core(TM) i7-8565U CPU @ 1.80GHz
BenchmarkSet-8      	5374263	      241.1 ns/op	     32 B/op	      1 allocs/op
BenchmarkSetX2-8    	  37250	   143659 ns/op	     34 B/op	      1 allocs/op
BenchmarkGet-8      	5046880	      267.7 ns/op	      0 B/op	      0 allocs/op
BenchmarkHas-8      	4878238	      273.1 ns/op	      0 B/op	      0 allocs/op
BenchmarkRemove-8   	9387120	      145.5 ns/op	      0 B/op	      0 allocs/op
BenchmarkPurge-8    	  10000	   184600 ns/op	 458826 B/op	      3 allocs/op
```
