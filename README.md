# Beaker

[Beaker](https://beaker-pub.allenai.org) is a collaborative platform for
rapid and reproducible research.

## Prerequisites

[Docker](https://www.docker.com/) is the foundation for Beaker experiments and
must be installed to take full advantage of Beaker.

## Getting Started

1. Create an account at [beaker-pub.allenai.org](https://beaker-pub.allenai.org)
   and follow the instructions in your [account settings](https://beaker-pub.allenai.org/user).

   These instructions will guide you through installing and configuring the
   Beaker CLI. See [below](#install-beaker-cli) for more options.
   
   Request "Scientist" or higher credentials from a Beaker admin to get authorization
   to create experiments.

2. Run your first experiment. The following example
   [counts words](https://beaker-pub.allenai.org/bp/bp_qbjvcda1sed7) in the text
   of [Moby Dick](https://beaker-pub.allenai.org/ds/ds_1hz9k6sgxi0a).

   ```bash
   beaker experiment run \
     --name wordcount-moby \
     --blueprint examples/wordcount \
     --source examples/moby:/input \
     --result-path /output
   ```

## Install Beaker CLI

The most direct way to install Beaker is to download a
[release](https://github.com/allenai/beaker/releases) and extract it to your path.

```bash
tar -xvzf beaker_*.tar.gz -C /usr/local/bin
```

OS X users can install Beaker through [Homebrew](https://brew.sh/) with a custom tap.


```bash
brew tap allenai/homebrew-beaker https://github.com/allenai/homebrew-beaker.git
brew install beaker
```

Beaker can also be installed from source using standard [Go](https://golang.org/) tools.

```bash
go get -u github.com/allenai/beaker/...
```
## Notices
[Beaker dependencies and licenses](https://app.fossa.io/attribution/a462337b-67c8-418e-8a05-9b6f67de4626)
