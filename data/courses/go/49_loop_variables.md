---
name: Итерационные переменные
module: Теория
tags:
  - грейды
  - озон
---

Что выведет данный код?

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	values := []int{1, 2, 3, 4, 5}
	for _, val := range values {
		go func() {
			fmt.Println(val)
		}()
	}
	time.Sleep(100 * time.Millisecond)
}
```

После ответа - как исправить?

**17 грейд**

Одинаковые адреса и одинаковые значения:

```
5
5
5
5
5
```

Исправить:

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	values := []int{1, 2, 3, 4, 5}
	for _, val := range values {
		val := val // Копируем переменную
		go func() {
			fmt.Println(val)
		}()
	}
	time.Sleep(100 * time.Millisecond)
}
```

или

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	values := []int{1, 2, 3, 4, 5}
	for _, val := range values {
		go func(val int) {
			fmt.Println(val)
		}(val)
	}
	time.Sleep(100 * time.Millisecond)
}
````

**18 грейд**

- Смог решить без использования time.Sleep.
- Знает, что начиная с go 1.22 были изменены правила определения области видимости для переменных в циклах.