package main

import "go/types"

type sourceLocation struct {
	file string
	line int
}

type fieldCase struct {
	path string
	typ  string
	sourceLocation
}

type layoutCase struct {
	name        string
	displayName string
	declaration string
	fields      []fieldCase
	typ         types.Type
	irPrefix    string
	sourceLocation
}
