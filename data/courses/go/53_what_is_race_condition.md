---
name: Race condition
module: Теория
tags:
  - грейды
---
Можно ли передать переменную в несколько горутин? Пример (с ошибкой) - что будет?

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	x := make(map[int]int, 1)
	go func() { x[1] = 2 }()
	go func() { x[1] = 7 }()
	go func() { x[1] = 10 }()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("x[1] =", x[1])
}
```

**18 грейд**

Можно, нужно учесть race condition (mutex) Что будет - будет панинка concurrent map writes (при условии GOMAXPROCS != 1). Решение:

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	x := make(map[int]int, 1)
	lock := sync.RWMutex{}
	go func() {
		lock.Lock()
		x[1] = 2
		lock.Unlock()
	}()
	go func() {
		lock.Lock()
		x[1] = 7
		lock.Unlock()
	}()
	go func() {
		lock.Lock()
		x[1] = 11
		lock.Unlock()
	}()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("x[1] =", x[1])
}
```

**19 грейд**

Может продемонстрировать решение без sync.Mutex/sync.RWMutex

**Дополнительные вопросы**

Q: Какие знаешь средства для предотвращения RC?
A: Например mutex (может свой вариант сделать)

Q: какие mutex бывают и чем отличаются для чего нужны?
A: Бывают полностью блокирующие или блокировка только на запись (RWMutex)
