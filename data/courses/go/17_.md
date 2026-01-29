---
name: Что будет если изменить слайс во время цикла for?
module: Скрининг
tags:
  - грейды
---
Что выведет код программы и почему?

```go
package main  
  
import "fmt"  
  
func main() {  
    lst := []string{"a", "b", "c", "d"}  
    for k, v := range lst {  
       if k == 0 {  
          lst = []string{"aa", "bb", "cc", "dd"}  
       }  
       fmt.Println(v)  
    }  
}
```

Это задача даже не на специфику работы оператора range, а просто на знание одного из основных принципов языка — everything in Go is passed by value. Массив является по сути struct 'ом из трех полей — `ссылка на array`, `len` и `cap`, при вызове range происходит ровно то же самое, что и при вызове функции — эта структура копируется (а array на который ссылаемся — нет). Присваивая переменной новое значение мы никак не меняем исходный array, поэтому и выведется:

```
a
b
c
d
```

А вот если менять значения в нижележащем array'е, то они будут меняться, и range их увидит:

```go
package main  
  
import "fmt"  
  
func main() {  
    lst := []string{"a", "b", "c", "d"}  
    for k, v := range lst {  
       if k == 0 {  
          lst[3] = "z"  
       }  
       fmt.Println(v)  
    }  
}
```

Код выведет:

```
a
b
c
z
```
