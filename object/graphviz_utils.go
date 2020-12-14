package object

import (
	"fmt"
	"koko/ast"
)

func escapeStringForGraphviz(s string) string {
	out := ""
	for _, c := range s {
		switch c {
		case '"':
			out += "\\" + string(c)
			break
		default:
			out += string(c)
			break
		}
	}
	return out
}

func GetAllDependenciesToDotLang(result Object) string {
	seenNodes := make(map[Object]bool)
	queue := []Object{}
	queue = append(queue, result)
	out := "digraph G {\n"
	seenOutputEdges := make(map[string]bool)
	for len(queue) > 0 {
		head := queue[0]
		fmt.Printf("HEAD: %+v\n", head)
		if seenNodes[head] {
			if len(queue) > 1 {
				queue = queue[1:]
			} else {
				queue = []Object{}
			}
			continue
		}
		seenNodes[head] = true
		for link := range head.GetDependencyLinks() {
			// TODO (Peter lots of conditions here!!! Clean them up!)
			if head.GetCreatorNode() == nil {
				panic(fmt.Sprintf("Graph construction failed %+v\n", head))
			}
			if link.GetCreatorNode() == nil {
				panic(fmt.Sprintf("Graph construction failed %+v\n", link))
			}
			if _, ok := head.GetCreatorNode().(*ast.BuiltinValue); ok {
				continue
			}
			if _, ok := link.GetCreatorNode().(*ast.BuiltinValue); ok {
				continue
			}
			if head.GetCreatorNode().String() == link.GetCreatorNode().String() {
				// copied dependencies look like the node points to itself
				// they are condensed in this representation
				continue
			}
			headSpan := head.GetCreatorNode().Span()
			linkSpan := link.GetCreatorNode().Span()
			headNode := fmt.Sprintf("%s\n line: %d, pos: %d", head.GetCreatorNode().String(), headSpan.BeginLine, headSpan.BeginPos)
			linkNode := fmt.Sprintf("%s\n line: %d, pos: %d", link.GetCreatorNode().String(), linkSpan.BeginLine, linkSpan.BeginPos)
			edge := fmt.Sprintf("\t\"%s\" -> \"%s\";\n", escapeStringForGraphviz(linkNode), escapeStringForGraphviz(headNode))
			if _, ok := seenOutputEdges[edge]; !ok {
				seenOutputEdges[edge] = true
				out += edge
			}
		}
		if len(queue) > 1 {
			queue = queue[1:]
			for link := range head.GetDependencyLinks() {
				queue = append(queue, link)
			}
		} else {
			for link := range head.GetDependencyLinks() {
				queue = append(queue, link)
			}
		}
	}
	out += "}"
	return out
}
