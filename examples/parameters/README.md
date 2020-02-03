# Parameter Injection

This example demonstrates how to inject values into an experiment.

## Templates

Values are injected into an experiment specification via Go templates. When
parsing a templated spec file, anything between double curly braces ( `'{{'` and
`'}}'` ) will be evaluated as expressions and replaced.

Exported environment variables can be expanded with built-in `{{.Env.varName}}`.

### Example

The following command demonstrates expansion of an environment variable.

1. Upload the `busybox` Docker image:
   ```bash
   docker pull busybox
   beaker image create --name busybox busybox
   ```

1. Create an experiment which prints a substituted value.
   ```yaml
   description: Print {{.Env.USER}}
   tasks:
   - spec:
       image: busybox
       resultPath: /none
       args: ['sh', '-c', 'echo Parameter value: $ENV']
       env:
         ENV: {{.Env.USER}}
   ```

1. Run: `beaker experiment create -f spec.yaml`

### Additional Reading

See [Go templates](https://golang.org/pkg/text/template/) for an in-depth
description of what is possible with templates.

## Parameter Spaces

With the `alpha tune` command, Beaker can expand a parameter space from a simple
YAML specification containing an optional seed and a map of named parameters.

Parameters can be specified with the built-in `{{.Parameter.varName}}`. If the
parameter name contains special characters such as punctuation or spaces, it can
be written as `{{index .Parameter "my parameter"}}`.

A small set of built-in distributions is provided:
- `uniform-int`: an integer sampled uniformly from `[min, max)`
- `uniform`: a real number sampled uniformly from `[min, max)`
- `log-uniform`: a real number sampled from a logarithmic distribution in `[min, max)`
- `choice`: a single choice sampled uniformly from a provided list (`choices`).
   Choices need not be of the same type.

Fixed values are expressed as a scalar of any type.

See the following sample schema for details:
```yaml
# If omitted, the seed is set to the current unix timestamp.
seed: <int> 
parameters:
  FIXED_VALUE: <any scalar>
  UNIFORM_INT:
    distribution: uniform-int
    bounds: [<min>, <max>]
  UNIFORM_FLOAT:
    distribution: uniform
    bounds: [<min>, <max>]
  LOG_UNIFORM:
    distribution: log-uniform
    bounds: [<min>, <max>]
  CHOOSE_ONE:
    distribution: choice
    choices: [<anything>, ...]
```

### Setup

1. Download the sample space and template:
   - [parameter-space.yaml](./parameter-space.yaml)
   - [template.yaml](./template.yaml)

1. Upload the `busybox` Docker image:
   ```bash
   docker pull busybox
   beaker image create --name busybox busybox
   ```

### Example

The following example command automatically creates a group containing 5
experiments sampled from the provided space.

```bash
beaker alpha tune \
    --template template.yaml \
    --search parameter-space.yaml \
    --group parameter-search-example \
    --count 5
```
