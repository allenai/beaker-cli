# Beaker Groups

### What is a group?
Groups are the way to organize your experiments and compare results across experiments. In addition to being a means of organizing similar experiments, groups allow you to compare the inputs and outputs of the experiments within the group, using metric and environment variable data from each task. Metric data is extracted from the `metrics.json` file in each experiment's result dataset, and environment variable data comes from the experiment's spec.

### How do I create a group?
Groups are created from the Experiments page by clicking the “Create Group” button. To create a group out of a set of experiments, simply click on the checkbox for each experiment to select it, and then click “Create Group” to both create the new group and automatically assign the selected experiments to it. Any experiment from yourself or another user can be added to a group, and Experiments may also belong to more than one group.

### How do I use a group?
Each group's detail page shows a table of all experiments, and their tasks, included in the group. This table is divided vertically, with information identifying each task and experiment on the left, and a user-configurable set of columns on the right. To select columns for display, click the Manage Comparisons button to bring up the Comparisons modal. Beaker will examine the members of the group and present you a list of all environment variables and metrics it found, and a count of how many tasks have data for that metric or variable.

### Metrics.Json
The metrics values for any one task come from the `metrics.json` file written to the experiment's output directory. In order for values to be correctly extracted, the file must contain data in JSON format, specifically a JSON object mapping metric names to values. The keys must be strings, but the values can be either numbers, strings, or arrays of numbers and/or strings. For example:

    ```json
    {
        "f1_score": 0.82,
        "precision": 0.77,
        "result_matrix": [[22, 28], [49, 64]],
        "top_concept": "gravity"
    }
    ```

Additional considerations:

- Nested values are not extracted into their own columns.
- Arrays of objects are not supported. For example:

    ```json
    [
        {
            "foo": "bar"
        },
        {
            "baz": 13
        }
    ]
    ```
