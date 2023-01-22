package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

// A task represents a meaninglessness of our life
type task struct {
	id         string
	cT         time.Time     // время создания
	fT         time.Duration // время выполнения
	taskRESULT []byte
}

func taskGenerator(ctx context.Context, wg *sync.WaitGroup, a chan task) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if len(a) < cap(a) {
				ft := time.Now()
				if ft.Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
					ft = time.Time{}
				}
				a <- task{cT: ft, id: uuid.New().String()} // передаем таск на выполнение
			} else {
				time.Sleep(10_000)
			}
		}
	}
}

func taskWorker(ctx context.Context, wg *sync.WaitGroup, tasks chan task, cb func(t task)) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case a := <-tasks:
			if a.cT.Unix() > 0 {
				a.taskRESULT = []byte("task has been success")
				a.fT = time.Now().Sub(a.cT)
			} else {
				a.taskRESULT = []byte("something went wrong")
			}
			go cb(a)
		}
	}
}

func main() {
	var wg = new(sync.WaitGroup)
	var ctx, cf = context.WithCancel(context.Background())
	var superChan = make(chan task, 6)

	var doneTasks = make([]task, 0, 0)
	var undoneTasks = make([]error, 0, 0)
	var mxTasks sync.Mutex

	taskSorter := func(t task) {
		defer mxTasks.Unlock()
		mxTasks.Lock()
		if string(t.taskRESULT[14:]) == "success" {
			doneTasks = append(doneTasks, t)
		} else {
			undoneTasks = append(undoneTasks, fmt.Errorf("task id %d time %s, error %s", t.id, t.cT, t.taskRESULT))
		}
	}

	wg.Add(1)
	go taskGenerator(ctx, wg, superChan)

	wg.Add(1)
	go taskWorker(ctx, wg, superChan, taskSorter)

	time.Sleep(time.Second * 3)
	cf()
	wg.Wait()

	println("Errors:", len(undoneTasks))

	println("Done tasks:", len(doneTasks))
}
