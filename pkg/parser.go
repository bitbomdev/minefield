package pkg

import (
	"fmt"
	"strings"

	"github.com/RoaringBitmap/roaring"
)

// ParseAndExecute parses and executes a script using the given storage backend.
func ParseAndExecute(script string, storage Storage, defaultNodeName string) (*roaring.Bitmap, error) {
	var stack []*roaring.Bitmap
	var operators []string

	applyOperator := func() error {
		if len(stack) < 2 {
			return fmt.Errorf("not enough operands for operation")
		}
		right := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		left := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if len(operators) == 0 {
			return fmt.Errorf("operator stack underflow")
		}
		op := operators[len(operators)-1]
		operators = operators[:len(operators)-1]

		var result *roaring.Bitmap
		switch op {
		case "or":
			result = roaring.Or(left, right)
		case "xor":
			result = roaring.Xor(left, right)
		case "and":
			result = roaring.And(left, right)
		default:
			return fmt.Errorf("unknown operator: %s", op)
		}

		stack = append(stack, result)
		return nil
	}

	tokens := strings.Fields(script) // Split script into tokens based on spaces
	for tokenIndex := 0; tokenIndex < len(tokens); tokenIndex++ {
		token := tokens[tokenIndex]
		switch {
		case strings.HasPrefix(token, "dependents") || strings.HasPrefix(token, "dependencies"):
			dir := token
			tokenIndex++
			nodeTypeQueried := tokens[tokenIndex]
			if defaultNodeName != "" {
				token = defaultNodeName
			} else {
				tokenIndex++
				token = tokens[tokenIndex]
			}
			nodeID, err := storage.NameToID(token)
			if err != nil {
				return nil, fmt.Errorf("failed to get node ID for name %s: %w", token, err)
			}
			var bitmap *roaring.Bitmap
			if strings.TrimSpace(dir) == "dependents" {
				bitmap, err = storage.QueryDependents(nodeID)
			} else {
				bitmap, err = storage.QueryDependencies(nodeID)
			}
			if err != nil {
				return nil, fmt.Errorf("failed to query dependents for node ID %d: %w", nodeID, err)
			}
			if bitmap == nil {
				continue
			}

			for _, id := range bitmap.ToArray() {
				node, err := storage.GetNode(id)
				if err != nil {
					return nil, err
				}

				if node.Type != nodeTypeQueried {
					bitmap.Remove(id)
				}
			}

			stack = append(stack, bitmap)

		case token == "or", token == "xor", token == "and":
			// Before pushing new operator, apply any previous operators if not blocked by '('
			for len(operators) > 0 && operators[len(operators)-1] != "[" {
				if err := applyOperator(); err != nil {
					return nil, err
				}
			}
			operators = append(operators, token)

		case token == "[":
			operators = append(operators, token)

		case token == "]":
			// Apply all operators until the opening '('
			for len(operators) > 0 && operators[len(operators)-1] != "[" {
				if err := applyOperator(); err != nil {
					return nil, err
				}
			}
			if len(operators) == 0 {
				return nil, fmt.Errorf("mismatched parentheses")
			}
			// Pop the '(' operator
			operators = operators[:len(operators)-1]

		default:
			return nil, fmt.Errorf("unrecognized token: %s", token)
		}
	}

	// Apply remaining operators
	for len(operators) > 0 {
		if operators[len(operators)-1] == "[" {
			return nil, fmt.Errorf("mismatched parentheses")
		}
		if err := applyOperator(); err != nil {
			return nil, err
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("invalid expression or incomplete operations")
	}
	return stack[0], nil
}
