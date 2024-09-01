# Minefield

[![Build and Test](https://github.com/bitbomdev/minefield/actions/workflows/build.yaml/badge.svg)](https://github.com/bitbomdev/minefield/actions/workflows/build.yaml)
[![OSV-Scanner Scheduled Scan](https://github.com/bitbomdev/minefield/actions/workflows/osv-schedule.yml/badge.svg)](https://github.com/bitbomdev/minefield/actions/workflows/osv-schedule.yml)

BitBom Minefield is a tool that uses roaring-**Bit**maps to graph S**BOM**s.

> The average user doesn't give a damn what happens, as long as (1) it works and (2) it's fast. - Daniel J. Bernstein

## Table of Contents

1. [Quickstart guide](#quickstart-guide)
2. [Example](#example)
3. [To start using Minefield](#to-start-using-minefield)
   - [Using Docker](#using-docker)
   - [Building from source](#building-from-source)
   - [Using go install](#using-go-install)
4. [How Minefield works](#how-minefield-works)
5. [Custom Query Commands](#custom-query-commands)
6. [Visualization of a query](#visualization-of-a-query)
7. [Star History](#star-history)
8. [Acknowledgements](#acknowledgements)


![demo.gif](demo.gif)
[View Minefield demo on asciinema](https://asciinema.org/a/674302)
## Quickstart guide

1. Ingest some data: 'minefield ingest sbom <sbom_file or sbom_dir>'  
2. Cache the data: 'minefield cache'
3. Run a query: 'minefield query <query_string>'

### Example

1. Ingest the `test` SBOM directory:
    ```sh
    minefield ingest sbom test
    ```
2. Cache the data:
    ```sh
    minefield cache
    ```
3. Run the leaderboard custom with "dependents PACKAGE":
    - This command generates a ranked list of packages, ordered by the number of other packages that depend on them
    ```sh
    minefield leaderboard custom "dependents PACKAGE"
    ```
4. Run a query on the top value from the leaderboard:
    - This command is now querying the dependents for a specific package, in this case dep2
    ```sh
    minefield query "dependents PACKAGE pkg:generic/dep2@1.0.0" 
    ```
5. Run queries to see the shared dependencies of lib-A and dep1, and lib-A and lib-B
    - These queries output the intersection of two queries, in this case we are finding package dependencies do each of the packages share between each other.
    ```sh
    minefield query "dependencies PACKAGE pkg:generic/dep1@1.0.0 and dependencies PACKAGE pkg:generic/lib-A@1.0.0" 
    ```
    ```sh
    minefield query "dependencies PACKAGE pkg:generic/lib-B@1.0.0 and dependencies PACKAGE pkg:generic/lib-A@1.0.0" 
    ```
6. Run queries with the visualizer
     ```sh
    minefield query "dependents PACKAGE pkg:generic/dep2@1.0.0 --visualize" 
    ```

## To start using Minefield

### Using Docker

```sh
docker pull ghcr.io/bit-bom/minefield:latest
docker run -it ghcr.io/bit-bom/minefield:latest
```

### Building from source

```sh
git clone https://github.com/bit-bom/minefield.git
cd minefield
go build -o minefield main.go
./minefield
```

### Using go install

```sh
go install github.com/bit-bom/minefield@latest
minefield
```
## How Minefield works

The design decisions and architecture of Minefield can be found [here](docs/bitbom.pdf).

## Custom Query Commands

For detailed information on available query commands and their usage, please refer to the [Custom Query Commands documentation](docs/customQueryCommands.md).

## Visualization of a query

![img.png](img.png)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=bitbomdev/minefield&type=Date)](https://star-history.com/#bitbomdev/minefield&Date)
## Acknowledgements

- https://github.com/RoaringBitmap/roaring
