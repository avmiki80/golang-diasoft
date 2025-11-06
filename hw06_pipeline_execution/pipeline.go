package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	current := in
	for _, stage := range stages {
		current = worker(current, done, stage)
	}
	return current
}

func worker(in In, done In, stage Stage) Out {
	out := make(Bi)

	go func() {
		defer close(out)

		stageOut := stage(in)

		for {
			select {
			case <-done:
				return
			case v, ok := <-stageOut:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				case out <- v:
				}
			}
		}
	}()

	return out
}
