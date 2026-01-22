package gopyre

import (
	"testing"

	"github.com/stretchr/testify/require"
)


func TestOpenLibPython(t *testing.T) {
	_, err := openLibPython()
	require.NoError(t, err)
}

func TestBasicExec(t *testing.T) {
	result, err := Exec(`"1"`, nil)
	require.NoError(t, err)
	resultString, ok := result.(string)
	require.True(t, ok)
	require.Equal(t, "1", resultString)
}

func TestArgumentsPassing(t *testing.T) {
	result, err := Exec(`x + y`, map[string]any{"x": 1, "y": 2})
	require.NoError(t, err)
	resultFloat, ok := result.(float64)
	require.True(t, ok)
	require.Equal(t, 3.0, resultFloat)
}
