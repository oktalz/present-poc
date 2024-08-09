package main

type Grocery struct {
	Name string
	Type string
}

type ShoopingList struct {
	Groceries map[string]Grocery
	// ...
}
