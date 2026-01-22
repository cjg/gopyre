package gopyre

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Exec executes inline Python code with a JSON-serializable input map.
// The last non-empty line is treated as an expression and returned as the result.
func Exec(code string, input map[string]any) (any, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errors.New("gopyre: empty code")
	}
	if input == nil {
		input = map[string]any{}
	}

	rt, err := getRuntime()
	if err != nil {
		return nil, err
	}

	state := rt.fns.PyGILState_Ensure()
	defer rt.fns.PyGILState_Release(state)

	globals, err := rt.newGlobals()
	if err != nil {
		return nil, err
	}
	defer rt.decRef(globals)

	inputObj, err := rt.jsonToPy(input)
	if err != nil {
		return nil, err
	}
	if err := rt.setDictItemString(globals, "input", inputObj); err != nil {
		rt.decRef(inputObj)
		return nil, err
	}
	rt.decRef(inputObj)

	execPart, evalPart := splitForEval(code)
	if execPart != "" {
		if _, err := rt.runString(execPart, pyFileInput, globals, globals); err != nil {
			return nil, err
		}
	}

	result, err := rt.runString(evalPart, pyEvalInput, globals, globals)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := rt.pyToJSON(result)
	if err != nil {
		return nil, err
	}

	var out any
	if err := json.Unmarshal(jsonBytes, &out); err != nil {
		return nil, fmt.Errorf("gopyre: decode result: %w", err)
	}

	return out, nil
}

func splitForEval(code string) (string, string) {
	lines := strings.Split(code, "\n")
	last := len(lines) - 1
	for last >= 0 && strings.TrimSpace(lines[last]) == "" {
		last--
	}
	if last <= 0 {
		return "", strings.TrimSpace(lines[0])
	}
	execPart := strings.Join(lines[:last], "\n")
	return execPart, strings.TrimSpace(lines[last])
}
