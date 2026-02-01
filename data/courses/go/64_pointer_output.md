---
name: Что выведет код ниже (задача на указатели)?
module: Практика
tags:
  - avito
  - easy
---

Что выведет программа и почему?

```go
package main

import "fmt"

type Person struct {
	Name string
}

func changeName(person *Person) {
	person = &Person{
		Name: "Alice",
	}
}

func main() {
	person := &Person{
		Name: "Bob",
	}
	fmt.Println(person.Name)
	changeName(person)
	fmt.Println(person.Name)
}
```

Ответ: ||`Bob Bob`||

||Как модифицировать программу, чтобы получить `Bob Alice`?||
