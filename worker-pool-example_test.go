package conc

import (
	"context"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"conc/wp"
)

type Sorted struct {
	result []int
}

type Divided struct {
	left  []int
	right []int
}

// Merge sort by work-pool example
func TestWorkerPoolExample(t *testing.T) {
	var pool *wp.WorkerPool
	in := []int{45, 1, 23, 56, 11, 777, 22, -12, 233}
	resultCh := make(chan []int)
	sortedCh := make(chan Sorted)
	dividedCh := make(chan Divided)
	taskDivide := func(in []int) interface{} {
		fmt.Printf("--divider-- got to divide - %v\n", in)
		divide(in, dividedCh, sortedCh)
		return nil
	}
	taskMerge := func(left []int, right []int) interface{} {
		fmt.Printf("--merger-- got to merge - %v & %v\n", left, right)
		merge(left, right, sortedCh)
		return nil
	}
	taskDispatch := func() interface{} {
		var left, right []int
		newTuple := true
		for {
			select {
			case s := <-sortedCh:
				if len(s.result) == len(in) {
					resultCh <- s.result
					return nil
				}
				if newTuple {
					fmt.Printf("--dispatcher-- got to dispatch LEFT sorted result %v\n", s.result)
					left = s.result
					newTuple = false
					continue
				}
				right = s.result
				fmt.Printf("--dispatcher-- got to dispatch RIGHT sorted result %v\n", s.result)
				leftC := make([]int, len(left), len(left))
				rightC := make([]int, len(right), len(right))
				_ = copy(leftC, left)
				_ = copy(rightC, right)
				pool.Add(func() interface{} {
					return taskMerge(leftC, rightC)
				})
				newTuple = true
				printStats(pool)
			case d := <-dividedCh:
				pool.Add(func() interface{} {
					return taskDivide(d.left)
				})
				pool.Add(func() interface{} {
					return taskDivide(d.right)
				})
				printStats(pool)
			}
		}
	}

	options := wp.DefaultOptions
	if runtime.GOMAXPROCS(0) == 1 {
		options.Size = 2
	}
	pool = wp.NewWorkerPool(options)
	pool.Add(func() interface{} {
		return taskDivide(in)
	})
	pool.Add(func() interface{} {
		return taskDispatch()
	})
	pool.Run(context.Background())
	got := <-resultCh
	assert.ElementsMatch(t, []int{-12, 1, 11, 22, 23, 45, 56, 233, 777}, got)
}

func divide(in []int, outDCh chan<- Divided, outSCh chan<- Sorted) {
	if len(in) == 1 {
		outSCh <- Sorted{result: in}
		return
	}
	half := len(in) / 2
	left := in[:half]
	right := in[half:]
	outDCh <- Divided{
		left:  left,
		right: right,
	}
}

func merge(left []int, right []int, outCh chan<- Sorted) {
	size, i, j := len(left)+len(right), 0, 0
	out := make([]int, size, size)
	for k := 0; k < size; k++ {
		if i > len(left)-1 && j <= len(right)-1 {
			out[k] = right[j]
			j++
		} else if j > len(right)-1 && i <= len(left)-1 {
			out[k] = left[i]
			i++
		} else if left[i] < right[j] {
			out[k] = left[i]
			i++
		} else {
			out[k] = right[j]
			j++
		}
	}
	outCh <- Sorted{result: out}
}

func printStats(wp *wp.WorkerPool) wp.Stats {
	stats := wp.Stats()
	fmt.Printf("--stats-- workers count %d\n", stats.WorkersCount)
	fmt.Printf("--stats-- idle workers %d\n", stats.Idlers)
	fmt.Printf("--stats-- tasks in queue %d\n", stats.TasksInQueue)
	return stats
}
