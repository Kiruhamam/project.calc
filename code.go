package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Task struct {
	ID       int
	Expr     string
	Result   float64
	Finished bool
}

type Orchestrator struct {
	tasks     map[int]*Task
	taskQueue chan int
	mu        sync.Mutex
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		tasks:     make(map[int]*Task),
		taskQueue: make(chan int),
	}
}

func (o *Orchestrator) AddTask(expr string) int {
	o.mu.Lock()
	defer o.mu.Unlock()

	taskID := len(o.tasks) + 1
	task := &Task{
		ID:     taskID,
		Expr:   expr,
		Result: 0,
	}
	o.tasks[taskID] = task
	go func() {
		o.taskQueue <- taskID
	}()

	return taskID
}

func (o *Orchestrator) TaskStatus(taskID int) (float64, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()

	task, ok := o.tasks[taskID]
	if !ok {
		return 0, false
	}
	return task.Result, task.Finished
}

func (o *Orchestrator) Run() {
	for taskID := range o.taskQueue {
		o.mu.Lock()
		task := o.tasks[taskID]
		o.mu.Unlock()

		result, err := compute(task.Expr)
		if err == nil {
			o.mu.Lock()
			task.Result = result
			task.Finished = true
			o.mu.Unlock()
		}
	}
}

func compute(expression string) (float64, error) {
	operands := []float64{0}
	operator := '+'

	for _, char := range expression {
		switch char {
		case '+', '-', '*', '/':
			operator = char
		default:
			num, err := strconv.ParseFloat(string(char), 64)
			if err != nil {
				continue
			}

			switch operator {
			case '+':
				operands[len(operands)-1] += num
			case '-':
				operands[len(operands)-1] -= num
			case '*':
				operands[len(operands)-1] *= num
			case '/':
				if num != 0 {
					operands[len(operands)-1] /= num
				}
			}
		}
	}

	return operands[0], nil
}

func main() {
	orchestrator := NewOrchestrator()
	go orchestrator.Run()

	for {
		var input string
		fmt.Print("Введите математическое выражение (для выхода введите 'exit'): ")
		fmt.Scanln(&input)

		if input == "exit" {
			fmt.Println("Программа завершена.")
			break
		}

		taskID := orchestrator.AddTask(input)
		fmt.Printf("Выражение отправлено на обработку. ID задачи: %d\n", taskID)

		for {
			time.Sleep(1 * time.Second)
			result, finished := orchestrator.TaskStatus(taskID)
			if finished {
				fmt.Printf("Результат вычисления: %f\n", result)
				break
			} else {
				fmt.Println("Ожидание вычисления...")
			}
		}
	}
}
