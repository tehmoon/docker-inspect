# ARCHIVED

This is replaced by:

```
{ docker ps -q | \
  xargs -L8 -P8 docker inspect
} | jq '.'
```


# docker-inspect
Use the docker API to inspect each container to stdout.

## Caveats:

  - Docker-inspect is not atomic. It filters containers based on the filter, then loop through them to inspect.
  - Docker-inspect calls `docker`. It needs to know the version of your running docker. You can avoid this by setting `DOCKER_API_VERSION` if you know the accepted version.
  - When using template, it has to use the `JSON` function like this: `{{ . | json }}`.
  - Using multiple template doesn't call `docker` multiple times. It is safe to assume that the order of datas in template1 will be the same as template2

## Examples:

Inspect every container that have alpine:latest for parent and are running:

```
docker-inspect -filter ancestors=alpine:latest -filter status=running
```

You can specify multiple templates, each template would output a JSON object on one line:

```
docker-inspect -template '{{ .ID | JSON }}' -template '{{ .NetworkSettings | JSON }}'
```

## Usage:

```
Usage of ./docker-inspect:
  -filter value
        Filter to pass to docker. Can be repeated.
  -template value
        JSON template to pass to docker inspect
```
