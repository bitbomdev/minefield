package pkg

import (
	"fmt"
	"strconv"

	"github.com/RoaringBitmap/roaring"
)

type stackElm struct {
	id        uint32
	todoIndex int
}

func Cache[T any](storage Storage[T]) error {
	uncachedNodes, err := storage.ToBeCached()
	if err != nil {
		return err
	}

	keys, err := storage.GetAllKeys()
	if err != nil {
		return fmt.Errorf("error getting keys")
	}

	childSCC, err := findCycles(storage, "children", len(keys))
	if err != nil {
		return err
	}

	cachedChildren, err := buildNodeCacheMap(storage, uncachedNodes, "children", childSCC)
	if err != nil {
		return err
	}

	parentSCC, err := findCycles(storage, "parent", len(keys))
	if err != nil {
		return err
	}

	cachedParents, err := buildNodeCacheMap(storage, uncachedNodes, "parents", parentSCC)
	if err != nil {
		return err
	}

	cachedChildKeys, cachedChildValues, err := cachedChildren.GetAllKeysAndValues()
	if err != nil {
		return err
	}

	for i := 0; i < len(cachedChildKeys); i++ {
		childId := cachedChildKeys[i]
		childIntId, err := strconv.Atoi(childId)
		if err != nil {
			return err
		}

		childBindValue := cachedChildValues[i].Clone()
		childBindValue.Remove(uint32(childIntId))

		tempValue, err := cachedParents.Get(strconv.Itoa(childIntId))
		if err != nil {
			return fmt.Errorf("error getting value for key %s, err: %v", childId, err)
		}
		parentBindValue := tempValue.Clone()
		parentBindValue.Remove(uint32(childIntId))

		if err := storage.SaveCache(NewNodeCache(uint32(childIntId), parentBindValue, childBindValue)); err != nil {
			return err
		}
	}

	return storage.ClearCacheStack()
}

func findCycles[T any](storage Storage[T], direction string, numOfNodes int) (map[uint32]uint32, error) {
	var stack []uint32
	var tarjanDFS func(nodeID uint32) error

	currentTarjanID := 0
	nodeToTarjanID := map[uint32]uint32{}
	lowLink := make(map[uint32]uint32)
	inStack := roaring.New()

	tarjanDFS = func(nodeID uint32) error {
		currentNode, err := storage.GetNode(nodeID)
		if err != nil {
			return err
		}

		var nextNodes []uint32
		if direction == "children" {
			nextNodes = currentNode.Child.ToArray()
		} else {
			nextNodes = currentNode.Parent.ToArray()
		}

		currentTarjanID++
		stack = append(stack, nodeID)
		inStack.Add(nodeID)
		nodeToTarjanID[nodeID] = uint32(currentTarjanID)
		lowLink[nodeID] = uint32(currentTarjanID)

		for _, nextNode := range nextNodes {
			if _, visited := nodeToTarjanID[nextNode]; !visited {
				if err := tarjanDFS(nextNode); err != nil {
					return err
				}
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
		return nil
	}

	for id := 1; id < numOfNodes+1; id++ {
		if _, visited := nodeToTarjanID[uint32(id)]; !visited {
			if err := tarjanDFS(uint32(id)); err != nil {
				return nil, err
			}
		}
	}

	return lowLink, nil
}

func buildNodeCacheMap[T any](storage Storage[T], uncachedNodes []uint32, direction string, scc map[uint32]uint32) (*NativeKeyManagement, error) {
	cache, children, parents := NewNativeKeyManagement(), NewNativeKeyManagement(), NewNativeKeyManagement()
	alreadyCached := roaring.New()

	err := addCyclesToBindMap[T](storage, scc, cache, children, parents)
	if err != nil {
		return nil, err
	}

	for _, nodeID := range uncachedNodes {
		stack := []stackElm{{nodeID, 0}}

		for len(stack) > 0 {
			curTodoIndex := stack[len(stack)-1].todoIndex
			curNode, err := storage.GetNode(stack[len(stack)-1].id) // Retrieve the current node from storage
			if err != nil {
				return nil, err
			}

			if alreadyCached.Contains(curNode.Id) {
				stack = stack[:len(stack)-1] // pop off the stack, this node is already cached
				continue
			}

			todoNodes, futureNodes, err := getTodoAndFutureNodes[T](children, parents, curNode, direction)

			if curTodoIndex == len(todoNodes) {
				stack = stack[:len(stack)-1] // pop off the stack, this node is now fully cached
				alreadyCached.Add(curNode.Id)

				for _, futureNodeID := range futureNodes {
					stack = append(stack, stackElm{id: futureNodeID, todoIndex: 0})
				}
			} else {
				stack = append(stack[:len(stack)-1], stackElm{id: curNode.Id, todoIndex: curTodoIndex + 1}) // move the current node to its next todo node
				if !(scc[curNode.Id] == scc[todoNodes[curTodoIndex]]) {
					if alreadyCached.Contains(todoNodes[curTodoIndex]) {
						if err := addToCache(cache, curNode.Id, todoNodes[curTodoIndex]); err != nil {
							return nil, err
						}
					} else {
						stack = append(stack, stackElm{id: todoNodes[curTodoIndex], todoIndex: 0})
					}
				}
			}
		}
	}

	return cache, nil
}

func getTodoAndFutureNodes[T any](children, parents *NativeKeyManagement, curNode *Node[T], direction string) ([]uint32, []uint32, error) {
	var todoNodes, futureNodes []uint32

	if direction == "children" {
		todoNodesBitmap, err := children.Get(strconv.Itoa(int(curNode.Id)))
		if err != nil {
			return nil, nil, err
		}
		futureNodesBitmap, err := parents.Get(strconv.Itoa(int(curNode.Id)))
		if err != nil {
			return nil, nil, err
		}
		todoNodes, futureNodes = todoNodesBitmap.ToArray(), futureNodesBitmap.ToArray()
	} else {
		todoNodesBitmap, err := parents.Get(strconv.Itoa(int(curNode.Id)))
		if err != nil {
			return nil, nil, err
		}
		futureNodesBitmap, err := children.Get(strconv.Itoa(int(curNode.Id)))
		if err != nil {
			return nil, nil, err
		}
		todoNodes, futureNodes = todoNodesBitmap.ToArray(), futureNodesBitmap.ToArray()
	}

	return todoNodes, futureNodes, nil
}

// addCyclesToBindMap takes in a union find which contains all the cycles found in the graph and
// a bind map which we use to store all children and parents for a given node.
// This function takes the data from the union find and adds it to the bind map so that we can
// initialize the bind map with all the cycles.
func addCyclesToBindMap[T any](storage Storage[T], scc map[uint32]uint32, cache, children, parents *NativeKeyManagement) error {
	parentToKeys := map[uint32][]string{}

	for k, v := range scc {
		parentToKeys[v] = append(parentToKeys[v], strconv.Itoa(int(k)))
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
			intkey, err := strconv.Atoi(key)
			if err != nil {
				return err
			}

			node, err := storage.GetNode(uint32(intkey))
			if err != nil {
				return err
			}

			childrenCache.Or(node.Child)
			parentCache.Or(node.Parent)
			keycache.Add(uint32(intkey))

		}
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

func addToCache(bm *NativeKeyManagement, curElem, todoElem uint32) error {
	curElemVal, err := bm.Get(strconv.Itoa(int(curElem)))
	if err != nil {
		return fmt.Errorf("error getting value for curElem keys from value %d, err: %v", curElem, err)
	}

	todoVal, err := bm.Get(strconv.Itoa(int(todoElem)))
	if err != nil {
		return fmt.Errorf("error getting value for curElem key %d, err: %v", todoElem, err)
	}

	curElemVal.Or(&todoVal)
	curElemVal.Add(curElem)

	err = bm.Set(strconv.Itoa(int(curElem)), curElemVal)
	if err != nil {
		return err
	}
	return nil
}
