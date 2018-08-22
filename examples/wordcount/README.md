# Example: Word Count

This example creates a simple hello-world style word-count experiment using the
standard `wc` command. The Docker image described here reads all files in the
`/input` directory and writes the resulting metrics to `/output`.

## Build

```bash
docker build -t wordcount .
```

## Configuration

This example accepts any number of sources under `/input`. For example, mount
two datasets by mapping them to `/input/1` and `/input/2`.

This example also supports the following environment variables:

- `COUNT_LINES` (optional): Set any value to count lines instead of words.

## Output

Results are placed in `/output`

## Example usage

```bash
beaker experiment run \
  --image wordcount \
  --source examples/moby:/input \
  --result-path /output
```
