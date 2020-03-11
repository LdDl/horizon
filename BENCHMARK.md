Our PC is:
```
    Processor: Intel(R) Core(TM) i9-7900X CPU @ 3.30GHz x 10
    Memory: 46.8GiB
    Linux Kernel: 4.15.0-20-generic
    OS: Linux Mint 19.1 Cinnamon
```

We've made three benchmarks: for [Euclidean space](map_matcher_euclidean_test.go#L99), for [Sphere (in terms of WGS84) with tiny road graph](map_matcher_4326_small_test.go#L71), for [Sphere (in terms of WGS84) with average road graph](map_matcher_4326_average_test.go#L90) (center of Moscow).

Benchmark for Euclidean space:
```bash
go test -benchmem -run=^$ mod -bench BenchmarkMapMatcherSRID_0

goos: linux
goarch: amd64
pkg: mod
BenchmarkMapMatcherSRID_0/Map_match_for_Euclidean_points/1/pts-4-20                 6050           2661081 ns/op            9331 B/op         39 allocs/op
BenchmarkMapMatcherSRID_0/Map_match_for_Euclidean_points/2/pts-4-20                  225           5576890 ns/op            9996 B/op         39 allocs/op
BenchmarkMapMatcherSRID_0/Map_match_for_Euclidean_points/4/pts-4-20                  198           6429569 ns/op           18853 B/op         39 allocs/op
BenchmarkMapMatcherSRID_0/Map_match_for_Euclidean_points/8/pts-4-20                  199           5820970 ns/op            6585 B/op         39 allocs/op
BenchmarkMapMatcherSRID_0/Map_match_for_Euclidean_points/16/pts-4-20                 189           6215218 ns/op            6620 B/op         39 allocs/op
BenchmarkMapMatcherSRID_0/Map_match_for_Euclidean_points/32/pts-4-20                 180           6400062 ns/op            6593 B/op         39 allocs/op
BenchmarkMapMatcherSRID_0/Map_match_for_Euclidean_points/64/pts-4-20                 164           6508480 ns/op           12274 B/op         39 allocs/op
BenchmarkMapMatcherSRID_0/Map_match_for_Euclidean_points/128/pts-4-20                165           6578800 ns/op            6613 B/op         39 allocs/op
BenchmarkMapMatcherSRID_0/Map_match_for_Euclidean_points/256/pts-4-20                172           6866143 ns/op            9431 B/op         39 allocs/op
BenchmarkMapMatcherSRID_0/Map_match_for_Euclidean_points/512/pts-4-20                156           7287148 ns/op           26037 B/op         39 allocs/op
BenchmarkMapMatcherSRID_0/Map_match_for_Euclidean_points/1024/pts-4-20               158           7232647 ns/op            6597 B/op         39 allocs/op
BenchmarkMapMatcherSRID_0/Map_match_for_Euclidean_points/2048/pts-4-20               157           7587050 ns/op            6584 B/op         39 allocs/op
BenchmarkMapMatcherSRID_0/Map_match_for_Euclidean_points/4096/pts-4-20               148           8462241 ns/op            6614 B/op         39 allocs/op
PASS
ok      mod     36.794s
```

Benchmark for Sphere (in terms of WGS84) with tiny road graph:
```bash
go test -benchmem -run=^$ mod -bench BenchmarkMapMatcherSRID_4326

goos: linux
goarch: amd64
pkg: mod
BenchmarkMapMatcherSRID_4326/Map_match_for_WGS84_points_(small_graph)/1/pts-4-20                    8098            152998 ns/op           25881 B/op        435 allocs/op
BenchmarkMapMatcherSRID_4326/Map_match_for_WGS84_points_(small_graph)/2/pts-4-20                    8384            136024 ns/op           25879 B/op        435 allocs/op
BenchmarkMapMatcherSRID_4326/Map_match_for_WGS84_points_(small_graph)/4/pts-4-20                    7592            148082 ns/op           25879 B/op        435 allocs/op
BenchmarkMapMatcherSRID_4326/Map_match_for_WGS84_points_(small_graph)/8/pts-4-20                   10000            137994 ns/op           25879 B/op        435 allocs/op
BenchmarkMapMatcherSRID_4326/Map_match_for_WGS84_points_(small_graph)/16/pts-4-20                   6916            145007 ns/op           25878 B/op        435 allocs/op
BenchmarkMapMatcherSRID_4326/Map_match_for_WGS84_points_(small_graph)/32/pts-4-20                   8173            142131 ns/op           25878 B/op        435 allocs/op
BenchmarkMapMatcherSRID_4326/Map_match_for_WGS84_points_(small_graph)/64/pts-4-20                   8368            148089 ns/op           25879 B/op        435 allocs/op
BenchmarkMapMatcherSRID_4326/Map_match_for_WGS84_points_(small_graph)/128/pts-4-20                  6818            149499 ns/op           25880 B/op        435 allocs/op
BenchmarkMapMatcherSRID_4326/Map_match_for_WGS84_points_(small_graph)/256/pts-4-20                  8055            145554 ns/op           25877 B/op        435 allocs/op
BenchmarkMapMatcherSRID_4326/Map_match_for_WGS84_points_(small_graph)/512/pts-4-20                  6855            148418 ns/op           25878 B/op        435 allocs/op
BenchmarkMapMatcherSRID_4326/Map_match_for_WGS84_points_(small_graph)/1024/pts-4-20                 8596            148680 ns/op           25880 B/op        435 allocs/op
BenchmarkMapMatcherSRID_4326/Map_match_for_WGS84_points_(small_graph)/2048/pts-4-20                 6896            145439 ns/op           25877 B/op        435 allocs/op
BenchmarkMapMatcherSRID_4326/Map_match_for_WGS84_points_(small_graph)/4096/pts-4-20                 8864            143908 ns/op           25878 B/op        435 allocs/op
PASS
ok      mod     21.888s
```

Benchmark for Sphere (in terms of WGS84) with average road graph (center of Moscow):
```bash
go test -benchmem -run=^$ mod -bench BenchmarkMapMatcherSRID_4326BIG

goos: linux
goarch: amd64
pkg: mod
BenchmarkMapMatcherSRID_4326BIG/Map_match_for_WGS84_points_(average_graph)/1/pts-10-20                42          28012087 ns/op        40584296 B/op      57558 allocs/op
BenchmarkMapMatcherSRID_4326BIG/Map_match_for_WGS84_points_(average_graph)/2/pts-10-20                52          27633486 ns/op        40587321 B/op      57609 allocs/op
BenchmarkMapMatcherSRID_4326BIG/Map_match_for_WGS84_points_(average_graph)/4/pts-10-20                37          28280790 ns/op        40587339 B/op      57622 allocs/op
BenchmarkMapMatcherSRID_4326BIG/Map_match_for_WGS84_points_(average_graph)/8/pts-10-20                46          28038703 ns/op        40587204 B/op      57629 allocs/op
BenchmarkMapMatcherSRID_4326BIG/Map_match_for_WGS84_points_(average_graph)/16/pts-10-20               38          28009554 ns/op        40585546 B/op      57601 allocs/op
BenchmarkMapMatcherSRID_4326BIG/Map_match_for_WGS84_points_(average_graph)/32/pts-10-20               44          27787651 ns/op        40587473 B/op      57623 allocs/op
BenchmarkMapMatcherSRID_4326BIG/Map_match_for_WGS84_points_(average_graph)/64/pts-10-20               49          25913703 ns/op        40582548 B/op      57513 allocs/op
BenchmarkMapMatcherSRID_4326BIG/Map_match_for_WGS84_points_(average_graph)/128/pts-10-20              57          27200184 ns/op        40586074 B/op      57606 allocs/op
BenchmarkMapMatcherSRID_4326BIG/Map_match_for_WGS84_points_(average_graph)/256/pts-10-20              37          28064470 ns/op        40587546 B/op      57632 allocs/op
BenchmarkMapMatcherSRID_4326BIG/Map_match_for_WGS84_points_(average_graph)/512/pts-10-20              55          26152540 ns/op        40589220 B/op      57646 allocs/op
BenchmarkMapMatcherSRID_4326BIG/Map_match_for_WGS84_points_(average_graph)/1024/pts-10-20             43          26873972 ns/op        40588389 B/op      57649 allocs/op
BenchmarkMapMatcherSRID_4326BIG/Map_match_for_WGS84_points_(average_graph)/2048/pts-10-20             43          26937283 ns/op        40587130 B/op      57609 allocs/op
BenchmarkMapMatcherSRID_4326BIG/Map_match_for_WGS84_points_(average_graph)/4096/pts-10-20             45          27221940 ns/op        40588577 B/op      57663 allocs/op
PASS
ok      mod     18.683s
```