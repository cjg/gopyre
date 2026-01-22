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
