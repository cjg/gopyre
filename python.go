package gopyre

import (
	"encoding/json"
	"fmt"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	pyFileInput = 257
	pyEvalInput = 258
)

type pyFns struct {
	Py_Initialize            func()
	Py_IsInitialized         func() int32
	PyEval_InitThreads       func()
	PyGILState_Ensure        func() uintptr
	PyGILState_Release       func(state uintptr)
	Py_NewInterpreter        func() unsafe.Pointer
	Py_EndInterpreter        func(tstate unsafe.Pointer)
	PyEval_GetBuiltins       func() unsafe.Pointer
	PyDict_New               func() unsafe.Pointer
	PyDict_SetItemString     func(dict unsafe.Pointer, key *byte, value unsafe.Pointer) int32
	PyImport_ImportModule    func(name *byte) unsafe.Pointer
	PyObject_GetAttrString   func(obj unsafe.Pointer, name *byte) unsafe.Pointer
	PyUnicode_FromString     func(s *byte) unsafe.Pointer
	PyUnicode_AsUTF8         func(obj unsafe.Pointer) *byte
	PyTuple_New              func(n int) unsafe.Pointer
	PyTuple_SetItem          func(tuple unsafe.Pointer, pos int, item unsafe.Pointer) int32
	PyObject_CallObject      func(callable unsafe.Pointer, args unsafe.Pointer) unsafe.Pointer
	PyRun_StringFlags        func(code *byte, start int, globals, locals, flags unsafe.Pointer) unsafe.Pointer
	PyErr_Fetch              func(ptype, pvalue, ptrace *unsafe.Pointer)
	PyErr_NormalizeException func(ptype, pvalue, ptrace *unsafe.Pointer)
	PyObject_Str             func(obj unsafe.Pointer) unsafe.Pointer
	Py_DecRef                func(obj unsafe.Pointer)

	PyThreadState_New            func(unsafe.Pointer) unsafe.Pointer
	PyThreadState_Swap           func(unsafe.Pointer) unsafe.Pointer
	PyThreadState_Clear          func(unsafe.Pointer)
	PyThreadState_GetInterpreter func(unsafe.Pointer) unsafe.Pointer
	PyThreadState_DeleteCurrent  func()
}

func (rt *pyRuntime) bind() {
	purego.RegisterLibFunc(&rt.fns.Py_Initialize, rt.lib, "Py_Initialize")
	purego.RegisterLibFunc(&rt.fns.Py_IsInitialized, rt.lib, "Py_IsInitialized")
	purego.RegisterLibFunc(&rt.fns.PyEval_InitThreads, rt.lib, "PyEval_InitThreads")
	purego.RegisterLibFunc(&rt.fns.PyGILState_Ensure, rt.lib, "PyGILState_Ensure")
	purego.RegisterLibFunc(&rt.fns.PyGILState_Release, rt.lib, "PyGILState_Release")
	purego.RegisterLibFunc(&rt.fns.Py_NewInterpreter, rt.lib, "Py_NewInterpreter")
	purego.RegisterLibFunc(&rt.fns.Py_EndInterpreter, rt.lib, "Py_EndInterpreter")
	purego.RegisterLibFunc(&rt.fns.PyEval_GetBuiltins, rt.lib, "PyEval_GetBuiltins")
	purego.RegisterLibFunc(&rt.fns.PyDict_New, rt.lib, "PyDict_New")
	purego.RegisterLibFunc(&rt.fns.PyDict_SetItemString, rt.lib, "PyDict_SetItemString")
	purego.RegisterLibFunc(&rt.fns.PyImport_ImportModule, rt.lib, "PyImport_ImportModule")
	purego.RegisterLibFunc(&rt.fns.PyObject_GetAttrString, rt.lib, "PyObject_GetAttrString")
	purego.RegisterLibFunc(&rt.fns.PyUnicode_FromString, rt.lib, "PyUnicode_FromString")
	purego.RegisterLibFunc(&rt.fns.PyUnicode_AsUTF8, rt.lib, "PyUnicode_AsUTF8")
	purego.RegisterLibFunc(&rt.fns.PyTuple_New, rt.lib, "PyTuple_New")
	purego.RegisterLibFunc(&rt.fns.PyTuple_SetItem, rt.lib, "PyTuple_SetItem")
	purego.RegisterLibFunc(&rt.fns.PyObject_CallObject, rt.lib, "PyObject_CallObject")
	purego.RegisterLibFunc(&rt.fns.PyRun_StringFlags, rt.lib, "PyRun_StringFlags")
	purego.RegisterLibFunc(&rt.fns.PyErr_Fetch, rt.lib, "PyErr_Fetch")
	purego.RegisterLibFunc(&rt.fns.PyErr_NormalizeException, rt.lib, "PyErr_NormalizeException")
	purego.RegisterLibFunc(&rt.fns.PyObject_Str, rt.lib, "PyObject_Str")
	purego.RegisterLibFunc(&rt.fns.Py_DecRef, rt.lib, "Py_DecRef")

	purego.RegisterLibFunc(&rt.fns.PyThreadState_New, rt.lib, "PyThreadState_New")
	purego.RegisterLibFunc(&rt.fns.PyThreadState_Swap, rt.lib, "PyThreadState_Swap")
	purego.RegisterLibFunc(&rt.fns.PyThreadState_Clear, rt.lib, "PyThreadState_Clear")
	purego.RegisterLibFunc(&rt.fns.PyThreadState_GetInterpreter, rt.lib, "PyThreadState_GetInterpreter")
	purego.RegisterLibFunc(&rt.fns.PyThreadState_DeleteCurrent, rt.lib, "PyThreadState_DeleteCurrent")
}

func (rt *pyRuntime) newGlobals() (unsafe.Pointer, error) {
	globals := rt.fns.PyDict_New()
	if globals == nil {
		return nil, rt.fetchError("create globals")
	}
	builtins := rt.fns.PyEval_GetBuiltins()
	if builtins != nil {
		if err := rt.setDictItemString(globals, "__builtins__", builtins); err != nil {
			rt.decRef(globals)
			return nil, err
		}
	}
	return globals, nil
}

func (rt *pyRuntime) jsonToPy(input map[string]any) (unsafe.Pointer, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("gopyre: encode input: %w", err)
	}

	jsonMod, err := rt.importModule("json")
	if err != nil {
		return nil, err
	}
	defer rt.decRef(jsonMod)

	loads, err := rt.getAttr(jsonMod, "loads")
	if err != nil {
		return nil, err
	}
	defer rt.decRef(loads)

	arg, err := rt.pyString(string(payload))
	if err != nil {
		return nil, err
	}

	args := rt.fns.PyTuple_New(1)
	if args == nil {
		rt.decRef(arg)
		return nil, rt.fetchError("create args")
	}
	if rt.fns.PyTuple_SetItem(args, 0, arg) != 0 {
		rt.decRef(args)
		rt.decRef(arg)
		return nil, rt.fetchError("set args")
	}

	obj := rt.fns.PyObject_CallObject(loads, args)
	rt.decRef(args)
	if obj == nil {
		return nil, rt.fetchError("json.loads")
	}
	return obj, nil
}

func (rt *pyRuntime) pyToJSON(obj unsafe.Pointer) ([]byte, error) {
	jsonMod, err := rt.importModule("json")
	if err != nil {
		rt.decRef(obj)
		return nil, err
	}
	defer rt.decRef(jsonMod)

	dumps, err := rt.getAttr(jsonMod, "dumps")
	if err != nil {
		rt.decRef(obj)
		return nil, err
	}
	defer rt.decRef(dumps)

	args := rt.fns.PyTuple_New(1)
	if args == nil {
		rt.decRef(obj)
		return nil, rt.fetchError("create args")
	}
	if rt.fns.PyTuple_SetItem(args, 0, obj) != 0 {
		rt.decRef(args)
		rt.decRef(obj)
		return nil, rt.fetchError("set args")
	}

	jsonObj := rt.fns.PyObject_CallObject(dumps, args)
	rt.decRef(args)
	if jsonObj == nil {
		return nil, rt.fetchError("json.dumps")
	}
	defer rt.decRef(jsonObj)

	jsonStr := rt.fns.PyUnicode_AsUTF8(jsonObj)
	if jsonStr == nil {
		return nil, rt.fetchError("decode result")
	}
	return []byte(cStringToGo(jsonStr)), nil
}

func (rt *pyRuntime) runString(code string, mode int, globals, locals unsafe.Pointer) (unsafe.Pointer, error) {
	raw, cstr := cString(code)
	result := rt.fns.PyRun_StringFlags(cstr, mode, globals, locals, nil)
	runtime.KeepAlive(raw)
	if result == nil {
		return nil, rt.fetchError("execute python")
	}
	return result, nil
}

func (rt *pyRuntime) importModule(name string) (unsafe.Pointer, error) {
	raw, cstr := cString(name)
	mod := rt.fns.PyImport_ImportModule(cstr)
	runtime.KeepAlive(raw)
	if mod == nil {
		return nil, rt.fetchError("import module " + name)
	}
	return mod, nil
}

func (rt *pyRuntime) getAttr(obj unsafe.Pointer, name string) (unsafe.Pointer, error) {
	raw, cstr := cString(name)
	attr := rt.fns.PyObject_GetAttrString(obj, cstr)
	runtime.KeepAlive(raw)
	if attr == nil {
		return nil, rt.fetchError("get attribute " + name)
	}
	return attr, nil
}

func (rt *pyRuntime) pyString(s string) (unsafe.Pointer, error) {
	raw, cstr := cString(s)
	obj := rt.fns.PyUnicode_FromString(cstr)
	runtime.KeepAlive(raw)
	if obj == nil {
		return nil, rt.fetchError("create string")
	}
	return obj, nil
}

func (rt *pyRuntime) setDictItemString(dict unsafe.Pointer, key string, value unsafe.Pointer) error {
	raw, cstr := cString(key)
	rc := rt.fns.PyDict_SetItemString(dict, cstr, value)
	runtime.KeepAlive(raw)
	if rc != 0 {
		return rt.fetchError("set dict " + key)
	}
	return nil
}

func (rt *pyRuntime) fetchError(context string) error {
	var ptype, pvalue, ptrace unsafe.Pointer
	rt.fns.PyErr_Fetch(&ptype, &pvalue, &ptrace)
	rt.fns.PyErr_NormalizeException(&ptype, &pvalue, &ptrace)

	msg := "python error"
	if pvalue != nil {
		strObj := rt.fns.PyObject_Str(pvalue)
		if strObj != nil {
			if cstr := rt.fns.PyUnicode_AsUTF8(strObj); cstr != nil {
				msg = cStringToGo(cstr)
			}
			rt.decRef(strObj)
		}
	}
	rt.decRef(ptype)
	rt.decRef(pvalue)
	rt.decRef(ptrace)

	return fmt.Errorf("gopyre: %s: %s", context, msg)
}

func (rt *pyRuntime) decRef(obj unsafe.Pointer) {
	if obj != nil {
		if rt.fns.Py_DecRef != nil {
			rt.fns.Py_DecRef(obj)
		}
	}
}

func cString(s string) ([]byte, *byte) {
	b := append([]byte(s), 0)
	return b, (*byte)(unsafe.Pointer(&b[0]))
}

func cStringToGo(ptr *byte) string {
	if ptr == nil {
		return ""
	}
	var n int
	for {
		b := *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + uintptr(n)))
		if b == 0 {
			break
		}
		n++
	}
	return string(unsafe.Slice(ptr, n))
}
