package worker

type Dispatcher struct {
	// The number of the worker.
	WorkerNum int

	// A buffered channel that we can send work requests on.
	JobQueue chan interface{}

	// The job handler, either which implements the interface Task,
	// or whose type is func(interface{}). This is mainly passed on to Worker.
	Handler interface{}

	// A pool of workers channels that are registered with the dispatcher
	workerPool chan chan interface{}

	workers []*worker
	quit    chan bool
	running bool
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(maxWorkers int, jobQueue chan interface{}, handler interface{}) *Dispatcher {
	return &Dispatcher{
		WorkerNum:  maxWorkers,
		JobQueue:   jobQueue,
		Handler:    handler,
		workerPool: make(chan chan interface{}, maxWorkers),
		quit:       make(chan bool),
	}
}

// Start the dispatcher and let the workers to handle the job.
func (d *Dispatcher) Run() {
	d.workers = make([]*worker, d.WorkerNum)

	// starting n number of workers
	for i := 0; i < d.WorkerNum; i++ {
		worker := newWorker(d.workerPool, d.Handler)
		d.workers[i] = worker
		worker.Start()
	}

	go d.dispatch()
	d.running = true
}

// Stop the dispatcher and all the workers.
func (d *Dispatcher) Stop() {
	d.quit <- true
	for _, worker := range d.workers {
		worker.Stop()
	}
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.JobQueue:
			// a job request has been received
			go func(job interface{}) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.workerPool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		case <-d.quit:
			d.running = false
			return
		}
	}
}