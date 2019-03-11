
# Multistage Experiments

## Count words in parallel

![Parallel wordcount experiment graph](./images/parallel-wordcount.png)

A Beaker experiment may contain multiple tasks.
A task may depend on the results of another task in its experiment.
In this example, the `merge` task depends on the results of the `count-1`, `count-2`, and `count-3` tasks.
The count tasks are executed in parallel and the merge task runs automatically once they complete.

This example uses two images:
1. `example/wordcount`: Counts all words in the `/input` directory and produces a `word_count` metric.
2. `example/merge-wordcount`: Sums the `word_count` metrics in the `/input` directory.

To run an experiment with multiple tasks, you must define an experiment specification in YAML.
The full spec is available [here](./parallel-wordcount.yml).

```yaml
description: Parallel wordcount
tasks:
  - name: count-1
    spec:
      description: Count words in wordcount-1
      image: example/wordcount
      resultPath: /output
      datasetMounts:
        - datasetId: example/wordcount-1
          containerPath: /input/1.txt

...

  - name: merge
    spec:
      description: Merge wordcounts
      image: example/merge-wordcount
      resultPath: /output
    dependsOn:
     - parentName: count-1
       containerPath: /input/1.txt
     - ...
```

Run the experiment using `beaker experiment create`:

```
▶ beaker experiment create -f parallel-wordcount.yaml
Experiment ex_k6biwrvfcgbu submitted. See progress at http://beaker.allenai.org/ex/ex_k6biwrvfcgbu
```

