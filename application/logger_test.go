package application

import (
	"fmt"
	"testing"
)

func TestLogger(t *testing.T) {
	l, err := openLogStorage("qq")
	fmt.Println(l, err)

	c, err := l.chain("qq")
	fmt.Println(c, err)

	q := []int{1, 2, 3, 4, 5, 6}

	fmt.Println(q[2:6])
}
