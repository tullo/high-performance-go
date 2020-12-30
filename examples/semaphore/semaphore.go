package p

type Work int

// Semaphore to limit the amount of concurrent blocking IO
var semaphore = make(chan struct{}, 10)

func processRequest(work *Work) {
	semaphore <- struct{}{} // acquire semaphore
	// process request
	<-semaphore // release semaphore
}
