module github.com/allenai/beaker

go 1.13

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.12 // indirect
	github.com/Sirupsen/logrus v1.0.6 // indirect
	github.com/allenai/bytefmt v0.1.0
	github.com/beaker/client v0.0.0-20210219201026-855103907db8
	github.com/beaker/fileheap v0.0.0-20210213001550-3d3932012952
	github.com/docker/distribution v2.7.0+incompatible
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.3.3 // indirect
	github.com/fatih/color v1.9.0
	github.com/hashicorp/go-hclog v0.10.1 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.0 // indirect
	github.com/spf13/cobra v1.1.1
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/spf13/viper => ./viperstub
