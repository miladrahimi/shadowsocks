package coordinator

import "time"

func (c *Coordinator) startWorkers() {
	go c.runJobs()
	go c.startWorker()
}

func (c *Coordinator) startWorker() {
	ticker := time.NewTicker(time.Duration(c.Config.Worker.Interval) * time.Second)
	for range ticker.C {
		go c.runJobs()
	}
}

func (c *Coordinator) runJobs() {
	go c.pullServers()
	go c.syncMetrics()
	go c.pushServers()
}
