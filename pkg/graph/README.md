# Caching

In Bitbom, each node in the graph can have dependencies (children) and dependents (parents). To optimize the performance of querying these relationships, Bitbom uses a caching mechanism. This mechanism precomputes and stores the full set of dependencies and dependents for each node, reducing the need for redundant computations during queries.

## Table of Contents
- [Introduction](#introduction)
- [How Caching Works](#how-caching-works)
  - [Step-by-Step Explanation](#step-by-step-explanation)
  - [Example](#example)
- [Conclusion](#conclusion)

## How Caching Works

### Step-by-Step Explanation

1. **Identify Uncached Nodes**: The system first identifies nodes that need to be cached.
2. **Build Cache Maps**: It then builds cache maps for both dependencies (children) and dependents (parents) using a depth-first search (DFS) approach.
3. **Process Nodes**: For each node, it processes its dependencies or dependents recursively, merging cached relationships as needed.
4. **Save Cache**: Finally, the computed caches are saved, and the list of nodes to be cached is cleared.

### Example

Consider a graph with nodes `A`, `C`, `D`, and `E` where:
- `A` is independent
- `C` is independent
- `D` depends on `E`

These nodes are already cached.

``` mermaid
flowchart TD
A
C
D --> E
```

Now, let's add a new node B to the graph:
- A depends on B
- B depends on C
- D depends on B
- D depends on E

``` mermaid
flowchart TD
A --> B
B --> C
B --> D
D --> E
```

#### Caching Process for Node B

1. **Identify Uncached Node**: Node B needs caching.

2. **Build Cache Maps**:
    - For "children":
        - Cache for `B`: `{C, D, E}`
        - Update cache for `A`: `{B, C, D, E}`
        - Update cache for `D`: `{E}`
    - For "parents":
        - Cache for `B`: `{A}`
        - Update cache for `C`: `{A, B}`
        - Update cache for `D`: `{A, B}`
        - Update cache for `E`: `{A, B, D}`

3. **Save Cache**:
    - Save the computed cache for node `B` and update caches for nodes `A`, `C`, `D`, and `E`.

#### Final Cached Graph
- Node `A`: Children = `{B, C, D, E}`, Parents = `{}`
- Node `B`: Children = `{C, D, E}`, Parents = `{A}`
- Node `C`: Children = `{}`, Parents = `{A, B}`
- Node `D`: Children = `{E}`, Parents = `{A, B}`
- Node `E`: Children = `{}`, Parents = `{A, B, D}`

## Conclusion

The caching mechanism in Bitbom optimizes the performance of querying dependencies and dependents in a graph by precomputing and storing these relationships.
