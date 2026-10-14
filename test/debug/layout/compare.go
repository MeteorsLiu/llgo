package main

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

type difference struct {
	Check string `json:"check"`
	Stage string `json:"stage,omitempty"`
	Item  string `json:"item"`
	Want  any    `json:"want"`
	Got   any    `json:"got"`
}

func compare(c layoutCase, reference, observed []sample) []difference {
	var differences []difference
	check := func(category, name string, want, got uint64) {
		if want != got {
			differences = append(differences, difference{category, "", c.name + "." + name, want, got})
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
	category := "access"
	if strings.Contains(site.label, "param") || strings.HasPrefix(site.label, "struct_values") {
		category = "call_return"
	}
	if strings.HasPrefix(site.label, "interface_") || strings.HasPrefix(site.label, "container_") || strings.HasPrefix(site.label, "function_") {
		category = "value_semantics"
	}
	check := func(name string, want, got uint64) {
		if want != got {
			differences = append(differences, difference{category, site.label, name, want, got})
		}
	}
	order, word := byteOrder(layoutWords), int(layoutWords[17])
	for j := 1; j < len(reference.Words); {
		want, got := reference.Words, observed.Words
		slot, n := want[j], int(want[j+3])
		name := site.names[slot]
		check(name+" slot", slot, got[j])
		check(name+" size", want[j+2], got[j+2])
		if got[j+3] != uint64(n) {
			check(name+" value length", uint64(n), got[j+3])
			return differences
		}
		payload := want[j+4 : j+4+n]
		if !slices.Equal(payload, got[j+4:j+4+n]) {
			differences = append(differences, difference{category, site.label, name + " value", payload, got[j+4 : j+4+n]})
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
					differences = append(differences, difference{category, site.label, name + " raw text", string(wantText), string(raw.Text)})
				}
			case reflect.Slice:
				check(name+" raw length", payload[1], readInteger(order, raw.Data[word:word*2]))
				check(name+" raw capacity", payload[2], readInteger(order, raw.Data[word*2:]))
			}
		}
		j += 4 + n
	}
	return differences
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
