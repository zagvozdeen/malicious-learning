---
name: За счет чего Go позволяет работать с большим количеством сетевых соединений одновременно?
module: Теория
tags:
  - avito
---

**Basic**

- Netpoll, неблокирующие сокеты

**Advanced**

- `Epoll`, `kqueue`, `IOCP`
- Знает преимущества `epoll` над `select`, `poll`
- Знает какие события в дескрипторе можно зарегистрировать

**Expert**

- Работал с `epoll`, kqueue писал свою реализацию
- Знает отличия `edge-triggered` от `level-triggered`
