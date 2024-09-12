<p align="center">
  <img src="images/bitbomLogoAndName.png" alt="BitBom Long Logo" >
</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/bit-bom/minefield)](https://goreportcard.com/report/github.com/bit-bom/minefield)
[![Build and Test](https://github.com/bitbomdev/minefield/actions/workflows/build.yaml/badge.svg)](https://github.com/bitbomdev/minefield/actions/workflows/build.yaml)

> Before moving on, please consider giving us a GitHub star ⭐️. Thank you!

BitBom Minefield is a tool that uses roaring-**Bit**maps to graph S**BOM**s FAST.

![Terminal Output](images/terminal.png)
> Caching 10,000 SBOMs packages transitive dependents in 30 seconds.

## Table of Contents

1. [Quickstart Guide](#quickstart-guide)
2. [Example](#example)
3. [To Start Using Minefield](#to-start-using-minefield)
   - [Using Docker](#using-docker)
   - [Building From Source](#building-from-source)
4. [How Minefield Works](#how-minefield-works)
5. [Custom Query Commands](#custom-query-commands)
6. [Visualization of a Query](#visualization-of-a-query)
7. [Documentation](#documentation)
8. [Blog](#blog)
9. [Star History](#star-history)
10. [Acknowledgements](#acknowledgements)

[View Minefield demo on asciinema](https://asciinema.org/a/674302)

## Quickstart Guide

1. **Ingest some data:**
   ```sh
   minefield ingest sbom <sbom_file or sbom_dir>
   ```
2. **Cache the data:**
   ```sh
   minefield cache
   ```
3. **Run a query:**
   ```sh
   minefield query <query_string>
   ```

### Example

_Redis must be running at `localhost:6379`. If not, please use `make docker-up` to start Redis._

_Redis must be running at `localhost:6379`, if not please use `make docker-up` to start Redis._
1. Start the api server
   ```shell
   minefield start-service 
   ```

2. Ingest the `test` SBOM directory:
    ```sh
    minefield ingest sbom testdata
    ```
2. **Cache the data:**
    ```sh
    minefield cache
    ```
3. **Run the leaderboard custom with "dependents PACKAGE":**
   - This command generates a ranked list of packages, ordered by the number of other packages that depend on them.
    ```sh
    minefield leaderboard custom "dependents PACKAGE"
    ```
4. **Run a query on the top value from the leaderboard:**
   - This command queries the dependents for a specific package, in this case `dep2`.
5. **Run queries to see the shared dependencies of `lib-A` and `dep1`, and `lib-A` and `lib-B`:**
   - These queries output the intersection of two queries, finding package dependencies shared between each pair.
    ```sh
    minefield query "dependencies PACKAGE pkg:generic/lib-B@1.0.0 and dependencies PACKAGE pkg:generic/lib-A@1.0.0"
    ```
6. **Run queries with the visualizer:**
    ```sh
    minefield query "dependents PACKAGE pkg:generic/dep2@1.0.0 --visualize"
## To Start Using Minefield

### Using Docker

```sh
docker pull ghcr.io/bitbomdev/minefield:latest
docker run -it ghcr.io/bitbomdev/minefield:latest
```

### Building From Source

```sh
git clone https://github.com/bitbomdev/minefield.git
cd minefield
go build -o minefield main.go
./minefield
```

## How Minefield Works

The design decisions and architecture of Minefield can be found [here](docs/bitbom.pdf).

## Visualization of a Query

![img.png](images/img.png)

## Documentation

For comprehensive guides and detailed documentation, please visit our [Docs](https://bitbom.dev/docs/intro).

## Blog

Stay updated with the latest news and insights by visiting our [Blog](https://bitbom.dev/blog).

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=bitbomdev/minefield&type=Date)](https://star-history.com/#bitbomdev/minefield&Date)

## Acknowledgements

- https://github.com/RoaringBitmap/roaring
