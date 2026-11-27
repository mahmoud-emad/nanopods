# nanopods

nanopods is a tiny container runtime written in Go, built to explore how containers work: Linux namespaces, cgroups, networking, and pods.

## Usage

```bash
nanopods run <container-name> <image[:tag]>
```

For example:

```bash
nanopods run web alpine:3.20
```

If no tag is given, `latest` is used.

## Build

```bash
make build
make clean
```

## What works today

- Parsing `image[:tag]` references.
- Requesting a pull token from the Docker registry and fetching the image manifest.
- Downloading the image config and layers, verifying their digests, and storing them in `~/.nanopods/images/<name>/<tag>/`.
- Creating a container with a generated ID.

Namespaces, cgroups, root filesystem setup, and networking are not implemented yet.

## Status

nanopods is a learning project. Instead of only using Docker, I wanted to build something myself and understand how things work behind the scenes.

## Resources

1. [Open Container Initiative Distribution Specification](https://github.com/opencontainers/distribution-spec/blob/main/spec.md)
2. [Image Specification](https://github.com/opencontainers/image-spec)
3. [Open Containers](https://github.com/opencontainers)
4. [runc](https://github.com/opencontainers/runc)