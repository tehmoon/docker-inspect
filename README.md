# docker-inspect
Use the docker API to inspect each container to stdout.


## Caveats:

  - Docker-inspect is not atomic. It filters containers based on the filter, then loop through them to inspect.
  - Docker-inspect calls `docker`. It needs to know the version of your running docker. You can avoid this by setting `DOCKER_API_VERSION` if you know the accepted version.

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
