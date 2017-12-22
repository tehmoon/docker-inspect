# docker-inspect
Use the docker API to inspect each container to stdout.

## Examples:

Inspect every container that have ubuntu:latest for parent.

```
docker-inspect -filters ancestors=ubuntu:latest
```

## Usage:

```
Usage of ./docker-inspect:
  -filters string
        Filters to apply separated by ,
```
