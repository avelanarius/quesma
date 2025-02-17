// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0

package core

import "lexer/core"

// TODO: should we declare that Nodes are immutable?
type Node interface {
	String() string
}

type NodeListNode struct {
	Nodes []Node
}

func (n NodeListNode) String() string {
	result := "NodeListNode[\n"
	for i, node := range n.Nodes {
		if i > 0 {
			result += ",\n"
		}
		result += node.String()
	}
	result += "\n]"
	return result
}

type TokenNode struct {
	Token core.Token
}

func (n TokenNode) String() string {
	return "TokenNode[" + n.Token.String() + "]"
}

func TokensToNode(tokens []core.Token) Node {
	var nodes []Node

	for _, token := range tokens {
		nodes = append(nodes, TokenNode{Token: token})
	}

	return NodeListNode{Nodes: nodes}
}
