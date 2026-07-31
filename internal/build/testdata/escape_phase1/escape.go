package escapephase1

func read(p *int) int {
	return *p
}

func Local() int {
	p := new(int)
	*p = 7
	return read(p)
}
