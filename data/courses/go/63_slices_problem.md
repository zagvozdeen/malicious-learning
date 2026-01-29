---
name: Задачка про слайсы
module: Скрининг
tags:
  - грейды
---
Что выведет программа?

```go
package main

import "fmt"

func a() {
	x := []int{}
	x = append(x, 0)
	x = append(x, 1)
	x = append(x, 2)
	y := append(x, 3)
	z := append(x, 4)
	fmt.Println(y, z)
}

func main() {
	a()
}
```

Ответ

`[0 1 2 4] [0 1 2 4]`
