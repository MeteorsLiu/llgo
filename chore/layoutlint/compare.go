package main

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"slices"
)

type difference struct {
	Check   string `json:"check"`
	Stage   string `json:"stage,omitempty"`
	Item    string `json:"item"`
	Want    any    `json:"want"`
	Got     any    `json:"got"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Warning bool   `json:"warning,omitempty"`
}

func compare(c layoutCase, reference, observed []sample) []difference {
	var differences []difference
	location := c.sourceLocation
	check := func(category, name string, want, got uint64) {
		if want != got {
			differences = append(differences, difference{category, "", c.displayName + "." + name, want, got, location.file, location.line, false})
		}
	}
	want, got := reference[0].Words, observed[0].Words
	for _, item := range []struct {
		index int
		name  string
	}{{6, "size"}, {7, "PtrBytes"}, {8, "alignment"}, {9, "field alignment"}, {10, "kind"}, {11, "reflect size"}, {12, "reflect alignment"}, {13, "reflect field alignment"}} {
		check("metadata", item.name, want[item.index], got[item.index])
	}
	check("layout", "array stride", want[2], got[2])
	check("layout", "unsafe.Sizeof", want[3], got[3])
	check("layout", "unsafe.Alignof", want[4], got[4])
	check("layout", "object address modulo alignment", 0, got[1]%want[4])
	check("layout", "metadata size vs actual stride", got[2], got[6])
	for i, field := range c.fields {
		location = field.sourceLocation
		index := headerWords + i*fieldWords
		for j, name := range []string{"unsafe offset", "physical offset", "metadata offset", "size", "alignment", "kind"} {
			category := "metadata"
			if j < 2 {
				category = "layout"
			}
			check(category, field.path+" "+name, want[index+j], got[index+j])
		}
	}
	order, word := byteOrder(want), int(want[17])
	location = c.sourceLocation
	raw := observed[0].TypeData
	check("metadata", "raw descriptor size", want[6], readInteger(order, raw[:word]))
	check("metadata", "raw descriptor PtrBytes", want[7], readInteger(order, raw[word:word*2]))
	check("metadata", "raw descriptor alignment", want[8], uint64(raw[word*2+5]))
	check("metadata", "raw descriptor field alignment", want[9], uint64(raw[word*2+6]))
	check("metadata", "raw descriptor kind", want[10], uint64(raw[word*2+7]&31))
	return differences
}

func compareValues(site observationSite, reference, observed sample, layoutWords []uint64) []difference {
	var differences []difference
	category := site.check
	location := site.sourceLocation
	check := func(name string, want, got uint64) {
		if want != got {
			differences = append(differences, difference{category, site.label, name, want, got, location.file, location.line, false})
		}
	}
	order, word := byteOrder(layoutWords), int(layoutWords[17])
	for j := 1; j < len(reference.Words); {
		want, got := reference.Words, observed.Words
		slot, n := want[j], int(want[j+3])
		name := site.names[slot]
		check(name+" slot", slot, got[j])
		location = site.locations[slot]
		check(name+" size", want[j+2], got[j+2])
		location = site.sourceLocation
		if got[j+3] != uint64(n) {
			check(name+" value length", uint64(n), got[j+3])
			return differences
		}
		payload := want[j+4 : j+4+n]
		actual := got[j+4 : j+4+n]
		if !slices.Equal(payload, actual) {
			// Allocation size classes can give append different capacities.
			// Retain that evidence without masking length or element differences,
			// including when slices occur inside other observed values.
			comparable := slices.Clone(actual)
			var wantCaps, gotCaps []uint64
			for _, index := range valueSliceCapacities(payload) {
				if actual[index] < actual[index-1] {
					differences = append(differences, difference{category, site.label, name + " slice capacity below length", actual[index-1], actual[index], site.file, site.line, false})
				}
				wantCaps = append(wantCaps, payload[index])
				gotCaps = append(gotCaps, actual[index])
				comparable[index] = payload[index]
			}
			if !slices.Equal(wantCaps, gotCaps) {
				item := name + " slice capacities"
				var wantCapacity, gotCapacity any = wantCaps, gotCaps
				if reflect.Kind(payload[0]) == reflect.Slice && len(wantCaps) == 1 {
					item = name + " capacity"
					wantCapacity, gotCapacity = wantCaps[0], gotCaps[0]
				}
				differences = append(differences, difference{category, site.label, item, wantCapacity, gotCapacity, site.file, site.line, true})
			}
			if !slices.Equal(payload, comparable) {
				differences = append(differences, difference{category, site.label, name + " value", payload, actual, site.file, site.line, false})
			}
		}
		for _, raw := range observed.Values {
			if raw.Slot != slot {
				continue
			}
			kind := reflect.Kind(payload[0])
			switch kind {
			case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64:
				mask := ^uint64(0)
				if len(raw.Data) < 8 {
					mask = (uint64(1) << (8 * len(raw.Data))) - 1
				}
				check(name+" raw value", payload[1]&mask, readInteger(order, raw.Data))
			case reflect.Complex64, reflect.Complex128:
				half := len(raw.Data) / 2
				check(name+" raw real", payload[1], readInteger(order, raw.Data[:half]))
				check(name+" raw imag", payload[2], readInteger(order, raw.Data[half:]))
			case reflect.Pointer, reflect.UnsafePointer, reflect.Func:
				nonnil := uint64(0)
				if readInteger(order, raw.Data[:word]) != 0 {
					nonnil = 1
				}
				check(name+" raw nonnil", payload[1], nonnil)
			case reflect.String:
				check(name+" raw length", payload[1], readInteger(order, raw.Data[word:]))
				wantText := make([]byte, len(payload)-2)
				for i, value := range payload[2:] {
					wantText[i] = byte(value)
				}
				if !slices.Equal(wantText, raw.Text) {
					differences = append(differences, difference{category, site.label, name + " raw text", string(wantText), string(raw.Text), site.file, site.line, false})
				}
			case reflect.Slice:
				check(name+" raw length", payload[1], readInteger(order, raw.Data[word:word*2]))
				check(name+" raw capacity vs LLGo reflection", actual[2], readInteger(order, raw.Data[word*2:]))
			}
		}
		j += 4 + n
	}
	return differences
}

// valueSliceCapacities follows the Go reference payload emitted by layoutValue.
// Only capacity positions are classified differently; all other words retain
// their exact comparison, such as a nested slice's length and element values.
func valueSliceCapacities(words []uint64) []int {
	var capacities []int
	index := 0
	var visit func()
	visit = func() {
		kind := reflect.Kind(words[index])
		index++
		switch kind {
		case reflect.Invalid:
		case reflect.Array, reflect.Slice, reflect.Struct:
			count := words[index]
			index++
			if kind == reflect.Slice {
				capacities = append(capacities, index)
				index++
			}
			for range count {
				visit()
			}
		case reflect.Interface:
			visit()
		case reflect.String:
			index += 1 + int(words[index])
		case reflect.Complex64, reflect.Complex128, reflect.Chan:
			index += 2
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64, reflect.Pointer, reflect.UnsafePointer, reflect.Func, reflect.Map:
			index++
		default:
			panic(fmt.Sprintf("unsupported reference value kind %s", kind))
		}
	}
	visit()
	return capacities
}

func readInteger(order binary.ByteOrder, data []byte) uint64 {
	switch len(data) {
	case 1:
		return uint64(data[0])
	case 2:
		return uint64(order.Uint16(data))
	case 4:
		return uint64(order.Uint32(data))
	case 8:
		return order.Uint64(data)
	default:
		panic(fmt.Sprintf("unsupported scalar width %d", len(data)))
	}
}
