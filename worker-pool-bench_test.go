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
		sortedCh := make(chan []int, 1000)
		taskDivide := func(in []int) error {
			divideEx(in, sortedCh)
			return nil
		}
		taskMerge := func(left []int, right []int) error {
			mergeEx(left, right, sortedCh)
			return nil
		}
		taskDispatch := func(pool *wp.WorkerPool) error {
			var left, right []int
			newTuple := true
			for s := range sortedCh {
				if len(s) == len(in) {
					resultCh <- s
					return nil
				}
				if newTuple {
					left = s
					newTuple = false
					continue
				}
				right = s
				leftC := make([]int, len(left), len(left))
				rightC := make([]int, len(right), len(right))
				_ = copy(leftC, left)
				_ = copy(rightC, right)
				pool.Add(func() error {
					return taskMerge(leftC, rightC)
				})
				newTuple = true
			}
			return nil
		}
		options := wp.DefaultOptions
		pool := wp.NewWorkerPool(options)
		pool.Add(func() error {
			return taskDivide(in)
		})
		pool.Add(func() error {
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

func divideEx(in []int, outCh chan []int) {
	if len(in) == 1 {
		outCh <- in
		return
	}
	half := len(in) / 2
	left := in[:half]
	right := in[half:]
	divideEx(left, outCh)
	divideEx(right, outCh)
}

func mergeEx(left []int, right []int, outCh chan<- []int) {
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
	outCh <- out
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
