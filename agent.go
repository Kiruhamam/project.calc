package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type AgentTask struct {
	ID         int
	Expression string
	Result     float64
}

type Agent struct {
	agentName string
	serverURL string
	client    *http.Client
}

func NewAgent(agentName, serverURL string) *Agent {
	return &Agent{
		agentName: agentName,
		serverURL: serverURL,
		client:    &http.Client{},
	}
}

func (a *Agent) Run() {
	for {
		task, err := a.getTask()
		if err != nil {
			log.Printf("Ошибка при получении задачи: %v", err)
			continue
		}

		result, err := a.compute(task.Expression)
		if err != nil {
			log.Printf("Ошибка при вычислении выражения: %v", err)
			continue
		}

		task.Result = result

		if err := a.sendResult(task); err != nil {
			log.Printf("Ошибка при отправке результата: %v", err)
			continue
		}

		log.Printf("Задача %d успешно выполнена. Результат: %f", task.ID, result)
	}
}

func (a *Agent) getTask() (*AgentTask, error) {
	resp, err := a.client.Get(a.serverURL + "/gettask?agentname=" + a.agentName)
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP-запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неправильный статус ответа от сервера: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}

	var task AgentTask
	if err := json.Unmarshal(body, &task); err != nil {
		return nil, fmt.Errorf("ошибка разбора JSON: %v", err)
	}

	return &task, nil
}

func (a *Agent) compute(expression string) (float64, error) {
	operands := strings.FieldsFunc(expression, func(r rune) bool {
		return r == '+' || r == '-' || r == '*' || r == '/'
	})
	operators := strings.FieldsFunc(expression, func(r rune) bool {
		return r == ' ' || (r != '+' && r != '-' && r != '*' && r != '/')
	})

	var nums []float64
	for _, operand := range operands {
		num, err := strconv.ParseFloat(operand, 64)
		if err != nil {
			return 0, fmt.Errorf("ошибка преобразования операнда в число: %v", err)
		}
		nums = append(nums, num)
	}

	for i := 0; i < len(operators); i++ {
		switch operators[i] {
		case "+":
			nums[i+1] = nums[i] + nums[i+1]
		case "-":
			nums[i+1] = nums[i] - nums[i+1]
		case "*":
			nums[i+1] = nums[i] * nums[i+1]
		case "/":
			if nums[i+1] == 0 {
				return 0, fmt.Errorf("деление на ноль")
			}
			nums[i+1] = nums[i] / nums[i+1]
		}
	}

	return nums[len(nums)-1], nil
}

func (a *Agent) sendResult(task *AgentTask) error {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %v", err)
	}

	resp, err := a.client.Post(a.serverURL+"/sendresult", "application/json", bytes.NewBuffer(taskJSON))
	if err != nil {
		return fmt.Errorf("ошибка HTTP-запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("неправильный статус ответа от сервера: %s", resp.Status)
	}

	return nil
}

func Agent_start() {
	agent := NewAgent("demon", "http://localhost:8080")
	agent.Run()
}
func main() {
	// Запуск агента
	Agent_start()
}
