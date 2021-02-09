# Beaker

[Beaker](https://beaker.org) is a collaborative platform for rapid and reproducible research.

## Getting Started

1. Create an account at [beaker.org](https://beaker.org)
   and follow the instructions in your [account settings](https://beaker.org/user).  If you are
   an AI2 employee please use Google OAuth and your AI2 account.

   These instructions will guide you through installing and configuring the
   Beaker CLI. See [below](#install-beaker-cli) for more options.

   If you signed up using your @allenai.org email address, then you should
   automatically have credentials.  Otherwise you will need to
   request credentials to run an experiment from a Beaker admin.
   You can either ask on #beaker-users or email bunsen@allenai.org.
   Please include the email address associated with your Beaker account.

2. Run your first experiment. The following example
   [counts words](https://beaker.org/im/im_qbjvcda1sed7) in the text
   of [Moby Dick](https://beaker.org/ds/ds_1hz9k6sgxi0a).

   ```bash
   cat > count.yaml << EOF
   tasks:
   - spec:
       image: examples/wordcount
       resultPath: /output
       datasetMounts:
       - datasetId: examples/moby
         containerPath: /input
   EOF
   ```

   ```bash
   beaker experiment create -f count.yaml
   ```

3. For more in-depth documentation, see https://github.com/beaker/docs.

## Installation

### Prerequisites

[Docker](https://www.docker.com/) is the foundation for Beaker experiments and
must be installed to take full advantage of Beaker.

### Install Beaker CLI

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

## Executor

The Beaker executor runs Beaker jobs.
If you want to run Beaker jobs on your own machines, you can install the Beaker executor.

### Prerequisites

The Beaker executor is only supported on Linux.
It requires [Docker](https://docs.docker.com/engine/install/)
and the [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html)
if using GPUs.

### Installation

Each executor belongs to a cluster.
A cluster is a grouping of machines used for scheduling tasks.
If a task is scheduled to a cluster, it may run on any of the machines in that cluster.

You may choose an existing cluster from the existing [on-premise clusters](https://beaker.org/clusters) or create a new one:

```
beaker cluster create <name>
```

Next, install the executor:

```
beaker executor install <cluster>
```

When creating new experiments, specify the `cluster` property in each task
and they will run on your machine.

```yaml
tasks:
- context:
    cluster: <cluster>
```

## Notices
[Beaker dependencies and licenses](https://app.fossa.io/attribution/a462337b-67c8-418e-8a05-9b6f67de4626)
