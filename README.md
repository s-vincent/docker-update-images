# Docker Image Updater

This utility takes care of updating container image you specified in a specific
toml file. So whenever the base image (i.e. the FROM section of the Dockerfile)
changed), it will rebuilt every images that depends on it.

## Pre-requisites

You need following Go packages:

 * golang.org/x/net/context
 * github.com/docker/docker/client
 * github.com/jhoonb/archivex
 * github.com/BurntSushi/toml

You have way to get dependencies. Either via `go get` or `dep`.

### Go get

Run the following:
```
go get golang.org/x/net/context
go get github.com/docker/docker/client
go get github.com/jhoonb/archivex
go get github.com/BurntSushi/toml
```

### Dep

Install `dep` via your package manager or via https://github.com/golang/dep
explaination.

BE SURE you have this project in $GOPATH/src/.

Then run the following:

```
dep ensure
```

## Compilation

Run: `make`

## Usage

docker-image-updater --file /path/to/myconfiguration.toml

You can pass options:
 * --api: to specify Docker API to use (see https://docs.docker.com/develop/sdk/#api-version-matrix)
 * --rebuild: to rebuild everything

## Sample configuration file

We ship an example of configuration file as well as Dockerfile samples:

```
[images.hello]
tag = ["test-hello:v1"]
dockerfile = "./samples/hello/Dockerfile"

[images.hello2]
tag = ["test-hello:v2"]
dockerfile = "./samples/hello2/Dockerfile"

[images.hello3]
tag = ["test-hello:v3"]
dockerfile = "./samples/hello3/Dockerfile"

[images.hello4]
tag = ["test-hello:v4"]
dockerfile = "./samples/hello4/Dockerfile"
```

