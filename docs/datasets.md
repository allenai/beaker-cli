# Working with Datasets

The following example shows how to create and use datasets. Beaker takes the view that all datasets
are files at their core. Datasets may be one file or many, large or small; Beaker makes as few
restrictions as possible on the shape of your data. Datasets are also immutable, ensuring that
Beaker can retain a clean record of all data files used by an experiment.

## Preparation: Create Some Files

First create a working directory to contain our example.

```bash
mkdir -p ~/dataset-example
cd ~/dataset-example
```

Next we'll create some example files to upload as datasets.

```bash
mkdir -p multifile/nested
echo 'First file' > multifile/first
echo 'Second file' > multifile/second
echo 'Nested file' > multifile/nested/file
echo 'This is a single file.' > singlefile
```

Verify the file contents. The following command finds all files in the current
directory and prints each file's path and contents.

```
▶ find . -type f -exec echo {} \; -exec cat {} \; -exec echo \;
./multifile/first
First file

./multifile/second
Second file

./multifile/nested/file
Nested file

./singlefile
This is a single file.
```

## Upload

You can upload any file or collection of files as a dataset with a single command. Here we'll create
two datasets.

Our first dataset is a single file. Single-file datasets are treated by the system as a file.

```
▶ beaker dataset create --name my-file-dataset ./singlefile
Uploading my-file-dataset (ds_as6t74lspoc5)...
Done.
```

Our second dataset contains several files. This type of dataset is treated by the system as a
directory. _(To upload a single file as a directory, simply place the file in an empty directory and
upload the directory.)_

```
▶ beaker dataset create --name my-dir-dataset ./multifile
Uploading my-dir-dataset (ds_4kjgriri1sro)...
Done.
```

Notice that each dataset is assigned a unique ID in addition to the name we chose. Any object,
including datasets, can be referred to by its name or ID. Like any object, a dataset can be renamed,
but its ID is guaranteed to remain stable. The following two commands are equivalent:

```bash
beaker dataset inspect my-file-dataset
beaker dataset inspect ds_as6t74lspoc5
```

## Inspect

A dataset can be inspected with `beaker dataset inspect`, which produces a JSON representation of
the dataset. An optional `--manifest` flag, if provided, will also produce the dataset's contents.

```
▶ beaker dataset inspect --manifest my-file-dataset
[
    {
        "id": "ds_as6t74lspoc5",
        "user": {
            ...
        },
        "name": "my-file-dataset",
        "created": "2018-08-06T18:00:36.235254Z",
        "committed": "2018-08-06T18:00:36.646142Z",
        "manifest": {
            "id": "ds_as6t74lspoc5",
            "single_file": true,
            "files": [
                {
                    "file": "/singlefile",
                    "size": 23,
                    "time_last_modified": "2018-08-06T18:00:36.609Z"
                }
            ]
        }
    }
]
```

## Download

You can download a dataset to your local drive at any time with `beaker dataset fetch`. Beaker's
`fetch` command follows the same rules as the standard `cp` command. The following example downloads
the single-file dataset to an empty directory. Notice how the original filename is restored by default.

```
▶ mkdir fetched
▶ beaker dataset fetch -o fetched my-file-dataset
Downloading dataset ds_as6t74lspoc5 to file fetched/singlefile ... done.

▶ cat fetched/singlefile
This is a single file.
```

## Use Datasets in an Experiment

To demonstrate how to use a dataset in an experiment, we'll run the same find command we ran above
as a Beaker experiment. The code for this experiment's blueprint can be found
[here](../examples/list-files).

```bash
beaker experiment run \
    --blueprint example/list-files \
    --env LIST_DIR=/data \
    --source my-file-dataset:/data/single \
    --source my-dir-dataset:/data/multi \
    --result-path /results
```

This command will print a URL to your experiment, which should complete momentarily. Observe the
logs emitted from the experiment should be similar to the find command above.

## Cleanup

To clean up, simply remove the `~/dataset-example` directory we created at the beginning.

Because datasets are immutable, they can't be deleted. It will be possible to archive datasets in
the near future.
