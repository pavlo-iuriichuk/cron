package internal

import (
	"context"
	"time"
)

type Executor interface {
	Run(ctx context.Context, now time.Time)
	IsDone() bool
	Id() uint
}

type Planner interface {
	Interval(now time.Time) time.Duration
}

type PlannerAndInt interface {
	Planner
	Int
}

type Job func(ctx context.Context)

type Task struct {
	id        uint
	priority  int
	fn        Job
	execCount int
	count     int
	every     bool
	at        time.Time
	interval  time.Duration
}

func (t *Task) AsInt() int {
	now := time.Now()
	return t.priority * int(now.Sub(t.at).Milliseconds())
}

func (t *Task) Interval(now time.Time) time.Duration {
	return time.Duration(t.at.Sub(now).Milliseconds())
}

func (t *Task) Run(ctx context.Context, at time.Time) {
	go t.fn(ctx)
	t.execCount++
	if t.every {
		t.at = at.Add(t.interval)
	}
}

func (t *Task) IsDone() bool {
	return (!t.every && t.execCount > 0) || (t.count != -1) || (t.execCount == t.count)
}

func (t *Task) Id() uint {
	return t.id
}

type Scheduler interface {
	Init()
	At(j Job, when time.Time) error
	Every(j Job, interval time.Duration, count int) error
	Done()
}

type HeapScheduler struct {
	ctx         context.Context
	idAllocator Allocator
	done        chan bool
	t           *time.Timer
	tasks       *Heap
}

func NewHeapScheduler(ctx context.Context, maxTasks int) *HeapScheduler {
	return &HeapScheduler{
		idAllocator: NewHAllocator(uint(maxTasks)),
		tasks:       NewHeap(maxTasks),
		done:        make(chan bool, 1),
		ctx:         ctx,
	}
}

func (s *HeapScheduler) Init() {
	s.idAllocator.Init()
	s.t = &time.Timer{}
	go func() {
		for {
			select {
			case t := <-s.t.C:
				s.run(t)
			case <-s.ctx.Done():
			case <-s.done:
				s.t.Stop()
				return
			}
		}
	}()
}

func (s *HeapScheduler) At(j Job, when time.Time) error {
	id, err := s.idAllocator.Alloc()
	if err != nil {
		return err
	}
	task := &Task{
		id:       id,
		priority: 1,
		fn:       j,
		at:       when,
		count:    1,
	}
	err = s.tasks.Push(task)
	if err != nil {
		return err
	}
	s.scheduleNext(task, when)
	return nil
}

func (s *HeapScheduler) Every(j Job, interval time.Duration, count int) error {
	id, err := s.idAllocator.Alloc()
	if err != nil {
		return err
	}
	when := time.Now().Add(interval)
	task := &Task{
		id:       id,
		priority: 1,
		fn:       j,
		at:       when,
		count:    count,
		interval: interval,
		every:    true,
	}
	err = s.tasks.Push(task)
	if err != nil {
		return err
	}
	s.scheduleNext(task, when)
	return nil
}

func (s *HeapScheduler) run(t time.Time) {
	task, err := s.tasks.Pop()
	if err != nil {
		return
	}
	runTask, _ := (task).(Executor)
	runTask.Run(s.ctx, t)
	if runTask.IsDone() {
		err = s.idAllocator.Free(runTask.Id())
	}
	task, err = s.tasks.Top()
	if err != nil {
		return
	}
	nextTask, _ := (task).(PlannerAndInt)
	s.scheduleNext(nextTask, t)
}

func (s *HeapScheduler) scheduleNext(task PlannerAndInt, now time.Time) {
	s.t.Reset(task.Interval(now))
	err := s.tasks.Push(task)
	if err != nil {
		return
	}
}

func (s *HeapScheduler) Done() {
	s.done <- true
}
