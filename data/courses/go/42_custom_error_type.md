---
name: Пользовательский тип ошибки
module: Практика
tags:
  - грейды
  - озон
---

Напишите функцию которая бы возвращала ошибку не импортируя для этого никаких пакетов:

```go
package main

func main() {
	println(handle())
}
func handle() error {
	return nil
}
```

Это вопрос на знание типа error и понимание основ работы с ошибками. error является интерфейсом, чтобы решить задачу требуется создать свою структуру реализующую этот интерфейс и вернуть из функции экземпляр структуры:

```go
package main

func main() {
	println(handle())
}

func handle() error {
	return &customError{}
}

type customError struct{}

func (e *customError) Error() string {
	return "Custom error!"
}
```

