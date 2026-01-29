---
name: Что выведет программа (на работу со слайсами)?
module: Скрининг
tags:
  - грейды
  - озон
---
Что выведет данная программа? Почему так?

```go
package main  
  
import "fmt"  
  
func main() {  
    nums := []int{1, 2, 3}  
    addNum(nums[0:2])  
    fmt.Println(nums)  
    addNums(nums[0:2])  
    fmt.Println(nums)  
}

func addNum(nums []int) {  
    nums = append(nums, 4)  
}

func addNums(nums []int) {  
    nums = append(nums, 5, 6)  
}
```

**17 грейд**

1) 1 2 4
2) 1 2 4

Расскажет про правила работы `append`.

**18 грейд**

Расскажет про синтаксис `[0:2:2]`