package interpolate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTraverse(t *testing.T) {
	t.Run("passes through to traverseAndEvaluate", func(t *testing.T) {
		input := map[string]any{"value": 42}
		variables := map[string]any{"$var": "test"}

		result, err := NewTraverse("simple string", input, variables)
		require.NoError(t, err)
		assert.Equal(t, "simple string", result)
	})
}

func TestTraverseAndEvaluate_String(t *testing.T) {
	t.Run("returns plain string as-is", func(t *testing.T) {
		result, err := traverseAndEvaluate("hello world", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)
	})

	t.Run("trims whitespace from string", func(t *testing.T) {
		result, err := traverseAndEvaluate("  hello world  ", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)
	})

	t.Run("trims newlines from string", func(t *testing.T) {
		result, err := traverseAndEvaluate("\n\thello world\n\t", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)
	})

	t.Run("evaluates strict expression with input", func(t *testing.T) {
		input := map[string]any{"name": "John"}
		result, err := traverseAndEvaluate("${ .name }", input, nil)
		require.NoError(t, err)
		assert.Equal(t, "John", result)
	})

	t.Run("evaluates nested path expression", func(t *testing.T) {
		input := map[string]any{
			"user": map[string]any{
				"profile": map[string]any{
					"email": "test@example.com",
				},
			},
		}
		result, err := traverseAndEvaluate("${ .user.profile.email }", input, nil)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})

	t.Run("evaluates expression with variables", func(t *testing.T) {
		variables := map[string]any{"$myVar": "hello"}
		result, err := traverseAndEvaluate("${ $myVar }", nil, variables)
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})
}

func TestTraverseAndEvaluate_Map(t *testing.T) {
	t.Run("returns empty map as-is", func(t *testing.T) {
		result, err := traverseAndEvaluate(map[string]any{}, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{}, result)
	})

	t.Run("preserves plain values in map", func(t *testing.T) {
		node := map[string]any{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}
		result, err := traverseAndEvaluate(node, nil, nil)
		require.NoError(t, err)

		resultMap := result.(map[string]any)
		assert.Equal(t, "value1", resultMap["key1"])
		assert.Equal(t, 42, resultMap["key2"])
		assert.Equal(t, true, resultMap["key3"])
	})

	t.Run("evaluates expressions in map values", func(t *testing.T) {
		node := map[string]any{
			"greeting": "${ .message }",
			"static":   "hello",
		}
		input := map[string]any{"message": "world"}

		result, err := traverseAndEvaluate(node, input, nil)
		require.NoError(t, err)

		resultMap := result.(map[string]any)
		assert.Equal(t, "world", resultMap["greeting"])
		assert.Equal(t, "hello", resultMap["static"])
	})

	t.Run("evaluates nested maps", func(t *testing.T) {
		node := map[string]any{
			"outer": map[string]any{
				"inner": "${ .value }",
			},
		}
		input := map[string]any{"value": "nested_result"}

		result, err := traverseAndEvaluate(node, input, nil)
		require.NoError(t, err)

		resultMap := result.(map[string]any)
		outerMap := resultMap["outer"].(map[string]any)
		assert.Equal(t, "nested_result", outerMap["inner"])
	})

	t.Run("returns nil for missing path in jq", func(t *testing.T) {
		node := map[string]any{
			"bad": "${ .nonexistent.deeply.nested }",
		}
		input := map[string]any{}

		result, err := traverseAndEvaluate(node, input, nil)
		// jq returns null for missing paths, not an error
		require.NoError(t, err)
		resultMap := result.(map[string]any)
		assert.Nil(t, resultMap["bad"])
	})
}

func TestTraverseAndEvaluate_Array(t *testing.T) {
	t.Run("returns empty array as-is", func(t *testing.T) {
		result, err := traverseAndEvaluate([]any{}, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, []any{}, result)
	})

	t.Run("preserves plain values in array", func(t *testing.T) {
		node := []any{"a", "b", "c"}
		result, err := traverseAndEvaluate(node, nil, nil)
		require.NoError(t, err)

		resultArr := result.([]any)
		assert.Equal(t, []any{"a", "b", "c"}, resultArr)
	})

	t.Run("evaluates expressions in array elements", func(t *testing.T) {
		node := []any{"${ .first }", "static", "${ .second }"}
		input := map[string]any{"first": "A", "second": "B"}

		result, err := traverseAndEvaluate(node, input, nil)
		require.NoError(t, err)

		resultArr := result.([]any)
		assert.Equal(t, "A", resultArr[0])
		assert.Equal(t, "static", resultArr[1])
		assert.Equal(t, "B", resultArr[2])
	})

	t.Run("evaluates nested arrays", func(t *testing.T) {
		node := []any{
			[]any{"${ .val }"},
		}
		input := map[string]any{"val": "inner"}

		result, err := traverseAndEvaluate(node, input, nil)
		require.NoError(t, err)

		resultArr := result.([]any)
		innerArr := resultArr[0].([]any)
		assert.Equal(t, "inner", innerArr[0])
	})

	t.Run("evaluates maps within arrays", func(t *testing.T) {
		node := []any{
			map[string]any{"key": "${ .value }"},
		}
		input := map[string]any{"value": "result"}

		result, err := traverseAndEvaluate(node, input, nil)
		require.NoError(t, err)

		resultArr := result.([]any)
		mapItem := resultArr[0].(map[string]any)
		assert.Equal(t, "result", mapItem["key"])
	})
}

func TestTraverseAndEvaluate_OtherTypes(t *testing.T) {
	t.Run("returns int as-is", func(t *testing.T) {
		result, err := traverseAndEvaluate(42, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("returns float as-is", func(t *testing.T) {
		result, err := traverseAndEvaluate(3.14, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
	})

	t.Run("returns bool as-is", func(t *testing.T) {
		result, err := traverseAndEvaluate(true, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("returns nil as-is", func(t *testing.T) {
		result, err := traverseAndEvaluate(nil, nil, nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestEvaluateJQExpression(t *testing.T) {
	t.Run("evaluates simple path expression", func(t *testing.T) {
		input := map[string]any{"foo": "bar"}
		result, err := evaluateJQExpression(".foo", input, nil)
		require.NoError(t, err)
		assert.Equal(t, "bar", result)
	})

	t.Run("evaluates identity expression", func(t *testing.T) {
		input := map[string]any{"a": 1, "b": 2}
		result, err := evaluateJQExpression(".", input, nil)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("evaluates arithmetic expression", func(t *testing.T) {
		input := map[string]any{"x": 10, "y": 5}
		result, err := evaluateJQExpression(".x + .y", input, nil)
		require.NoError(t, err)
		assert.Equal(t, 15, result)
	})

	t.Run("evaluates with variables", func(t *testing.T) {
		variables := map[string]any{"$name": "Alice"}
		result, err := evaluateJQExpression("$name", nil, variables)
		require.NoError(t, err)
		assert.Equal(t, "Alice", result)
	})

	t.Run("evaluates with multiple variables", func(t *testing.T) {
		variables := map[string]any{
			"$a": 10,
			"$b": 20,
		}
		result, err := evaluateJQExpression("$a + $b", nil, variables)
		require.NoError(t, err)
		assert.Equal(t, 30, result)
	})

	t.Run("evaluates array index", func(t *testing.T) {
		input := map[string]any{"items": []any{"first", "second", "third"}}
		result, err := evaluateJQExpression(".items[1]", input, nil)
		require.NoError(t, err)
		assert.Equal(t, "second", result)
	})

	t.Run("evaluates pipe expression", func(t *testing.T) {
		input := map[string]any{
			"users": []any{
				map[string]any{"name": "Alice"},
				map[string]any{"name": "Bob"},
			},
		}
		result, err := evaluateJQExpression(".users | length", input, nil)
		require.NoError(t, err)
		assert.Equal(t, 2, result)
	})

	t.Run("returns error for invalid jq syntax", func(t *testing.T) {
		_, err := evaluateJQExpression(".foo[", nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse jq expression")
	})

	t.Run("returns error for undefined variable", func(t *testing.T) {
		_, err := evaluateJQExpression("$undefined", nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to compile jq expression")
	})
}

func TestGetVariableNamesAndValues(t *testing.T) {
	t.Run("returns empty slices for empty map", func(t *testing.T) {
		names, values := getVariableNamesAndValues(map[string]any{})
		assert.Empty(t, names)
		assert.Empty(t, values)
	})

	t.Run("returns matching names and values", func(t *testing.T) {
		vars := map[string]any{
			"$a": 1,
			"$b": "two",
			"$c": true,
		}
		names, values := getVariableNamesAndValues(vars)

		assert.Len(t, names, 3)
		assert.Len(t, values, 3)

		// Create a map to verify pairs match
		result := make(map[string]any)
		for i, name := range names {
			result[name] = values[i]
		}

		assert.Equal(t, 1, result["$a"])
		assert.Equal(t, "two", result["$b"])
		assert.Equal(t, true, result["$c"])
	})

	t.Run("handles nil values", func(t *testing.T) {
		vars := map[string]any{
			"$nil": nil,
		}
		names, values := getVariableNamesAndValues(vars)

		assert.Equal(t, []string{"$nil"}, names)
		assert.Equal(t, []any{nil}, values)
	})
}

func TestComplexScenarios(t *testing.T) {
	t.Run("deeply nested structure with expressions", func(t *testing.T) {
		node := map[string]any{
			"level1": map[string]any{
				"level2": []any{
					map[string]any{
						"value": "${ .data.item }",
					},
				},
			},
		}
		input := map[string]any{
			"data": map[string]any{
				"item": "deep_value",
			},
		}

		result, err := traverseAndEvaluate(node, input, nil)
		require.NoError(t, err)

		resultMap := result.(map[string]any)
		level1 := resultMap["level1"].(map[string]any)
		level2 := level1["level2"].([]any)
		innerMap := level2[0].(map[string]any)
		assert.Equal(t, "deep_value", innerMap["value"])
	})

	t.Run("mixed expressions and static values", func(t *testing.T) {
		node := map[string]any{
			"static":  "plain",
			"dynamic": "${ .count }",
			"nested": map[string]any{
				"also_static":  100,
				"also_dynamic": "${ .name }",
			},
			"list": []any{
				"${ .first }",
				"middle",
				"${ .last }",
			},
		}
		input := map[string]any{
			"count": 42,
			"name":  "test",
			"first": "A",
			"last":  "Z",
		}

		result, err := traverseAndEvaluate(node, input, nil)
		require.NoError(t, err)

		resultMap := result.(map[string]any)
		assert.Equal(t, "plain", resultMap["static"])
		assert.Equal(t, 42, resultMap["dynamic"])

		nested := resultMap["nested"].(map[string]any)
		assert.Equal(t, 100, nested["also_static"])
		assert.Equal(t, "test", nested["also_dynamic"])

		list := resultMap["list"].([]any)
		assert.Equal(t, "A", list[0])
		assert.Equal(t, "middle", list[1])
		assert.Equal(t, "Z", list[2])
	})

	t.Run("expression with both input and variables", func(t *testing.T) {
		node := "${ .base + $multiplier }"
		input := map[string]any{"base": 10}
		variables := map[string]any{"$multiplier": 5}

		result, err := traverseAndEvaluate(node, input, variables)
		require.NoError(t, err)
		assert.Equal(t, 15, result)
	})
}
