---
name: Задача на прогноз погоды
module: Практика
tags:
  - avito
  - medium
---

- Есть функция которая через нейронную сеть вычисляет прогноз погоды за ~1 секунду.
- Есть highload RPC ручка с нагрузкой 10k RPS
- Необходимо реализовать код этой ручки

1. Какую структуру правильно выбрать для кэша?
2. Как исключить гонки при паралельном доступе к структуре?
3. Как инвалидировать кеш?
4. Как реализовать прогрев кеша перед стартом приложения?

Пример решения:

```go
package main

import (
	"context"
	"math/rand/v2"
	"net/http"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func getTemperature() int {
	time.Sleep(time.Second)
	return rand.IntN(70) - 30
}

var temperature int
var mu sync.RWMutex

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup
	wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
			mu.Lock()
			temperature = getTemperature()
			mu.Unlock()
		}
	})

	http.HandleFunc("GET /temperature", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		t := temperature
		mu.RUnlock()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(strconv.Itoa(t)))
	})
	_ = http.ListenAndServe(":8000", nil)

	wg.Wait()
}
```