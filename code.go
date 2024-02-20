package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Введите математическое выражение (для выхода введите 'exit'): ")
		expression, _ := reader.ReadString('\n')
		expression = strings.TrimSpace(expression)

		if expression == "exit" {
			fmt.Println("Программа завершена.")
			break
		}

		result, err := compute(expression)
		if err != nil {
			fmt.Printf("Ошибка при вычислении выражения: %v\n", err)
		} else {
			fmt.Printf("Результат вычисления: %f\n", result)
		}
	}
}

func compute(expression string) (float64, error) {
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
