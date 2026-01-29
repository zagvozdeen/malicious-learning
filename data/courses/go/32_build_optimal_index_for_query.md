---
name: Построить оптимальный индекс для запроса
module: Базы данных
tags:
  - грейды
---
```sql
SELECT * FROM employee
WHERE sex = 'm' AND salary > 300000 AND age = 20
ORDER BY created_at
```
