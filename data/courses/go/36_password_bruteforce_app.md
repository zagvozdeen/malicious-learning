---
name: Напиши приложение для перебора паролей
module: Практика
tags:
  - грейды
---
У нас есть база данных с паролями пользователей, пароли захешированы (функция
hashPassword), а так же известен набор символов которые могут быть использованы в
паролях (переменная alphabet).
Наша задача реализовать функцию RecoverPassword так, чтобы она восстанавливала
пароль по известному хэшу и TestRecoverPassword завершился успешно

Базовые требования: решить как угодно.

```go
package main  
  
import (  
    "crypto/md5"  
    "testing")  
  
var alphabet = []rune{'a', 'b', 'c', 'd', '1', '2', '3'}  
  
func RecoverPassword(h []byte) string {  
    // TODO: implement me  
    return ""  
}  
  
func TestRecoverPassword(t *testing.T) {  
    for _, exp := range []string{  
       "a",  
       "12",  
       "abc333d",  
    } {  
       t.Run(exp, func(t *testing.T) {  
          act := RecoverPassword(hashPassword(exp))  
          if act != exp {  
             t.Error("recovered:", act, "expected:", exp)  
          }  
       })  
    }  
}  
func hashPassword(in string) []byte {  
    h := md5.Sum([]byte(in))  
    return h[:]  
}
```

Это вопрос на написание кода, который потом будет приложен к протоколу собеседования.
Простой пример по перебору (максимальное число перебираемых комбинаций - math.MaxInt64).

```go
package main  
  
import (  
    "bytes"  
    "crypto/md5"
	"testing"
)  
  
var alphabet = []rune{'a', 'b', 'c', 'd', '1', '2', '3'}  
  
func RecoverPassword(h []byte) string {  
    var step int  
    for ; ; step++ {  
       guess := genPassword(step)  
       if bytes.Equal(hashPassword(guess), h) {  
          return guess  
       }  
    }  
}  
  
func genPassword(step int) (res string) {  
    for {  
       res = string(alphabet[step%len(alphabet)]) + res  
       step = step/len(alphabet) - 1  
       if step < 0 {  
          break  
       }  
    }  
    return  
}  
  
func TestRecoverPassword(t *testing.T) {  
    for _, exp := range []string{  
       "a",  
       "12",  
       "abc333d",  
    } {  
       t.Run(exp, func(t *testing.T) {  
          act := RecoverPassword(hashPassword(exp))  
          if act != exp {  
             t.Error("recovered:", act, "expected:", exp)  
          }  
       })  
    }  
}  
  
func hashPassword(in string) []byte {  
    h := md5.Sum([]byte(in))  
    return h[:]  
}
```

**Дополнительные вопросы**
Q: Как сделать подбор константным по сложности, если мы можем ограничить длину пароля?
A: Использовать rainbow table.

Q: Вычислительная сложность подбора пароля?
A: O(a^n), а для n-битовой хеш-функции сложность нахождения первого прообраза
составляет O(2^n)

Q: Как атакующий может скомпрометировать криптосистему?
A: Через timing-attack. Защитой будет сравнение за константное время, кол-во попыток,
ограничение по времени в случае одноразовых паролей из SMS.

Q: Как разработчики сервиса могли бы усложнить подбор паролей?
A: крипто стойкое хеширование, соль, сравнение за константное время
