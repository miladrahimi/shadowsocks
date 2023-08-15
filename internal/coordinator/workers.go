package coordinator

import "time"

func (c *Coordinator) startWorkers() {
	go c.run10SecondJobs()
	go c.runMinuteJobs()
	go c.start10SecondWorker()
	go c.startMinuteWorker()
}

func (c *Coordinator) start10SecondWorker() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		c.run10SecondJobs()
	}
}

func (c *Coordinator) run10SecondJobs() {
	go c.pullServers()
}

func (c *Coordinator) startMinuteWorker() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		go c.runMinuteJobs()
	}
}

func (c *Coordinator) runMinuteJobs() {
	go c.syncMetrics()
	go c.pushServers()
}
