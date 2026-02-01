---
name: Задача на сборку сниппета
module: Практика
tags:
  - avito
  - easy
---

Мы делаем сервис, который собирает сниппет товара. Каждый сниппет состоит из отформатированного описания и стоимости товара в рублях.

Чтобы собрать сниппет, нужно:

- Получить описание из одного сервиса, а затем отформатированное его через `prettify`
- Получить цену (в долларах) из другого сервиса, а затем перевести ее в рубли через `priceToRub`
- Вернуть готовый сниппет

Дополнительные вопросы:

- Как распараллелить пачку задач и дождаться завершения всех? Работа с общей памятью, атомарность операций и необходимость синхронизации.
- Предположим, что сетевые запросы теперь могут выдать ошибку, как это обработать? Какие есть варианты? Показать цену в долларах или спятисотить или ретраить?
- Какой таймаут правильно выставить на вызовы во внешние сервисы?
- Отличие между тредами, процессами и корутинами?
- Как работает веб-сервер?
- Как устроен шедулер, как устроены приоритеты, в какие моменты корутины засыпают и просыпаются?
- Мьютексы и как ими пользоваться (лок на чтение и на запись, отпускание локов после exception'ов)?
- Атомарные операции, compare-and-swap etc, rw-мьютексы, спинлоки, как оно работает внутри?
- Синхронизация между процессами на одной машине и между разными машинами, типовые способы ее избежать?

Заготовка:

```go
package main

import (
	"strconv"
)

type Snippet struct {
	Price       float64
	Description string
}

func itemDescription(itemID int) string { return strconv.Itoa(itemID) }
func prettify(v string) string          { return v }
func itemPrice(itemID int) float64      { return float64(itemID) }
func priceToRub(v float64) float64      { return v }

func BuildSnippet(itemID int) Snippet {
	return Snippet{
		Price:       0,
		Description: "",
	}
}
```

Предполагаемое решение:

```go
package main

import (
	"strconv"
	"sync"
)

type Snippet struct {
	Price       float64
	Description string
}

func itemDescription(itemID int) string { return strconv.Itoa(itemID) }
func prettify(v string) string          { return v }
func itemPrice(itemID int) float64      { return float64(itemID) }
func priceToRub(v float64) float64      { return v }

func BuildSnippet(itemID int) (s Snippet) {
	var wg sync.WaitGroup
	wg.Go(func() {
		rawDescription := itemDescription(itemID)
		s.Description = prettify(rawDescription)
	})
	wg.Go(func() {
		rawPrice := itemPrice(itemID)
		s.Price = priceToRub(rawPrice)
	})
	wg.Wait()
	return
}
```
