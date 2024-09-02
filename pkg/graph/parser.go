package graph

import (
	"fmt"
	"github.com/RoaringBitmap/roaring"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

const (
	dependencies = "dependencies"
	dependents   = "dependents"
	or           = "or"
	and          = "and"
	xor          = "xor"
)

// Define the grammar using Go structs and Participle tags
type Expression struct {
	Left  *Term       `@@`
	Op    *string     `@("and" | "or" | "xor")?`
	Right *Expression `@@?`
}

type Term struct {
	Query      *Query      `@@`
	Expression *Expression `| "(" @@ ")" | "[" @@ "]"`
}

type Query struct {
	QueryType string  `@Ident`  // For example "dependencies" or "dependents"
	NodeType  string  `@Ident`  // For example "library" or "vulns"
	NodeName  *string `@Ident?` // NodeName is now optional // The purl being inputted
}

var (
	simpleLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Operator", `\b(?:and|or|xor)\b`},      // Prioritize operators
		{"Ident", `[a-zA-Z][a-zA-Z0-9:/._@-]*`}, // Updated to handle colons, slashes, dots, underscores, hyphens, and @
		{"String", `"(?:\\.|[^"])*"`},
		{"Whitespace", `[ \t\n\r]+`},
		{"LBracket", `\[`},
		{"RBracket", `\]`},
		{"LParen", `\(`},
		{"RParen", `\)`},
	})
	parser = participle.MustBuild[Expression](
		participle.Lexer(simpleLexer),
		participle.Elide("Whitespace"),
	)
)

// ParseAndExecute parses and executes a script using the given storage backend.
func ParseAndExecute(script string, storage Storage, defaultNodeName string, nodes map[uint32]*Node, caches map[uint32]*NodeCache, isCached bool) (*roaring.Bitmap, error) {
	nameToIDs := make(map[string]uint32, len(nodes))
	for _, node := range nodes {
		nameToIDs[node.Name] = node.ID
	}

	expression, err := parser.ParseString("", script)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse expression: %v", err)
	}

	// Collect all packages for batch querying
	dependenciesToQuery, dependentsToQuery := collectPackages(expression, defaultNodeName)

	var nodeDependencies, nodeDependents []*Node

	for _, dependency := range dependenciesToQuery {
		id := nameToIDs[dependency]
		nodeDependencies = append(nodeDependencies, nodes[id])
	}

	for _, dependent := range dependentsToQuery {
		id := nameToIDs[dependent]
		nodeDependents = append(nodeDependents, nodes[id])
	}

	dependenciesForID, err := BatchQueryDependencies(storage, nodeDependencies, caches, isCached)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies from batch query: %v", err)
	}
	dependentsForID, err := BatchQueryDependents(storage, nodeDependents, caches, isCached)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependents from batch query: %v", err)
	}

	// Iterate through the parsed structure
	bm, err := iterateExpression(expression, dependenciesForID, dependentsForID, nameToIDs, defaultNodeName)
	if err != nil {
		return nil, fmt.Errorf("failed to iterate expression: %v", err)
	}

	return bm, nil
}

func collectPackages(expr *Expression, defaultNodeName string) ([]string, []string) {
	var dependenciesToQuery []string
	var dependentsToQuery []string
	collectPackagesFromExpression(expr, &dependenciesToQuery, &dependentsToQuery, defaultNodeName)
	return dependenciesToQuery, dependentsToQuery
}

func collectPackagesFromExpression(expr *Expression, dependenciesToQuery, dependentsToQuery *[]string, defaultNodeName string) {
	if expr == nil {
		return
	}

	collectPackagesFromTerm(expr.Left, dependenciesToQuery, dependentsToQuery, defaultNodeName)
	collectPackagesFromExpression(expr.Right, dependenciesToQuery, dependentsToQuery, defaultNodeName)
}

func collectPackagesFromTerm(term *Term, dependenciesToQuery, dependentsToQuery *[]string, defaultNodeName string) {
	if term == nil {
		return
	}

	if term.Query != nil {
		switch term.Query.QueryType {
		case dependencies:
			if term.Query.NodeName != nil {
				*dependenciesToQuery = append(*dependenciesToQuery, *term.Query.NodeName)
			} else {
				*dependenciesToQuery = append(*dependenciesToQuery, defaultNodeName)
			}
		case dependents:
			if term.Query.NodeName != nil {
				*dependentsToQuery = append(*dependentsToQuery, *term.Query.NodeName)
			} else {
				*dependentsToQuery = append(*dependentsToQuery, defaultNodeName)
			}
		}
	}

	if term.Expression != nil {
		collectPackagesFromExpression(term.Expression, dependenciesToQuery, dependentsToQuery, defaultNodeName)
	}
}

func iterateExpression(expr *Expression, dependenciesForID, dependentsForID map[uint32]*roaring.Bitmap, nameToIDs map[string]uint32, defaultNodeName string) (*roaring.Bitmap, error) {
	if expr == nil {
		return nil, nil
	}

	bm, err := iterateTerm(expr.Left, dependenciesForID, dependentsForID, nameToIDs, defaultNodeName)
	if err != nil {
		return nil, err
	}

	if expr.Op != nil {
		bm2, err := iterateExpression(expr.Right, dependenciesForID, dependentsForID, nameToIDs, defaultNodeName)

		if err != nil {
			return nil, err
		}

		switch *expr.Op {
		case or:
			bm.Or(bm2)
		case and:
			bm.And(bm2)
		case xor:
			bm.Xor(bm2)
		default:
			return nil, fmt.Errorf("unknown operator: %s", *expr.Op)
		}
	}

	return bm, nil
}

func iterateTerm(term *Term, dependenciesForID, dependentsForID map[uint32]*roaring.Bitmap, nameToIDs map[string]uint32, defaultNodeName string) (*roaring.Bitmap, error) {
	if term == nil {
		return nil, nil
	}

	bm := roaring.New()

	if term.Query != nil {
		id := uint32(0)
		if term.Query.NodeName != nil {
			id = nameToIDs[*term.Query.NodeName]
		} else {
			id = nameToIDs[defaultNodeName]
		}

		switch term.Query.QueryType {
		case dependencies:
			bm = dependenciesForID[id]
		case dependents:
			bm = dependentsForID[id]
		default:
			return nil, fmt.Errorf("unknown query: %s", term.Query.QueryType)
		}
	}

	if term.Expression != nil {
		_, err := iterateExpression(term.Expression, dependenciesForID, dependentsForID, nameToIDs, defaultNodeName)
		if err != nil {
			return nil, err
		}
	}

	return bm, nil
}
