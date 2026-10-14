package main

type fieldCase struct {
	path string
	typ  string
}

type layoutCase struct {
	name        string
	declaration string
	fields      []fieldCase
}
