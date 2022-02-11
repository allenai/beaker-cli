module github.com/allenai/beaker

go 1.17

require (
	github.com/allenai/bytefmt v0.1.2
	github.com/beaker/client v0.0.0-20220118174312-992c493d18d8
	github.com/beaker/fileheap v0.0.0-20211007204440-1bd3920c4320
	github.com/beaker/runtime v0.0.0-20211213171103-95151aa06fad
	github.com/docker/distribution v2.8.0+incompatible
	github.com/docker/docker v20.10.12+incompatible
	github.com/fatih/color v1.13.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.5.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d // indirect
	github.com/beaker/unique v0.0.0-20210625205350-416101674f78 // indirect
	github.com/containerd/containerd v1.5.9 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/goware/urlx v0.3.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/vbauerster/mpb/v4 v4.12.2 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	golang.org/x/crypto v0.0.0-20211209193657-4570a0811e8b // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/sys v0.0.0-20220207234003-57398862261d // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220208230804-65c12eb4c068 // indirect
	google.golang.org/grpc v1.44.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
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

// See https://github.com/google/gnostic/issues/262
replace github.com/googleapis/gnostic v0.5.6 => github.com/googleapis/gnostic v0.5.5

replace github.com/googleapis/gnostic v0.5.7 => github.com/googleapis/gnostic v0.5.5
