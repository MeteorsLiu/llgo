package main

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
)

type sample struct {
	Words    []uint64      `json:"words"`
	Memory   []byte        `json:"memory,omitempty"`
	TypeData []byte        `json:"type_data,omitempty"`
	Values   []valueMemory `json:"values,omitempty"`
}
type valueMemory struct {
	Slot uint64 `json:"slot"`
	Data []byte `json:"data"`
	Text []byte `json:"text,omitempty"`
}

func byteOrder(words []uint64) binary.ByteOrder {
	if words[18] == 1 {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

func inspect(adapter, executable, directory string, fixture *existingFixture, logPath string, reference []sample, layoutWords []uint64, arguments []string) (samples []sample, err error) {
	client, err := startDAP(adapter, logPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := client.close(); err == nil {
			err = closeErr
		}
	}()
	var capabilities struct {
		ReadMemory bool `json:"supportsReadMemoryRequest"`
	}
	if err = client.request("initialize", map[string]any{"adapterID": "lldb", "clientID": "layoutlint", "linesStartAt1": true, "columnsStartAt1": true, "pathFormat": "path", "supportsMemoryReferences": true}, &capabilities); err != nil {
		return nil, err
	}
	if !capabilities.ReadMemory {
		return nil, fmt.Errorf("LLDB adapter does not support readMemory")
	}
	launch, err := client.send("launch", map[string]any{"program": executable, "cwd": directory, "args": append([]string{}, arguments...), "stopOnEntry": false})
	if err != nil {
		return nil, err
	}
	if err = client.event("initialized", nil); err != nil {
		return nil, err
	}
	lines := make(map[int]bool)
	for _, file := range fixture.files {
		if file.checkpointLine == 0 {
			continue
		}
		var breakpoints struct {
			Breakpoints []struct {
				Verified bool   `json:"verified"`
				Message  string `json:"message"`
			} `json:"breakpoints"`
		}
		if err = client.request("setBreakpoints", map[string]any{"source": map[string]any{"path": filepath.Join(directory, filepath.Base(file.path))}, "breakpoints": []map[string]int{{"line": file.checkpointLine}}}, &breakpoints); err != nil {
			return nil, err
		}
		if len(breakpoints.Breakpoints) != 1 || !breakpoints.Breakpoints[0].Verified {
			return nil, fmt.Errorf("LLDB did not resolve %s:%d: %+v", file.path, file.checkpointLine, breakpoints.Breakpoints)
		}
		lines[file.checkpointLine] = true
	}
	if err = client.request("configurationDone", struct{}{}, nil); err != nil {
		return nil, err
	}
	if err = client.response(launch, nil); err != nil {
		return nil, err
	}
	order, word := byteOrder(layoutWords), int(layoutWords[17])
	for stage, want := range reference {
		var stopped struct {
			Reason   string `json:"reason"`
			ThreadID int    `json:"threadId"`
		}
		if err = client.event("stopped", &stopped); err != nil {
			return samples, err
		}
		if stopped.Reason != "breakpoint" {
			return samples, fmt.Errorf("observation %d stopped for %s", stage, stopped.Reason)
		}
		var stack struct {
			Frames []struct {
				ID   int `json:"id"`
				Line int `json:"line"`
			} `json:"stackFrames"`
		}
		if err = client.request("stackTrace", map[string]any{"threadId": stopped.ThreadID, "levels": 1}, &stack); err != nil {
			return samples, err
		}
		if len(stack.Frames) != 1 || !lines[stack.Frames[0].Line] {
			return samples, fmt.Errorf("unexpected snapshot frame: %+v", stack.Frames)
		}
		var evaluated struct {
			Result string `json:"result"`
		}
		if err = client.request("evaluate", map[string]any{"frameId": stack.Frames[0].ID, "expression": "(void*)data", "context": "watch"}, &evaluated); err != nil {
			return samples, err
		}
		address, parseErr := strconv.ParseUint(evaluated.Result, 0, 64)
		if parseErr != nil {
			return samples, fmt.Errorf("snapshot pointer %q: %w", evaluated.Result, parseErr)
		}
		header, readErr := client.readMemory(address, 24)
		if readErr != nil {
			return samples, readErr
		}
		id, count := order.Uint64(header), order.Uint64(header[16:])
		if id != want.Words[0] || count != uint64(len(want.Words)) {
			return samples, fmt.Errorf("snapshot %d: id/count=%d/%d, want %d/%d", stage, id, count, want.Words[0], len(want.Words))
		}
		data, readErr := client.readMemory(order.Uint64(header[8:]), int(count)*8)
		if readErr != nil {
			return samples, readErr
		}
		current := sample{Words: make([]uint64, int(count))}
		for i := range current.Words {
			current.Words[i] = order.Uint64(data[i*8:])
		}
		if id%2 == 0 {
			stride := current.Words[2]
			if stride != 0 {
				current.Memory, err = client.readMemory(current.Words[1], int(stride*elementCount))
				if err != nil {
					return samples, err
				}
			}
			current.TypeData, err = client.readMemory(current.Words[5], 2*word+8)
			if err != nil {
				return samples, err
			}
		} else {
			for j := 1; j < len(current.Words); {
				words := current.Words
				if j+5 > len(words) || words[j+3] > uint64(len(words)-j-4) {
					return samples, fmt.Errorf("invalid LLGo value observation %d", stage)
				}
				kind := reflect.Kind(words[j+4])
				size := 0
				switch kind {
				case reflect.Bool, reflect.Int8, reflect.Uint8:
					size = 1
				case reflect.Int16, reflect.Uint16:
					size = 2
				case reflect.Int32, reflect.Uint32, reflect.Float32:
					size = 4
				case reflect.Int64, reflect.Uint64, reflect.Float64, reflect.Complex64:
					size = 8
				case reflect.Complex128:
					size = 16
				case reflect.Int, reflect.Uint, reflect.Uintptr, reflect.Pointer, reflect.UnsafePointer:
					size = word
				case reflect.Func, reflect.String:
					size = 2 * word
				case reflect.Slice:
					size = 3 * word
				}
				if size != 0 {
					raw := valueMemory{Slot: words[j]}
					raw.Data, err = client.readMemory(words[j+1], size)
					if err != nil {
						return samples, err
					}
					if kind == reflect.String {
						length := readInteger(order, raw.Data[word:])
						// Read only the length accepted by the Go observation. A bad header
						// remains a comparison failure without following an arbitrary extent.
						if length == want.Words[j+5] && length != 0 {
							raw.Text, err = client.readMemory(readInteger(order, raw.Data[:word]), int(length))
							if err != nil {
								return samples, err
							}
						}
					}
					current.Values = append(current.Values, raw)
				}
				j += 4 + int(words[j+3])
			}
		}
		samples = append(samples, current)
		if err = client.request("continue", map[string]any{"threadId": stopped.ThreadID}, nil); err != nil {
			return samples, err
		}
	}
	var exited struct {
		ExitCode int `json:"exitCode"`
	}
	if err = client.event("exited", &exited); err != nil {
		return samples, err
	}
	if exited.ExitCode != 0 {
		return samples, fmt.Errorf("debuggee exit code %d", exited.ExitCode)
	}
	return samples, nil
}
