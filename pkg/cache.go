package pkg

import (
	"fmt"
	"strconv"

	"github.com/RoaringBitmap/roaring"
)

type stackElm struct {
	id   uint32
	from uint32
}

type key struct {
	id           uint32
	visitedNodes *roaring.Bitmap
	visited      map[uint32]*roaring.Bitmap
}

func findCycles[T any](storage Storage[T], direction string, startNode int, numOfNodes int) (*unionFind, error) {
	totalVisited := map[uint32]bool{}

	parents := make([]uint32, numOfNodes+1)
	for i := 0; i < numOfNodes+1; i++ {
		parents[i] = uint32(i)
	}
	uf := &unionFind{
		count:   numOfNodes,
		parents: parents,
	}

	for i := startNode; i <= numOfNodes; i++ {
		if totalVisited[uint32(i)] {
			continue
		}

		stack := []key{{uint32(i), roaring.New(), make(map[uint32]*roaring.Bitmap)}}

		for len(stack) > 0 {
			curElm := stack[len(stack)-1] // Get the top element of the stack
			stack = stack[:len(stack)-1]
			curNode, err := storage.GetNode(curElm.id) // Retrieve the current node from storage
			if err != nil {
				return nil, err
			}

			var nextNodes []uint32

			if direction == "children" {
				nextNodes = curNode.Child.ToArray()
			} else {
				nextNodes = curNode.Parent.ToArray()
			}

			visitedNodes := curElm.visitedNodes.Clone()

			visitedNodes.Add(curElm.id)

			if _, ok := curElm.visited[curElm.id]; ok {
				curElm.visited[curElm.id].Xor(visitedNodes) // find the nodes in the cycle
				curElm.visited[curElm.id].Add(curElm.id)

				arr := curElm.visited[curElm.id].ToArray()
				for i := 1; i < len(arr); i++ {
					uf.Union(arr[i-1], arr[i])
				}
				uf.Union(arr[0], arr[len(arr)-1])

				continue
			} else {
				totalVisited[curElm.id] = true

				for _, node := range nextNodes {
					newVisited := make(map[uint32]*roaring.Bitmap)
					for k, v := range curElm.visited {
						newVisited[k] = v
					}

					newVisited[curElm.id] = visitedNodes.Clone()
					stack = append(stack, key{node, newVisited[curElm.id], newVisited})
				}
			}
		}
	}

	return uf, nil
}

func buildNodeCacheMap[T any](storage Storage[T], uncachedNodes []uint32, direction string, uf *unionFind) (*NativeKeyManagement, error) {
	alreadyCached := roaring.New() // Tracks nodes whose children have already been cached.
	processed := roaring.New()     // Tracks nodes that are being processed, so we do not have to re-add nodes to the stack, we cannot use this when adding mustFinish nodes, since for those order matters
	allChildAndParentsCached := roaring.New()

	bm := NewNativeKeyManagement()

	err := addCyclesToBindMap(uf, bm)
	if err != nil {
		return nil, err
	}

	curTodo := make(map[uint32]int)

	for _, nodeID := range uncachedNodes {
		stack := []stackElm{{nodeID, nodeID}} // Initialize stack with the current node to process
		curTodo[nodeID] = 0
		processed.Add(nodeID)

		for len(stack) > 0 {
			curElm := stack[len(stack)-1]              // Get the top element of the stack
			curNode, err := storage.GetNode(curElm.id) // Retrieve the current node from storage
			if err != nil {
				return nil, err
			}

			if allChildAndParentsCached.Contains(curElm.id) {
				stack = stack[:len(stack)-1] // Pop the current element from the stack if already cached
				continue
			}

			var todoNodes, futureNodes []uint32 // Nodes that must be processed before and after the current node

			if direction == "children" {
				todoNodes = curNode.GetChildren().ToArray()  // Nodes that the current node relies on
				futureNodes = curNode.GetParents().ToArray() // Nodes that rely on the current node
			} else {
				todoNodes = curNode.GetParents().ToArray()    // Nodes that rely on the current node
				futureNodes = curNode.GetChildren().ToArray() // Nodes that the current node relies on
			}

			if curTodo[curElm.id] == len(todoNodes) { // No more nodes are needed to be processed to cache the current node
				err = setBitmapValueWithChild(bm, curElm.from, curElm.id)
				if err != nil {
					return nil, err
				}

				stack = stack[:len(stack)-1] // This node is fully processed, pop from stack, no to-do nodes left
				alreadyCachedAllParents := true
				for _, futureNode := range futureNodes {
					if !processed.Contains(futureNode) {
						processed.Add(futureNode)
						stack = append(stack, stackElm{futureNode, futureNode}) // Push new nodes to be processed onto the stack, now that all nodes that must be processed are done
						curTodo[futureNode] = 0
					}

					if !alreadyCached.Contains(futureNode) {
						alreadyCachedAllParents = false
					}
				}
				alreadyCached.Add(curElm.id) // Mark the current node as cached
				if alreadyCachedAllParents {
					allChildAndParentsCached.Add(curElm.id)
				}
			} else if len(todoNodes) > 0 { // There are more nodes to be processed to cache the current node
				todoID := uint32(0)

				if curTodo[curElm.id] == len(todoNodes) {
					todoID = todoNodes[curTodo[curElm.id]-1] // Get the next node to process
				} else {
					todoID = todoNodes[curTodo[curElm.id]] // Get the next node to process
					curTodo[curElm.id] += 1
				}

				if alreadyCached.Contains(todoID) { // If the next node is already cached, we can add its cache to the current node's cache
					err = setBitmapValueWithChild(bm, curElm.id, todoID)
					if err != nil {
						return nil, err
					}
				} else {
					stack = append(stack, stackElm{todoID, curElm.id}) // Push the dependency to the stack to process its dependencies before curElm
					if !processed.Contains(todoID) {
						curTodo[todoID] = 0
						processed.Add(todoID)
					}
				}
			}
		}
	}

	return bm, nil
}

// addCyclesToBindMap takes in a union find which contains all the cycles found in the graph and
// a bind map which we use to store all children and parents for a given node.
// This function takes the data from the union find and adds it to the bind map so that we can
// initialize the bind map with all the cycles.
func addCyclesToBindMap(uf *unionFind, bm *NativeKeyManagement) error {
	parentToKeys := map[uint32][]string{}

	for i := 1; i < len(uf.parents); i++ {
		parentToKeys[uf.parents[i]] = append(parentToKeys[uf.parents[i]], strconv.Itoa(i))
	}

	for _, keysForAParent := range parentToKeys {
		_, err := bm.BindKeys(keysForAParent)
		if err != nil {
			return err
		}
	}

	for _, keysForAParent := range parentToKeys {
		for _, key := range keysForAParent {
			got, err := bm.Get(key)
			if err != nil {
				return fmt.Errorf("error getting key from parents key, %w", err)
			}

			n, err := strconv.Atoi(key)
			if err != nil {
				return fmt.Errorf("error converting string to integer, %w", err)
			}
			got.Add(uint32(n))

			err = bm.Set(key, got)
			if err != nil {
				return fmt.Errorf("error setting value for key %s, %w", key, err)
			}
		}
	}
	return nil
}

func setBitmapValueWithChild(bm *NativeKeyManagement, curElem, todoElem uint32) error {
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

func Cache[T any](storage Storage[T]) error {
	uncachedNodes, err := storage.ToBeCached()
	if err != nil {
		return err
	}

	keys, err := storage.GetAllKeys()
	if err != nil {
		return fmt.Errorf("error getting keys")
	}

	childUf, err := findCycles(storage, "children", 1, len(keys))
	if err != nil {
		return err
	}

	cachedChildren, err := buildNodeCacheMap(storage, uncachedNodes, "children", childUf)
	if err != nil {
		return err
	}

	parentUf, err := findCycles(storage, "parent", 1, len(keys))
	if err != nil {
		return err
	}

	cachedParents, err := buildNodeCacheMap(storage, uncachedNodes, "parents", parentUf)
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
