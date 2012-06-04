package main

import "sync"

var (
	schedule_test     = make(chan *Test)
	test_complete     = make(chan string, 1) //needs to buffer to avoid deadlock on the active test mutex
	active_tests      = make(map[string]*Test)
	active_tests_lock sync.RWMutex
)

func run_test_scheduler() {
	for {
		select {
		case t := <-schedule_test:
			id := t.WholeID()

			active_tests_lock.Lock()
			active_tests[id] = t
			active_tests_lock.Unlock()

			schedule_run <- id
		case id := <-test_complete:
			<-run_buffer

			active_tests_lock.Lock()
			t, ok := active_tests[id]
			if ok {
				delete(active_tests, id)
				save_item <- t
			}

			active_tests_lock.Unlock()
		}
	}
}
