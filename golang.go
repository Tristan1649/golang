package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ...

func main() {
	// ... (весь код до создания каналов и горутин)

	doneTasks := make(chan Ttype)
	undoneTasks := make(chan error)
	doneWg := sync.WaitGroup{}
	undoneWg := sync.WaitGroup{}
	resultMu := sync.Mutex{}
	result := make(map[int]Ttype)
	err := []error{}

	go func() {
		for r := range doneTasks {
			resultMu.Lock()
			result[r.id] = r
			resultMu.Unlock()
		}
		doneWg.Done()
	}()

	go func() {
		for r := range undoneTasks {
			undoneWg.Done()
			err = append(err, r)
		}
	}()

	go func() {
		// получение тасков
		for t := range superChan {
			// Используем doneWg и undoneWg, чтобы дождаться завершения обработки задач
			doneWg.Add(1)
			undoneWg.Add(1)

			go func(t Ttype) {
				defer func() {
					// Закрываем каналы только после завершения всех горутин
					doneWg.Wait()
					undoneWg.Wait()
					close(doneTasks)
					close(undoneTasks)
				}()

				t = task_worker(t)

				if string(t.taskRESULT) == "success" {
					doneTasks <- t
				} else {
					undoneTasks <- fmt.Errorf("Task id %d time %s, error %s", t.id, t.cT, t.taskRESULT)
				}
			}(t)
		}
	}()

	// ... (код после создания горутин)

	// Обработка сигнала завершения программы
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	println("Errors:")
	for _, r := range err {
		fmt.Println(r)
	}

	println("Done tasks:")
	resultMu.Lock()
	for id := range result {
		println(id)
	}
	resultMu.Unlock()
}
