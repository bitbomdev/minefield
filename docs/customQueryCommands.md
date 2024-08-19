# Query Commands

Custom queries allow users to perform complex searches by combining multiple commands. These queries can be used to analyze the relationships between different nodes in the graph.

When running a custom query, there are multiple commands that can be used. For example, the custom command:

```sh
minefield query "dependencies library pkg:generic/dep1@1.0.0 and dependencies library pkg:generic/lib-A@1.0.0"
```


This custom command uses the commands "dependencies" and "and".

## Available Commands

### dependencies
- **Description**: Retrieves all dependencies of a specified node.
- **Usage**: `dependencies <node_type> <node_name>`
- **Example**: `dependencies library pkg:generic/dep1@1.0.0`

### dependents
- **Description**: Retrieves all dependents of a specified node.
- **Usage**: `dependents <node_type> <node_name>`
- **Example**: `dependents library pkg:generic/dep1@1.0.0`

### and
- **Description**: Combines two queries with a logical AND, returning nodes that satisfy both conditions.
- **Usage**: `<query1> and <query2>`
- **Example**: `dependencies library pkg:generic/dep1@1.0.0 and dependencies library pkg:generic/lib-A@1.0.0`

### or
- **Description**: Combines two queries with a logical OR, returning nodes that satisfy either condition.
- **Usage**: `<query1> or <query2>`
- **Example**: `dependencies library pkg:generic/dep1@1.0.0 or dependencies library pkg:generic/lib-A@1.0.0`

### xor
- **Description**: Combines two queries with a logical XOR, returning nodes that satisfy one condition but not both.
- **Usage**: `<query1> xor <query2>`
- **Example**: `dependencies library pkg:generic/dep1@1.0.0 xor dependencies library pkg:generic/lib-A@1.0.0`

## Examples

To find the dependencies shared by `dep1` and `lib-A`, you can use the following custom query:

```sh
minefield query "dependencies library pkg:generic/dep1@1.0.0 and dependencies library pkg:generic/lib-A@1.0.0"
```

This query will return all nodes that are dependencies of both `dep1` and `lib-A`.

### Leaderboard Custom Command

The `leaderboard custom` command can also utilize these queries to generate ranked lists based on the specified criteria.

```sh
minefield leaderboard custom "dependents library"
```


This command will generate a leaderboard of nodes based on the number of dependents they have.