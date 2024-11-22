package graph

import (
	"fmt"

	"github.com/RoaringBitmap/roaring"
	"github.com/bitbomdev/minefield/pkg/utils"
)

func Cache(storage Storage) error {
	uncachedNodes, err := storage.ToBeCached()
	if err != nil {
		return fmt.Errorf("error getting uncached nodes: %w", err)
	}
	if len(uncachedNodes) == 0 {
		return nil
	}
	keys, err := storage.GetAllKeys()
	if err != nil {
		return fmt.Errorf("error getting keys: %w", err)
	}

	// Retrieve all nodes at once
	allNodes, err := storage.GetNodes(keys)
	if err != nil {
		return fmt.Errorf("error getting all nodes: %w", err)
	}

	scc := findCycles(len(keys), allNodes)

	cachedChildren, err := buildCache(uncachedNodes, ChildrenDirection, scc, allNodes)
	if err != nil {
		return fmt.Errorf("error building cached children: %w", err)
	}

	cachedParents, err := buildCache(uncachedNodes, ParentsDirection, scc, allNodes)
	if err != nil {
		return fmt.Errorf("error building cached parents: %w", err)
	}

	cachedChildKeys, cachedChildValues, err := cachedChildren.GetAllKeysAndValues()
	if err != nil {
		return fmt.Errorf("error getting cached child keys and values: %w", err)
	}

	var caches []*NodeCache

	for i := 0; i < len(cachedChildKeys); i++ {
		childId := cachedChildKeys[i]
		childIntId, err := utils.StrToUint32(childId)
		if err != nil {
			return fmt.Errorf("error converting child key %s: %w", childId, err)
		}

		childBindValue := cachedChildValues[i].Clone()
		childBindValue.Add(childIntId)

		tempValue, err := cachedParents.Get(utils.Uint32ToStr(childIntId))
		if err != nil {
			return fmt.Errorf("error getting value for key %s: %w", childId, err)
		}
		parentBindValue := tempValue.Clone()
		parentBindValue.Add(childIntId)
		caches = append(caches, NewNodeCache(childIntId, parentBindValue, childBindValue))
	}

	if err := storage.SaveCaches(caches); err != nil {
		return fmt.Errorf("error saving caches: %w", err)
	}
	return storage.ClearCacheStack()
}

func findCycles(numOfNodes int, allNodes map[uint32]*Node) map[uint32]uint32 {
	var stack []uint32
	var tarjanDFS func(nodeID uint32)

	currentTarjanID := 0
	nodeToTarjanID := map[uint32]uint32{}
	lowLink := make(map[uint32]uint32)
	inStack := roaring.New()

	tarjanDFS = func(nodeID uint32) {
		currentNode := allNodes[nodeID]
		currentTarjanID++
		stack = append(stack, nodeID)
		inStack.Add(nodeID)
		nodeToTarjanID[nodeID] = uint32(currentTarjanID)
		lowLink[nodeID] = uint32(currentTarjanID)

		for _, nextNode := range currentNode.Children.ToArray() {
			if _, visited := nodeToTarjanID[nextNode]; !visited {
				tarjanDFS(nextNode)

				lowLink[nodeID] = min(lowLink[nodeID], lowLink[nextNode])
			} else if inStack.Contains(nextNode) {
				lowLink[nodeID] = min(lowLink[nodeID], nodeToTarjanID[nextNode])
			}
		}

		if nodeToTarjanID[nodeID] == lowLink[nodeID] {
			for len(stack) > 0 {
				id := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				inStack.Remove(id)
				lowLink[id] = nodeToTarjanID[nodeID]
				if nodeID == id {
					break
				}
			}
		}
	}

	for id := 1; id < numOfNodes+1; id++ {
		if _, visited := nodeToTarjanID[uint32(id)]; !visited {
			tarjanDFS(uint32(id))
		}
	}

	return lowLink
}

type stackElm struct {
	id        uint32
	todoIndex int
}

type todoFuturePair struct {
	todoNodes   []uint32
	futureNodes []uint32
}

func buildCache(uncachedNodes []uint32, direction Direction, scc map[uint32]uint32, allNodes map[uint32]*Node) (*NativeKeyManagement, error) {
	cache, children, parents := NewNativeKeyManagement(), NewNativeKeyManagement(), NewNativeKeyManagement()
	alreadyCached := roaring.New()
	todoFutureCache := make(map[uint32]todoFuturePair)

	err := addCyclesToBindMap(scc, cache, children, parents, allNodes)
	if err != nil {
		return nil, fmt.Errorf("error adding cycles to bind map: %w", err)
	}

	nodesToCache := roaring.New()

	for _, nodeID := range uncachedNodes {
		nodesToCache.Add(nodeID)
	}

	nodesToProcess := nodesToCache.ToArray()
	for _, nodeID := range nodesToProcess {

		stack := []stackElm{{id: nodeID, todoIndex: 0}}

		for len(stack) > 0 {

			todoIndex := stack[len(stack)-1].todoIndex
			curNode := allNodes[stack[len(stack)-1].id]

			if alreadyCached.Contains(curNode.ID) {
				stack = stack[:len(stack)-1]
				continue
			}

			todoNodes, futureNodes, err := getTodoAndFutureNodesCached(children, parents, curNode, direction, todoFutureCache)
			if err != nil {
				return nil, err
			}

			if todoIndex == len(todoNodes) {
				alreadyCached.Add(curNode.ID)
				stack = stack[:len(stack)-1]

				for _, nextNode := range futureNodes {
					stack = append(stack, stackElm{id: nextNode, todoIndex: 0})
				}
			} else {
				if scc[curNode.ID] != scc[todoNodes[todoIndex]] {
					stack = append(stack[:len(stack)-1], stackElm{id: curNode.ID, todoIndex: todoIndex + 1})
					if alreadyCached.Contains(todoNodes[todoIndex]) {
						if err := addToCache(cache, curNode.ID, todoNodes[todoIndex]); err != nil {
							return nil, err
						}
					} else {
						stack = append(stack, stackElm{id: todoNodes[todoIndex], todoIndex: 0})
					}
				}
			}

		}
	}

	return cache, nil
}

func getTodoAndFutureNodesCached(children, parents *NativeKeyManagement, curNode *Node, direction Direction, cache map[uint32]todoFuturePair) ([]uint32, []uint32, error) {
	if pair, exists := cache[curNode.ID]; exists {
		return pair.todoNodes, pair.futureNodes, nil
	}

	todoNodes, futureNodes, err := getTodoAndFutureNodes(children, parents, curNode, direction)
	if err != nil {
		return nil, nil, err
	}

	cache[curNode.ID] = todoFuturePair{todoNodes: todoNodes, futureNodes: futureNodes}
	return todoNodes, futureNodes, nil
}

func addCyclesToBindMap(scc map[uint32]uint32, cache, children, parents *NativeKeyManagement, allNodes map[uint32]*Node) error {
	parentToKeys := map[uint32][]string{}

	for k, v := range scc {
		parentToKeys[v] = append(parentToKeys[v], utils.Uint32ToStr(k))
	}

	for _, keysForAParent := range parentToKeys {
		if _, err := cache.BindKeys(keysForAParent); err != nil {
			return err
		}
		if _, err := children.BindKeys(keysForAParent); err != nil {
			return err
		}
		if _, err := parents.BindKeys(keysForAParent); err != nil {
			return err
		}

		keycache := roaring.New()
		childrenCache, parentCache := roaring.New(), roaring.New()
		for _, key := range keysForAParent {
			intkey, err := utils.StrToUint32(key)
			if err != nil {
				return err
			}

			node := allNodes[uint32(intkey)]

			childrenCache.Or(node.Children)
			parentCache.Or(node.Parents)
			keycache.Add(uint32(intkey))
			keycache.Add(node.ID)
		}
		childrenCache.AndNot(keycache)
		parentCache.AndNot(keycache)
		if len(keysForAParent) > 0 {
			if err := cache.Set(keysForAParent[0], *keycache); err != nil {
				return fmt.Errorf("error setting value for key in cache %s, %w", keysForAParent[0], err)
			}
			if err := children.Set(keysForAParent[0], *childrenCache); err != nil {
				return fmt.Errorf("error setting value for key in grouped children relationship%s, %w", keysForAParent[0], err)
			}
			if err := parents.Set(keysForAParent[0], *parentCache); err != nil {
				return fmt.Errorf("error setting value for key in grouped parent relationship%s, %w", keysForAParent[0], err)
			}
		}

	}
	return nil
}

func getTodoAndFutureNodes(children, parents *NativeKeyManagement, curNode *Node, direction Direction) ([]uint32, []uint32, error) {
	var todoNodes, futureNodes []uint32
	nodeIDStr := utils.Uint32ToStr(curNode.ID)

	if direction == ChildrenDirection {
		todoNodesBitmap, err := children.Get(nodeIDStr)
		if err != nil {
			return nil, nil, err
		}
		futureNodesBitmap, err := parents.Get(nodeIDStr)
		if err != nil {
			return nil, nil, err
		}
		todoNodes, futureNodes = todoNodesBitmap.ToArray(), futureNodesBitmap.ToArray()
	} else {
		todoNodesBitmap, err := parents.Get(nodeIDStr)
		if err != nil {
			return nil, nil, err
		}
		futureNodesBitmap, err := children.Get(nodeIDStr)
		if err != nil {
			return nil, nil, err
		}
		todoNodes, futureNodes = todoNodesBitmap.ToArray(), futureNodesBitmap.ToArray()
	}

	return todoNodes, futureNodes, nil
}

func addToCache(bm *NativeKeyManagement, curElem, todoElem uint32) error {
	curElemStr := utils.Uint32ToStr(curElem)
	todoElemStr := utils.Uint32ToStr(todoElem)

	curElemVal, err := bm.Get(curElemStr)
	if err != nil {
		return fmt.Errorf("error getting value for curElem key from value %d: %w", curElem, err)
	}

	todoVal, err := bm.Get(todoElemStr)
	if err != nil {
		return fmt.Errorf("error getting value for todoElem key %d: %w", todoElem, err)
	}

	curElemVal.Or(&todoVal)
	curElemVal.Add(curElem)

	return bm.Set(curElemStr, curElemVal)
}
