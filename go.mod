module github.com/allenai/beaker

go 1.16

require (
	github.com/Microsoft/go-winio v0.4.19 // indirect
	github.com/allenai/bytefmt v0.1.2
	github.com/beaker/client v0.0.0-20211109210027-f47a58bed970
	github.com/beaker/fileheap v0.0.0-20211007204440-1bd3920c4320
	github.com/beaker/runtime v0.0.0-20211109210053-606594000f3e
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.7+incompatible
	github.com/fatih/color v1.12.0
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.0
	golang.org/x/net v0.0.0-20211108170745-6635138e15ea // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

// See https://github.com/advisories/GHSA-c2h3-6mxw-7mvq
exclude github.com/containerd/containerd v1.2.10

exclude github.com/containerd/containerd v1.3.0-beta.2.0.20190828155532-0293cbd26c69

exclude github.com/containerd/containerd v1.3.0

exclude github.com/containerd/containerd v1.3.1-0.20191213020239-082f7e3aed57

exclude github.com/containerd/containerd v1.3.2

exclude github.com/containerd/containerd v1.4.0-beta.2.0.20200729163537-40b22ef07410

exclude github.com/containerd/containerd v1.4.1

exclude github.com/containerd/containerd v1.4.3

exclude github.com/containerd/containerd v1.5.0-beta.1

exclude github.com/containerd/containerd v1.5.0-beta.3

exclude github.com/containerd/containerd v1.5.0-beta.4

exclude github.com/containerd/containerd v1.5.0-rc.0

exclude github.com/containerd/containerd v1.5.1

replace github.com/spf13/viper => ./viperstub
