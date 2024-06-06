# Bitbom

Bitbom is a tool for graphing Software Bill of Materials (SBOMs) using [Roaring BitMaps](https://github.com/RoaringBitmap/roaring). It processes SBOMs and utilizes [`protobom`](https://github.com/protobom/protobom) to read them, enabling the creation of a graph that represents nodes and their connections.

## How the Graph is Constructed with a Bitmap

Each node in the graph has two bitmaps, a child and a parent bitmap. A dependency graph is created by adding the dependency node into the dependent node's child bitmap, and adding the dependent node into the dependency nodes parent bitmap, effectively representing the relationships between nodes.

## Dependency Graph Representation

The mermaid graph represents a simplified dependency graph. Each node in the graph has two bitmasks: a child bitmask and a parent bitmask. These bitmasks are used to represent the dependencies between nodes.

- **Child Bitmask**: Indicates the nodes that the current node depends on.
- **Parent Bitmask**: Indicates the nodes that depend on the current node.

In the graph:
- `Node A` has a child connection to `Node B`, meaning `Node A` depends on `Node B`.
    - And `Node B` has a parent connection to `Node A`, meaning that `Node B` is a dependency of `Node A`.
- `Node B` has a child connection to `Node C`, meaning `Node B` depends on `Node C`.
    - And `Node C` has a parent connection to `Node B`, meaning that `Node C` is a dependency of `Node B`.
- `Node C` has a child connection to `Node D`, meaning `Node C` depends on `Node D`.
    - And `Node D` has a parent connection to `Node C`, meaning that `Node D` is a dependency of `Node C`.

The arrows indicate the direction of the dependency, with the child bitmask of one node pointing to the next node and the parent bitmask of the next node pointing back to the previous node.


``` mermaid
graph TB

subgraph D_Node[Node D]
    subgraph D_Child_Node[Child Bitmask]
    end
    subgraph D_Parent_Node[Parent Bitmask]
    end
end

subgraph C_Node[Node C]
    subgraph C_Child_Node[Child Bitmask]
    end
    subgraph C_Parent_Node[Parent Bitmask]
    end
end

subgraph B_Node[Node B]
    subgraph B_Child_Node[Child Bitmask]
    end
    subgraph B_Parent_Node[Parent Bitmask]
    end
end

subgraph A_Node[Node A]
    subgraph A_Child_Node[Child Bitmask]
    end
    subgraph A_Parent_Node[Parent Bitmask]
    end
end

A_Child_Node --> B_Node
B_Parent_Node --> A_Node
B_Child_Node --> C_Node
C_Parent_Node --> B_Node
C_Child_Node --> D_Node
D_Parent_Node --> C_Node
```

## Acknowledgements

- https://github.com/RoaringBitmap/roaring
- https://github.com/protobom/protobom