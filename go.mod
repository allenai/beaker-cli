module github.com/allenai/beaker

go 1.13

require (
	github.com/allenai/bytefmt v0.1.1
	github.com/beaker/client v0.0.0-20210421233118-5dcaf8303b8d
	github.com/beaker/fileheap v0.0.0-20210213001550-3d3932012952
	github.com/beaker/runtime v0.0.0-20210506004741-2e80fbeccd99
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.6+incompatible
	github.com/fatih/color v1.10.0
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	golang.org/x/net v0.0.0-20210505214959-0714010a04ed // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/spf13/viper => ./viperstub
