package graph

import (
	"fmt"
	"github.com/RoaringBitmap/roaring"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
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
	NodeName  *string `@Ident?` // NodeName is now optional - For example
}

var (
	simpleLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Ident", `[a-zA-Z][a-zA-Z0-9:/._@-]*`}, // Updated to handle colons, slashes, dots, underscores, hyphens, and @
		{"String", `"(?:\\.|[^"])*"`},
		{"Operator", `\b(?:and|or|xor)\b`},
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
	dependencies, dependents := collectPackages(expression, defaultNodeName)

	var nodeDependencies, nodeDependents []*Node

	for _, dependency := range dependencies {
		id := nameToIDs[dependency]
		nodeDependencies = append(nodeDependencies, nodes[id])
	}

	for _, dependent := range dependents {
		id := nameToIDs[dependent]
		nodeDependents = append(nodeDependents, nodes[id])
	}

	nodeToDependencies, err := BatchQueryDependencies(storage, nodeDependencies, caches, isCached)
	if err != nil {
		return nil, fmt.Errorf("Failed to get dependencies: %v", err)
	}
	nodeToDependents, err := BatchQueryDependents(storage, nodeDependents, caches, isCached)
	if err != nil {
		return nil, fmt.Errorf("Failed to get dependents: %v", err)
	}

	// Iterate through the parsed structure
	bm, err := iterateExpression(expression, nodeToDependencies, nodeToDependents, nameToIDs, defaultNodeName)
	if err != nil {
		return nil, fmt.Errorf("failed to iterate expression: %v", err)
	}

	return bm, nil
}

func collectPackages(expr *Expression, defaultNodeName string) ([]string, []string) {
	var dependencies []string
	var dependents []string
	collectPackagesFromExpression(expr, &dependencies, &dependents, defaultNodeName)
	return dependencies, dependents
}

func collectPackagesFromExpression(expr *Expression, dependencies, dependents *[]string, defaultNodeName string) {
	if expr == nil {
		return
	}

	collectPackagesFromTerm(expr.Left, dependencies, dependents, defaultNodeName)
	collectPackagesFromExpression(expr.Right, dependencies, dependents, defaultNodeName)
}

func collectPackagesFromTerm(term *Term, dependencies, dependents *[]string, defaultNodeName string) {
	if term == nil {
		return
	}

	if term.Query != nil {
		switch term.Query.QueryType {
		case "dependencies":
			if term.Query.NodeName != nil {
				*dependencies = append(*dependencies, *term.Query.NodeName)
			} else {
				*dependencies = append(*dependencies, defaultNodeName)
			}
		case "dependents":
			*dependents = append(*dependents, *term.Query.NodeName)
		}
	}

	if term.Expression != nil {
		collectPackagesFromExpression(term.Expression, dependencies, dependents, defaultNodeName)
	}
}

func iterateExpression(expr *Expression, dependencies, dependents map[uint32]*roaring.Bitmap, nameToIDs map[string]uint32, defaultNodeName string) (*roaring.Bitmap, error) {
	if expr == nil {
		return nil, nil
	}

	bm, err := iterateTerm(expr.Left, dependencies, dependents, nameToIDs, defaultNodeName)
	if err != nil {
		return nil, err
	}

	if expr.Op != nil {
		bm2, err := iterateExpression(expr.Right, dependencies, dependents, nameToIDs, defaultNodeName)

		if err != nil {
			return nil, err
		}

		switch *expr.Op {
		case "or":
			bm.Or(bm2)
		case "and":
			bm.And(bm2)
		case "xor":
			bm.Xor(bm2)
		default:
			return nil, fmt.Errorf("unknown operator: %s", *expr.Op)
		}
	}

	return bm, nil
}

func iterateTerm(term *Term, dependencies, dependents map[uint32]*roaring.Bitmap, nameToIDs map[string]uint32, defaultNodeName string) (*roaring.Bitmap, error) {
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
		case "dependencies":
			bm = dependencies[id]
		case "dependents":
			bm = dependents[id]
		default:
			return nil, fmt.Errorf("unknown query: %s", term.Query.QueryType)
		}
	}

	if term.Expression != nil {
		_, err := iterateExpression(term.Expression, dependencies, dependents, nameToIDs, defaultNodeName)
		if err != nil {
			return nil, err
		}
	}

	return bm, nil
}
