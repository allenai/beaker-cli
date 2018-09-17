# Working with Blueprints

Blueprints are Beaker's unit of executable code. A blueprint combines a Docker image with metadata,
such as its author and description, and an optional richer narrative in markdown. Please refer to
[the wordcount example](https://beaker-pub.allenai.org/bp/bp_qbjvcda1sed7) for an overview.

Like datasets, blueprints are immutable. The following example shows how to create and use blueprints.

## Preparation: Build a Docker Image

This example refers to the [wordcount](../examples/wordcount) example, but you can use any Docker
image at your disposal. To build it, copy the example's files and build them with the following
command

```bash
docker build -t wordcount <path/to/wordcount/directory>
```

## Create and Upload

You can create and upload any image as a blueprint with a single command. Behind the scenes, Beaker pushes your
Docker image to a private repository. This guarantees that the image will remain available and
unchanged for future experiments.

```
▶ beaker blueprint create --name wordcount wordcount
Pushing wordcount as wordcount (bp_qbjvcda1sed7)...
The push refers to repository [gcr.io/ai2-beaker-core/public/bduufrl06q5ner2l0440]
172b7a93847f: Preparing
bca0cc28f8e3: Preparing
3a1dff9afffd: Preparing
3a1dff9afffd: Layer already exists
bca0cc28f8e3: Pushed
172b7a93847f: Pushed
latest: digest: sha256:4c70545c15cca8d30b3adfd004a708fcdec910f162fa825861fe138200f80e19 size: 940
Done.
```

Notice that each blueprint is assigned a unique ID in addition to the name we chose. Any object,
including blueprints, can be referred to by its name or ID. Like any object, a blueprint can be
renamed, but its ID is guaranteed to remain stable. The following two commands are equivalent:

```bash
beaker blueprint inspect examples/wordcount
beaker blueprint inspect bp_qbjvcda1sed7
```

## Inspect

A blueprint's metadata can be retrieved with `beaker blueprint inspect`.

```
▶ beaker blueprint inspect examples/wordcount
[
    {
        "id": "bp_qbjvcda1sed7",
        "user": {
            "id": "us_gpx6zozipf5o",
            "name": "examples",
            "display_name": "Beaker Examples"
        },
        "name": "wordcount",
        "created": "2018-08-22T22:47:10.915236Z",
        "committed": "2018-08-22T22:47:23.422262Z",
        "original_tag": "wordcount",
        "description": "A simple example to count words in text."
    }
]
```

## Download

You can pull a blueprint to your local machine at any time with `beaker blueprint pull`.

```
▶ beaker blueprint pull examples/wordcount
Pulling gcr.io/ai2-beaker-core/public/bduufrl06q5ner2l0440 ...
latest: Pulling from ai2-beaker-core/public/bduufrl06q5ner2l0440
Digest: sha256:4c70545c15cca8d30b3adfd004a708fcdec910f162fa825861fe138200f80e19
Status: Downloaded newer image for gcr.io/ai2-beaker-core/public/bduufrl06q5ner2l0440:latest
Done.
```

In order to avoid accidentally overwriting your local images, this command assigns Beaker's random
internally assigned tag  `gcr.io/ai2-beaker-core/public/bduufrl06q5ner2l0440` by default. To assign
a more human-friendly tag, set it with an additional argument:

```
▶ beaker blueprint pull examples/wordcount friendly-name
```

## Cleanup

To clean up, simply `docker image rm` any images created above.

Because blueprints are immutable, they can't be deleted. It will be possible to archive them in the
near future.
