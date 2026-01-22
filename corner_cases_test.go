package gopyre

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecCornerCases(t *testing.T) {
	t.Run("EmptyCode", func(t *testing.T) {
		_, err := Exec("", nil)
		require.Error(t, err)
	})

	t.Run("WhitespaceOnlyCode", func(t *testing.T) {
		_, err := Exec(" \n\t\n", nil)
		require.Error(t, err)
	})

	t.Run("TrailingBlankLines", func(t *testing.T) {
		result, err := Exec("x + y\n\n", map[string]any{"x": 1, "y": 2})
		require.NoError(t, err)
		require.Equal(t, 3.0, result)
	})

	t.Run("ExecThenEvalSplit", func(t *testing.T) {
		result, err := Exec("a = 2\nb = 3\na + b", nil)
		require.NoError(t, err)
		require.Equal(t, 5.0, result)
	})

	t.Run("NilInput", func(t *testing.T) {
		result, err := Exec("1 + 1", nil)
		require.NoError(t, err)
		require.Equal(t, 2.0, result)
	})

	t.Run("NonSerializableInput", func(t *testing.T) {
		_, err := Exec("x", map[string]any{"x": make(chan int)})
		require.Error(t, err)
	})

	t.Run("PythonException", func(t *testing.T) {
		_, err := Exec("1 / 0", nil)
		require.Error(t, err)
	})

	t.Run("ReturnNone", func(t *testing.T) {
		result, err := Exec("None", nil)
		require.NoError(t, err)
		require.Nil(t, result)
	})

	t.Run("ReturnNestedStructures", func(t *testing.T) {
		result, err := Exec(`{"a": [1, 2], "b": {"c": "d"}}`, nil)
		require.NoError(t, err)
		out, ok := result.(map[string]any)
		require.True(t, ok)
		require.Equal(t, []any{1.0, 2.0}, out["a"])
		require.Equal(t, map[string]any{"c": "d"}, out["b"])
	})

	t.Run("InputOverridesBuiltins", func(t *testing.T) {
		result, err := Exec("len", map[string]any{"len": 5})
		require.NoError(t, err)
		require.Equal(t, 5.0, result)
	})

	t.Run("InvalidIdentifierKey", func(t *testing.T) {
		result, err := Exec(`globals()["x-y"]`, map[string]any{"x-y": 3})
		require.NoError(t, err)
		require.Equal(t, 3.0, result)
	})

	t.Run("InputKeyNamedInput", func(t *testing.T) {
		result, err := Exec("input", map[string]any{"input": 9})
		require.NoError(t, err)
		require.Equal(t, 9.0, result)
	})

	t.Run("InputMutationDoesNotLeak", func(t *testing.T) {
		input := map[string]any{"x": 1}
		_, err := Exec(`input["x"] = 6`, input)
		t.Logf("Error %v", err)
		require.NoError(t, err)
		require.Equal(t, 1, input["x"])
	})

	t.Run("IsolationAcrossCalls", func(t *testing.T) {
		result, err := Exec("marker = 7\nmarker", nil)
		require.NoError(t, err)
		require.Equal(t, 7.0, result)

		result, err = Exec("globals().get(\"marker\")", nil)
		require.NoError(t, err)
		require.Nil(t, result)
	})
}
