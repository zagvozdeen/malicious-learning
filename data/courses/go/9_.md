---
name: Ревью кода кеша
module: Теория
tags:
  - грейды
  - код-ревью
---
Разработчик дал на ревью код своего нового кэша, нам необходимо провести код-ревью.

Кэш будет использоваться под высокой нагрузкой в проде. Частота записи/чтения 20%/80% соответственно.

Код:

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	fmt.Println(GetOrCreate("hello", "world"))
	fmt.Println(Get("hello"))
}

var cache = make(map[string]string)

// GetOrCreate проверяет существование ключа key.
// Если такого нет, то создает новое значение.
func GetOrCreate(key, value string) string {
	var m sync.Mutex

	m.Lock()
	value = cache[key]
	m.Unlock()

	if value != "" {
		return value
	}

	m.Lock()
	cache[key] = value
	m.Unlock()
	return value
}

func Get(key string) string {
	var m sync.Mutex
	m.Lock()
	v := cache[key]
	m.Unlock()
	return v
}
```

**17 грейд**

- Заметил что мьютекс — локальная переменная
- Дал рекомендацию оформить это в структуру

**18 грейд**

- Посоветовал использовать defer для мьютексов
- Предложил переписать GetOrCreate чтобы лочить мьютекс только один раз

**19 грейд**

- Предложил использовать RWMutex
- Заметил, что два отдельных лока в GetOrCreate могут привести к конкурентной записи

**20 грейд**

- Предложил использовать sync.Map, рассказал чем подход с RWMutex (а за одно и посомневался почему RWMutex может оказаться медленнее)
- Предложил сделать разбить кеш на несколько шардов с отдельным мьютексом
