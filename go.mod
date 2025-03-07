module auto-pprof

go 1.24.0

require (
	github.com/google/pprof v0.0.0-20240312041847-bd984b5ce465
	github.com/syndtr/goleveldb v1.0.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/ianlancetaylor/demangle v0.0.0-20240312041847-bd984b5ce465 // indirect
)

replace github.com/google/pprof => ./pprof
