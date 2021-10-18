module github.com/allenai/beaker

go 1.16

require (
	github.com/Microsoft/go-winio v0.4.19 // indirect
	github.com/allenai/bytefmt v0.1.2
	github.com/beaker/client v0.0.0-20211008183042-8f6405c659a9
	github.com/beaker/fileheap v0.0.0-20211007204440-1bd3920c4320
	github.com/beaker/runtime v0.0.0-20211014235104-30b5707384fb
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.7+incompatible
	github.com/fatih/color v1.12.0
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.0
	golang.org/x/net v0.0.0-20211015210444-4f30a5c0130f // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/spf13/viper => ./viperstub
