package conc

import (
	"context"
	"math/rand"
	"testing"

	"conc/wp"
)

func BenchmarkWorkerPoolMergeSort(b *testing.B) {
	in := genIntSlice()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		resultCh := make(chan []int)
		sortedCh := make(chan Sorted, 1000)
		dividedCh := make(chan Divided, 1000)
		taskDivide := func(in []int) interface{} {
			divide(in, dividedCh, sortedCh)
			return nil
		}
		taskMerge := func(left []int, right []int) interface{} {
			merge(left, right, sortedCh)
			return nil
		}
		taskDispatch := func(pool *wp.WorkerPool) interface{} {
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
						left = s.result
						newTuple = false
						continue
					}
					right = s.result
					leftC := make([]int, len(left), len(left))
					rightC := make([]int, len(right), len(right))
					_ = copy(leftC, left)
					_ = copy(rightC, right)
					pool.Add(func() interface{} {
						return taskMerge(leftC, rightC)
					})
					newTuple = true
				case d := <-dividedCh:
					pool.Add(func() interface{} {
						return taskDivide(d.left)
					})
					pool.Add(func() interface{} {
						return taskDivide(d.right)
					})
				}
			}
		}
		options := wp.DefaultOptions
		options.Size = 1000
		pool := wp.NewWorkerPool(options)
		pool.Add(func() interface{} {
			return taskDivide(in)
		})
		pool.Add(func() interface{} {
			return taskDispatch(pool)
		})
		pool.Run(ctx)
		<-resultCh
		cancel()
		<-pool.Done()
	}
}

func BenchmarkMergeSort(b *testing.B) {
	in := genIntSlice()
	for i := 0; i < b.N; i++ {
		mSort(in)
	}
}

func genIntSlice() (out []int) {
	size := 1000
	out = make([]int, size, size)
	rand.Seed(1111)
	for i := 0; i < size; i++ {
		out[i] = rand.Intn(1000000)
	}
	return out
}

func mSort(slice []int) []int {

	if len(slice) < 2 {
		return slice
	}
	mid := (len(slice)) / 2
	return mrg(mSort(slice[:mid]), mSort(slice[mid:]))
}

func mrg(left, right []int) []int {

	size, i, j := len(left)+len(right), 0, 0
	slice := make([]int, size, size)

	for k := 0; k < size; k++ {
		if i > len(left)-1 && j <= len(right)-1 {
			slice[k] = right[j]
			j++
		} else if j > len(right)-1 && i <= len(left)-1 {
			slice[k] = left[i]
			i++
		} else if left[i] < right[j] {
			slice[k] = left[i]
			i++
		} else {
			slice[k] = right[j]
			j++
		}
	}
	return slice
}
