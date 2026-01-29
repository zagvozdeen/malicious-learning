---
name: Вопрос про "звездочку"
module: Практика
tags:
  - грейды
---

Это почти реальный пример из одного из наших сервисов. Этот код - обертка над кешем, который соотвественно пишет и читает данные из кэша. Несмотря на то, что внедрение данного кэша должно было облегчить основное хранилище, однако это не произошло. Почему?

```go
package main

type Storage struct {
	cache *lru.Cache
}

func (s *Storage) Set(wh *warehouse.Warehouse) {
	s.cache.Put(wh.Id, *wh)
}
func (s *Storage) Get(id types.WarehouseId) *warehouse.Warehouse {
	item, ok := s.cache.Get(id)
	if ok {
		if wh, ok := item.(*warehouse.Warehouse); ok {
			return wh
		}
	}
	return nil
}
```

В кэш кладется `warehouse.Warehouse`, а assert делается с `*warehouse.Warehouse`.