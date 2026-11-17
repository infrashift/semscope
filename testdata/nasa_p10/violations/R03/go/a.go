package a

var table = make([]int, 8)

func NewBuffer(n int) []int {
	return make([]int, n)
}

func grow(xs []int) []int {
	return append(xs, 1)
}
