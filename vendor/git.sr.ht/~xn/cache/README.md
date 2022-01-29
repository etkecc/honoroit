# cache [![Go Report Card](https://goreportcard.com/badge/git.sr.ht/~xn/cache)](https://goreportcard.com/report/git.sr.ht/~xn/cache)[![Go Reference](https://pkg.go.dev/badge/git.sr.ht/~xn/cache.svg)](https://pkg.go.dev/git.sr.ht/~xn/cache)

different cache algorithms implementation in golang

```bash
go get git.sr.ht/~xn/cache
```

```go
// example: LRU

lru := cache.NewLRU(1000)
lru.Set(1, 10)
```

If you need only specific cache algorithm, you can import it without additional dependencies:

```bash
go get git.sr.ht/~xn/cache/lfu
```

```go
lfuCache := lfu.New(1000)
lfuCache.Set(1, 10)
```

# status

### implemented

* [Least Frequently Used](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least-frequently_used_(LFU)) - `cache.NewLFU`
* [Least Recently Used](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least_recently_used_(LRU)) - `cache.NewLRU()`
* [Time aware Least Recently Used](https://en.wikipedia.org/wiki/Cache_replacement_policies#Time_aware_least_recently_used_(TLRU)) - `cache.NewTLRU`

### tests

```bash
git.sr.ht/~xn/cache/cache.go:26:	NewLRU		100.0%
git.sr.ht/~xn/cache/cache.go:31:	NewTLRU		100.0%
git.sr.ht/~xn/cache/cache.go:36:	NewLFU		100.0%
git.sr.ht/~xn/cache/lfu/lfu.go:21:	New		100.0%
git.sr.ht/~xn/cache/lfu/lfu.go:32:	removeLFU	100.0%
git.sr.ht/~xn/cache/lfu/lfu.go:45:	Set		100.0%
git.sr.ht/~xn/cache/lfu/lfu.go:55:	Has		100.0%
git.sr.ht/~xn/cache/lfu/lfu.go:61:	Get		100.0%
git.sr.ht/~xn/cache/lfu/lfu.go:75:	Remove		100.0%
git.sr.ht/~xn/cache/lfu/lfu.go:85:	Purge		100.0%
git.sr.ht/~xn/cache/lru/lru.go:21:	New		100.0%
git.sr.ht/~xn/cache/lru/lru.go:32:	removeLRU	100.0%
git.sr.ht/~xn/cache/lru/lru.go:45:	Set		100.0%
git.sr.ht/~xn/cache/lru/lru.go:55:	Has		100.0%
git.sr.ht/~xn/cache/lru/lru.go:61:	Get		100.0%
git.sr.ht/~xn/cache/lru/lru.go:75:	Remove		100.0%
git.sr.ht/~xn/cache/lru/lru.go:85:	Purge		100.0%
git.sr.ht/~xn/cache/tlru/tlru.go:22:	New		100.0%
git.sr.ht/~xn/cache/tlru/tlru.go:36:	cleanup		100.0%
git.sr.ht/~xn/cache/tlru/tlru.go:51:	removeLRU	100.0%
git.sr.ht/~xn/cache/tlru/tlru.go:64:	Set		100.0%
git.sr.ht/~xn/cache/tlru/tlru.go:74:	Has		100.0%
git.sr.ht/~xn/cache/tlru/tlru.go:80:	Get		100.0%
git.sr.ht/~xn/cache/tlru/tlru.go:94:	Remove		100.0%
git.sr.ht/~xn/cache/tlru/tlru.go:104:	Purge		100.0%
total:					(statements)	100.0%
```

### benchmarks

```bash
goos: linux
goarch: amd64
pkg: git.sr.ht/~xn/cache/lfu
cpu: Intel(R) Core(TM) i7-8565U CPU @ 1.80GHz
BenchmarkSet-8      	5137729	         293.8 ns/op	     47 B/op	      2 allocs/op
BenchmarkSetX2-8    	  36928	        120980 ns/op	     51 B/op	      3 allocs/op
BenchmarkGet-8      	8728536	         152.3 ns/op	      0 B/op	      0 allocs/op
BenchmarkHas-8         17306770          89.02 ns/op	      0 B/op	      0 allocs/op
BenchmarkRemove-8   	9493314	         233.2 ns/op	      0 B/op	      0 allocs/op
BenchmarkPurge-8    	  14232	        175344 ns/op	 909389 B/op	      3 allocs/op
```

```bash
goos: linux
goarch: amd64
pkg: git.sr.ht/~xn/cache/lru
cpu: Intel(R) Core(TM) i7-8565U CPU @ 1.80GHz
BenchmarkSet-8      	 4158920	       297.3 ns/op	      47 B/op	       2 allocs/op
BenchmarkSetX2-8    	   38054	      123251 ns/op	      50 B/op	       3 allocs/op
BenchmarkGet-8      	 6358285	       223.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkHas-8      	16789678	       85.98 ns/op	       0 B/op	       0 allocs/op
BenchmarkRemove-8   	 9418458	       228.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkPurge-8    	   14974	      234563 ns/op	  909389 B/op	       3 allocs/op
```

```bash
goos: linux
goarch: amd64
pkg: git.sr.ht/~xn/cache/tlru
cpu: Intel(R) Core(TM) i7-8565U CPU @ 1.80GHz
BenchmarkSet-8      	3922689	         316.2 ns/op	     47 B/op	      2 allocs/op
BenchmarkSetX2-8    	  38558	        124703 ns/op	     50 B/op	      3 allocs/op
BenchmarkGet-8      	5020369	         232.1 ns/op	      0 B/op	      0 allocs/op
BenchmarkHas-8         14243359          95.20 ns/op	      0 B/op	      0 allocs/op
BenchmarkRemove-8   	8639568	         143.2 ns/op	      0 B/op	      0 allocs/op
BenchmarkPurge-8    	  10000	        186508 ns/op	 458824 B/op	      3 allocs/op
```
