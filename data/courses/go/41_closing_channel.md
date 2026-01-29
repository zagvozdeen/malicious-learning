---
name: Закрытие канала
module: Практика
tags:
  - грейды
---
Что выведет данная программа? Она отработает корректно?

```go
package main

import "fmt"

func main() {
	ch := make(chan int)
    
	go func() {
		for i := 0; i < 5; i++ {
			ch <- i
		}
	}()

	for n := range ch {
		fmt.Println(n)
	}
}
```

В программе содержится ошибка, она сначала выведет числа от 0 до 4, а потом упадет с сообщением "all goroutines are asleep - deadlock!". Чтобы починить программу надо в конце горутины закрывать канал:

```go
package main

import "fmt"

func main() {
	ch := make(chan int)
	go func() {
		for i := 0; i < 5; i++ {
			ch <- i
		}
		close(ch) // <<<
	}()
	for n := range ch {
		fmt.Println(n)
	}
}
```