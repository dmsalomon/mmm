package main

func contains(c []string, e string) bool {
	for _, v := range c {
		if e == v {
			return true
		}
	}
	return false
}
