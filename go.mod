module github.com/allenai/beaker

go 1.13

require (
	github.com/allenai/bytefmt v0.1.2
	github.com/beaker/client v0.0.0-20210712202434-7bf49eded4ee
	github.com/beaker/fileheap v0.0.0-20210701203425-01e1890a1025
	github.com/beaker/runtime v0.0.0-20210701181439-059d402e0a86
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.7+incompatible
	github.com/fatih/color v1.12.0
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.0
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/spf13/viper => ./viperstub
