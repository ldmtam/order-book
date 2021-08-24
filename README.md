```
device: Macbook Pro 15inch 2017
goos: darwin
goarch: amd64
pkg: github.com/ldmtam/order-book
BenchmarkProcessOrder1kLevels-8    	 4211292	       282 ns/op	      16 B/op	       2 allocs/op
BenchmarkProcessOrder5kLevels-8    	 2978466	       399 ns/op	      16 B/op	       2 allocs/op
BenchmarkProcessOrder10kLevels-8   	 2249125	       545 ns/op	      16 B/op	       2 allocs/op
BenchmarkProcessOrder20kLevels-8   	 1283671	       879 ns/op	      18 B/op	       2 allocs/op
BenchmarkPutAndGet-8               	123047564	        14.8 ns/op	       0 B/op	       0 allocs/op
--- BENCH: BenchmarkPutAndGet-8
    orderqueue_test.go:46: b.N: 1, Put: 1, Get: 0, Reset: 0
    orderqueue_test.go:46: b.N: 100, Put: 48, Get: 52, Reset: 0
    orderqueue_test.go:46: b.N: 10000, Put: 5047, Get: 4953, Reset: 0
    orderqueue_test.go:46: b.N: 1000000, Put: 500361, Get: 499639, Reset: 0
    orderqueue_test.go:46: b.N: 100000000, Put: 50002851, Get: 49997149, Reset: 1
    orderqueue_test.go:46: b.N: 123047564, Put: 61527850, Get: 61519714, Reset: 2
BenchmarkPutAndReset-8             	142053111	        13.1 ns/op	       0 B/op	       0 allocs/op
--- BENCH: BenchmarkPutAndReset-8
    orderqueue_test.go:84: b.N: 1, Put: 1, Reset: 0
    orderqueue_test.go:84: b.N: 100, Put: 52, Reset: 48
    orderqueue_test.go:84: b.N: 10000, Put: 4990, Reset: 5010
    orderqueue_test.go:84: b.N: 1000000, Put: 499722, Reset: 500278
    orderqueue_test.go:84: b.N: 97123572, Put: 48565933, Reset: 48557639
    orderqueue_test.go:84: b.N: 142053111, Put: 71024221, Reset: 71028890
PASS
ok  	github.com/ldmtam/order-book	124.274s
```
