# cache [![Go Report Card](https://goreportcard.com/badge/git.sr.ht/~xn/cache)](https://goreportcard.com/report/git.sr.ht/~xn/cache)[![Go Reference](https://pkg.go.dev/badge/git.sr.ht/~xn/cache.svg)](https://pkg.go.dev/git.sr.ht/~xn/cache)[![Go Documentation](https://godocs.io/git.sr.ht/~xn/cache?status.svg)](https://godocs.io/git.sr.ht/~xn/cache)

Thread-safe implementation of different cache algorithms in golang

[issues](https://todo.sr.ht/~xn/cache) | [lists](https://sr.ht/~xn/cache/lists)

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

* `Null` - empty/stub cache client, when you want to disable cache, but your code require something that implements `cache.Cache` interface - `cache.NewNull()`
* [Least Frequently Used](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least-frequently_used_(LFU)) - `cache.NewLFU()`
* [Least Recently Used](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least_recently_used_(LRU)) - `cache.NewLRU()`
* [Time aware Least Recently Used](https://en.wikipedia.org/wiki/Cache_replacement_policies#Time_aware_least_recently_used_(TLRU)) - `cache.NewTLRU()`
* [Memcached (distributed memory caching)](https://en.wikipedia.org/wiki/Memcached) - wrapper of [github.com/bradfitz/gomemcache](https://github.com/bradfitz/gomemcache) with `cache.Cache` interface - `cache.NewMemcached()`

### tests

```bash
total:						(statements)	96.2%
```

### benchmarks

```bash
goos: linux
goarch: amd64
pkg: git.sr.ht/~xn/cache/lfu
cpu: Intel(R) Core(TM) i7-8565U CPU @ 1.80GHz
BenchmarkSet-8      	4637701	      222.0 ns/op	     47 B/op	      2 allocs/op
BenchmarkSetX2-8    	  36528	     121528 ns/op	     50 B/op	      3 allocs/op
BenchmarkGet-8      	7890040	      157.2 ns/op	      0 B/op	      0 allocs/op
BenchmarkHas-8      	8764112	      154.3 ns/op	      0 B/op	      0 allocs/op
BenchmarkRemove-8   	8780908	      146.5 ns/op	      0 B/op	      0 allocs/op
BenchmarkPurge-8    	  14826	     199172 ns/op	 909389 B/op	      3 allocs/op
```

```bash
goos: linux
goarch: amd64
pkg: git.sr.ht/~xn/cache/lru
cpu: Intel(R) Core(TM) i7-8565U CPU @ 1.80GHz
BenchmarkSet-8      	4089240	      300.2 ns/op	     47 B/op	      2 allocs/op
BenchmarkSetX2-8    	  37603	     120438 ns/op	     50 B/op	      3 allocs/op
BenchmarkGet-8      	5671987	      248.3 ns/op	      0 B/op	      0 allocs/op
BenchmarkHas-8      	8222047	      158.3 ns/op	      0 B/op	      0 allocs/op
BenchmarkRemove-8   	8205411	      142.4 ns/op	      0 B/op	      0 allocs/op
BenchmarkPurge-8    	  10000	     152320 ns/op	 458828 B/op	      3 allocs/op
```

```bash
goos: linux
goarch: amd64
pkg: git.sr.ht/~xn/cache/tlru
cpu: Intel(R) Core(TM) i7-8565U CPU @ 1.80GHz
BenchmarkSet-8      	4036375	      315.4 ns/op	     55 B/op	      2 allocs/op
BenchmarkSetX2-8    	  36310	     117559 ns/op	     57 B/op	      3 allocs/op
BenchmarkGet-8      	4849862	      259.5 ns/op	      0 B/op	      0 allocs/op
BenchmarkHas-8         10223767	      155.3 ns/op	      0 B/op	      0 allocs/op
BenchmarkRemove-8   	9375744	      141.9 ns/op	      0 B/op	      0 allocs/op
BenchmarkPurge-8    	  15564	     173270 ns/op	 909389 B/op	      3 allocs/op
```

### v2 todo

* generics
* `key` arg type change from `interface{}` to `string`
* [request more changes](https://todo.sr.ht/~xn/cache)
