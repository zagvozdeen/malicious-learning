---
name: Написать код функции, которая делает merge N каналов.
module: Практика
tags:
  - грейды
---

Написать код функции, которая делает merge N каналов. Весь входной поток перенаправляется в один канал.

```go
package main

func merge(cs ...<-chan int) <-chan int {
	return nil
}
```


Это вопрос на написание кода, который потом будет приложен к протоколу собеседования. Пример реализации:

```go
package main

import "sync"

func merge(cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)
	// Start an output goroutine for each input channel in cs. output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan int) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}
	// Start a goroutine to close out once all the output goroutines are
	// done. This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
```
