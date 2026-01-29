---
name: Задачка про слайсы
module: 
tags:
  - грейды
---
Что выведет программа? Почему?

```go
package main

import "time"

func main() {
	timeStart := time.Now()
	_, _ = <-worker(), <-worker()
	println(int(time.Since(timeStart).Seconds()))
}

func worker() chan int {
	ch := make(chan int)
	go func() {
		time.Sleep(3 * time.Second)
		ch <- 1
	}()
	return ch
}
```

Ответ: 6.