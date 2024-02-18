package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Task struct {
	ID         int
	Expression string
	Status     int     // 1 - в обработке, 2 - отдана вычислителю, 3 - завершена
	Result     float64 // Результат вычисления
}

type Orchestrator struct {
	tasks     map[int]*Task
	taskIDSeq int
	agents    map[string]bool // Монитор вычислителей
	mutex     sync.Mutex
}

func (o *Orchestrator) RegisterAgent(agentName string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.agents[agentName] = true
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		tasks:  make(map[int]*Task),
		agents: make(map[string]bool),
	}
}

func (o *Orchestrator) AddTask(expression string) int {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	taskID := o.taskIDSeq
	o.taskIDSeq++

	o.tasks[taskID] = &Task{
		ID:         taskID,
		Expression: expression,
		Status:     1, // В обработке
	}

	return taskID
}

func (o *Orchestrator) GetTask(agentName string) *Task {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Регистрация вычислителя, если его нет в мониторе
	if _, ok := o.agents[agentName]; !ok {
		o.agents[agentName] = true
	}

	// Поиск задачи со статусом "в обработке" для вычислителя
	for _, task := range o.tasks {
		if task.Status == 1 {
			task.Status = 2 // Отдана вычислителю
			return task
		}
	}

	return nil // Нет задач для выполнения
}

func (o *Orchestrator) SendResult(task *Task) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if _, ok := o.tasks[task.ID]; ok {
		o.tasks[task.ID].Result = task.Result
		o.tasks[task.ID].Status = 3 // Завершена
	}
}
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "http://localhost:63343") // Замените на адрес вашего фронтенда
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	orchestrator := NewOrchestrator()

	http.HandleFunc("/startserver", func(w http.ResponseWriter, r *http.Request) {
		// Здесь вы можете выполнить любые дополнительные действия при запуске сервера, если нужно
		fmt.Fprintf(w, "Server started")
	})

	http.HandleFunc("/addtask", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		expression := r.URL.Query().Get("expression")
		taskID := orchestrator.AddTask(expression)
		fmt.Fprintf(w, "%d", taskID)
	})

	http.HandleFunc("/gettask", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		agentName := r.URL.Query().Get("agentname")
		task := orchestrator.GetTask(agentName)
		if task != nil {
			json.NewEncoder(w).Encode(task)
		} else {
			fmt.Fprintf(w, "No task available for agent: %s\n", agentName)
		}
	})

	http.HandleFunc("/registertask", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		agentName := r.URL.Query().Get("agentname")
		orchestrator.RegisterAgent(agentName)
		fmt.Fprintf(w, "Agent %s registered successfully\n", agentName)
	})

	http.HandleFunc("/sendresult", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		var task Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, "Failed to decode task result", http.StatusBadRequest)
			return
		}
		orchestrator.SendResult(&task)
		fmt.Fprintf(w, "Result for task %d received successfully\n", task.ID)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
