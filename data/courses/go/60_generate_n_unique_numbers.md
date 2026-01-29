---
name: Генерация N уникальных чисел
module: Скрининг
tags:
  - грейды
  - озон
---
Требуется реализовать функцию uniqRandn, которая генерирует слайс длины n уникальных, рандомных чисел.

```go
package main

import (
	"fmt"
)

func main() {
	fmt.Println(uniqRandn(10))
}
func uniqRandn(n int) []int {
	return nil
}
```

Ответ:

```go
package main

import "math/rand"

func uniqRandn(n int) []int {
	res, resMap := make([]int, 0, n), make(map[int]struct{}, n)
	for len(res) < n {
		val := rand.Int()
		if _, ok := resMap[val]; ok {
			continue
		}
		res = append(res, val)
		resMap[val] = struct{}{}
	}
	return res
}
```