---
name: Функция zip
module: Скрининг
tags:
  - грейды
---
Требуется реализовать функцию zip, которая соединяет элементы двух слайсов в слайс пар

```go
package main

import "fmt"

func main() {
	s1, s2 := []int{1, 2, 3}, []int{4, 5, 6, 7, 8}
	fmt.Println(zip(s1, s2)) // [[1 4] [2 5] [3 6]]
}

func zip(s1 []int, s2 []int) [][]int {
	return nil
}
```

Ответ

```go
package main

func zip(s1 []int, s2 []int) [][]int {
	minLen := len(s1)
	if len(s2) < minLen {
		minLen = len(s2)
	}
	res := make([][]int, 0, minLen)
	for i := 0; i < minLen; i++ {
		res = append(res, []int{s1[i], s2[i]})
	}
	return res
}
```

**Дополнительные вопросы**

Q: Реализовать версию функции zip которая сможет соединять произвольное количество слайсов
A: Ответ:

```go
package main

func zip(s ...[]int) [][]int {
	if len(s) == 0 {
		return [][]int{}
	}
	minLen := len(s[0])
	for i := 1; i < len(s); i++ {
		if len(s[i]) < minLen {
			minLen = len(s[i])
		}
	}
	res := make([][]int, 0, minLen)
	for i := 0; i < minLen; i++ {
		x := make([]int, 0, len(s))
		for k := 0; k < len(s); k++ {
			x = append(x, s[k][i])
		}
		res = append(res, x)
	}
	return res
}
```
