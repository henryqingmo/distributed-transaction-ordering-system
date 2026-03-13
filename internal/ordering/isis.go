package ordering

type Queue struct {
	items []int
}

type isisOrdering struct {
	queue Queue
}

func (q *Queue) Enqueue(x int) {
	q.items = append(q.items, x)
}

func (q *Queue) Dequeue() int {
	x := q.items[0]
	q.items = q.items[1:]
	return x
}
