package hw06pipelineexecution

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})
	t.Run("empty stages case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]int, 0)
		for v := range ExecutePipeline(in, nil) {
			result = append(result, v.(int))
		}

		require.Equal(t, data, result)
	})
}

func TestPipelineWithSlowStages(t *testing.T) {
	stages := []Stage{
		func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(10 * time.Millisecond)
					out <- v
				}
			}()
			return out
		},
		func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					out <- v.(int) * 2
				}
			}()
			return out
		},
	}

	t.Run("slow stage case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		start := time.Now()
		result := make([]int, 0)
		for v := range ExecutePipeline(in, nil, stages...) {
			result = append(result, v.(int))
		}
		elapsed := time.Since(start)

		require.Equal(t, []int{2, 4, 6}, result)
		require.GreaterOrEqual(t, elapsed, 30*time.Millisecond)
	})
}

func TestPipelineWithStringStages(t *testing.T) {
	stages := []Stage{
		func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					out <- "prefix_" + v.(string)
				}
			}()
			return out
		},
		func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					out <- v.(string) + "_suffix"
				}
			}()
			return out
		},
	}

	t.Run("string processing case", func(t *testing.T) {
		in := make(Bi)
		data := []string{"a", "b", "c"}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0)
		for v := range ExecutePipeline(in, nil, stages...) {
			result = append(result, v.(string))
		}

		require.Equal(t, []string{"prefix_a_suffix", "prefix_b_suffix", "prefix_c_suffix"}, result)
	})
}

func TestPipelineWithDoneSignal(t *testing.T) {
	stages := []Stage{
		func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					// Simulate some work
					time.Sleep(5 * time.Millisecond)
					out <- v.(int) * 2
				}
			}()
			return out
		},
		func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					out <- v.(int) + 1
				}
			}()
			return out
		},
	}

	t.Run("done in the middle case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)

		go func() {
			for i := 0; i < 100; i++ {
				select {
				case <-done:
					return
				case in <- i:
				}
			}
			close(in)
		}()

		// Send done signal after some time
		go func() {
			time.Sleep(15 * time.Millisecond)
			close(done)
		}()

		result := make([]int, 0)
		for v := range ExecutePipeline(in, done, stages...) {
			result = append(result, v.(int))
		}

		// Should have some results but not all 100
		require.NotEmpty(t, result)
		require.Less(t, len(result), 100)
	})
}

func TestPipelineWithNilValues(t *testing.T) {
	stages := []Stage{
		func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					out <- v
				}
			}()
			return out
		},
	}

	t.Run("nil values case", func(t *testing.T) {
		in := make(Bi)
		data := []interface{}{1, nil, "test", nil, 42}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]interface{}, 0)
		for v := range ExecutePipeline(in, nil, stages...) {
			result = append(result, v)
		}

		require.Equal(t, data, result)
	})
}
