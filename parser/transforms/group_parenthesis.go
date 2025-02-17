// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0

package transforms

import (
	"fmt"
	"parser/core"
	"slices"
)

// Based on sqlparse: https://github.com/andialbrecht/sqlparse/blob/a801100e9843786a9139bebb97c951603637129c/sqlparse/engine/grouping.py#L56C11-L57

// TODO: visitor pattern?

func GroupParenthesis(node core.Node) (core.Node, error) {
	nodeList, ok := node.(core.NodeListNode)
	if !ok {
		// TODO: this should recurse into the node generally
		return node, nil
	}

	groupedNodes := slices.Clone(nodeList.Nodes)

	// This is a bit unintuitive, but we will iterate in reverse order
	// and modify the list in place, grouping parenthesized subnodes to NodeListNode.
	parensCount := 0
	lastParenEnd := -1

	startIter, endIter := len(groupedNodes)-1, 0

	// Special case: don't group if the first node is '(' and the last node is ')'
	if len(groupedNodes) >= 2 {
		if tokenNode, ok := groupedNodes[0].(core.TokenNode); ok {
			if tokenNode.Token.RawValue == "(" {
				if tokenNode, ok := groupedNodes[len(groupedNodes)-1].(core.TokenNode); ok {
					if tokenNode.Token.RawValue == ")" {
						startIter, endIter = len(groupedNodes)-2, 1
					}
				}
			}
		}
	}

	for i := startIter; i >= endIter; i-- {
		tokenNode, ok := groupedNodes[i].(core.TokenNode)
		if !ok {
			continue
		}

		if tokenNode.Token.RawValue == ")" {
			parensCount++
			if parensCount == 1 {
				lastParenEnd = i
			}
		}
		if tokenNode.Token.RawValue == "(" {
			parensCount--
			if parensCount == 0 {
				// We have found a valid top-level pair of parenthesis
				// Group them!

				newGroupedNodes := slices.Clone(groupedNodes[:i])
				newGroupedNodes = append(newGroupedNodes, core.NodeListNode{Nodes: groupedNodes[i : lastParenEnd+1]})
				newGroupedNodes = append(newGroupedNodes, groupedNodes[lastParenEnd+1:]...)

				groupedNodes = newGroupedNodes
			} else if parensCount < 0 {
				return nil, fmt.Errorf("unbalanced parenthesis") // TODO: better error message, maybe even this transform should be nonvalidating?
			}
		}
	}

	for i := 0; i < len(groupedNodes); i++ {
		var err error
		groupedNodes[i], err = GroupParenthesis(groupedNodes[i])
		if err != nil {
			return nil, err
		}
	}

	return &core.NodeListNode{Nodes: groupedNodes}, nil
}
