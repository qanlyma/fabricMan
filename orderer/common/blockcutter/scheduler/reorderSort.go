// fabricMan

package scheduler

import (
	"math"
)

func ReorderSort(graph, invgraph [][]int32) ([]int32, error) {
	indegree := make(map[int32]int)
	outdegree := make(map[int32]int)
	nodeset := make(map[int32]bool)

	// Calculate in-degrees using invgraph
	for i := 0; i < len(invgraph); i++ {
		for node := range invgraph {
			indegree[int32(node)] = len(invgraph[int32(node)])
			nodeset[int32(node)] = true
		}
	}
	// Calculate out-degrees using graph
	for i := 0; i < len(graph); i++ {
		for node := range graph {
			outdegree[int32(node)] = len(graph[int32(node)])
		}
	}

	var result []int32
	var nodeToSort int32

	for len(nodeset) > 0 {
		// Find the node with min in-dgree
		minIndegree := math.MaxInt32
		for node := range nodeset {
			if indegree[node] < minIndegree {
				minIndegree = indegree[node]
				nodeToSort = node
			} else if indegree[node] == minIndegree && outdegree[node] < outdegree[nodeToSort] {
				nodeToSort = node
			}
		}

		// remove nodes which affet nodeToSort
		for _, nodeToRemove := range invgraph[nodeToSort] {
			if exist := nodeset[nodeToRemove]; !exist {
				continue
			}
			delete(nodeset, nodeToRemove)
			for _, v := range invgraph[nodeToRemove] {
				outdegree[v]--
			}
			for _, v := range graph[nodeToRemove] {
				indegree[v]--
			}
		}
		result = append(result, nodeToSort)
		for _, v := range graph[nodeToSort] {
			indegree[v]--
		}
		delete(nodeset, nodeToSort)
	}

	reverse(result)
	return result, nil
}

func reverse(slice []int32) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}
