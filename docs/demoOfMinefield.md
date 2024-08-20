To ingest data, we first need to clone `bitbomdev/bomsilo`. This is the repo that contains the data we will be ingesting.


``` sh
git clone https://github.com/bitbomdev/bomsilo
```

After cloning the repository, follow these steps to ingest data using Minefield:

1. Ingest SBOM data:
   ```sh
   minefield ingest sbom ../bomsilo/sboms/go
   ```
   This command ingests the SBOM data from the specified directory.

2. Cache the ingested data:
   ```sh
   minefield cache
   ```
   This step caches the ingested data for faster access and processing.

3. Generate a custom leaderboard:
   ```sh
   minefield leaderboard custom "dependents library"
   ```
   This command creates a custom leaderboard based on the "dependents library" criteria.

4. We should get an output containing these two purl's, which are node names in the graph
```
pkg:generic/checkout        
pkg:generic/setup-go
```
5. Run the following command to query the graph
```sh
minefield query "dependents library pkg:generic/checkout"
```
6. We can visualize this query using the following command
```sh
minefield query "dependents library pkg:generic/checkout" --visualize
```
7. We can run a similar query to get the dependents of another package
```sh
minefield query "dependents library pkg:generic/setup-go" --visualize
```
8. We can merge these queries together to see the shared dependents of both packages
```sh
minefield query "dependents library pkg:generic/checkout and dependents library pkg:generic/setup-go" --visualize
```
9. We can also find the diff between the two queries
```sh
minefield query "dependents library pkg:generic/checkout xor dependents library pkg:generic/setup-go" --visualize
```

Both queries 8 and 9 are equivalent to a sbom shared elements query, and a sbom diff query, though they can do much more, since if we root our query from a node that was an element of a sbom, we can do partial SBOM diffs, where we only compare a sub-set of a sbom, to another sub-set of another sbom.

