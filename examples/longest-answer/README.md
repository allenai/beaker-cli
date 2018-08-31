# Example: Word Count

This example creates a simple experiment for Leaderboard.
The Docker image described here reads from
`/swag.csv` and writes the results to `/results/predictions.csv`.

## Build

```bash
docker build -t node-docker .
```

## Configuration

This example supports the following environment variables:

- `TESTING` (optional): Set to true to use the test data (test-data/swag-dev-head-200.csv).

## Output

Results are placed in `/results/predictions.csv`.

## Example usage

```bash
 docker run --rm node-docker
```
or for test
```bash
 docker run --rm -e TESTING=true node-docker
```
