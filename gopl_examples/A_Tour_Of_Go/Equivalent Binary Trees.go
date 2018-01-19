package main

import "golang.org/x/tour/tree"

// Walk walks the tree t sending all values
// from the tree to the channel ch.
func Walk(t *tree.Tree, ch chan int) {
	if t == nil {
		return
	}
	Walk(t.Left, ch)
	ch <- t.Value
	Walk(t.Right, ch)
}

// Same determines whether the trees
// t1 and t2 contain the same values.
func Same(t1, t2 *tree.Tree) bool {
	ch1 := make(chan int)
	ch2 := make(chan int)
	go func() {
		Walk(t1, ch1)
		ch1 <- 0
	}()

	go func() {
		Walk(t2, ch2)
		ch2 <- 0
	}()

	for {
		t1 := <-ch1
		t2 := <-ch2
		if t1 == 0 && t2 == 0 {
			return true
		}
		if t1 == t2 {
			continue
		} else {
			return false
		}
	}
}

func main() {
	ch := make(chan int)
	go func() {
		Walk(tree.New(1), ch)
		ch <- 0
	}()

	for {
		t := <-ch
		if t == 0 {
			break
		}
		println(t)
	}
	println(Same(tree.New(1), tree.New(2)))
}
