---
name: SQL запрос на JOIN нескольких таблиц с условием
module: Скрининг
tags:
  - грейды
---
Есть база с такой схемой данных:

```text
// user
id | firstname | lastname | birth
1  | Ivan      | Petrov   | 1996-05-01
2  | Anna      | Petrova  | 1999-06-01
3  | Anna      | Petrova  | 1990-10-02

// purchase
sku | price | user_id | date
1   | 5500  | 1       | 2021-02-15
1   | 5700  | 1       | 2021-01-15
2   | 4000  | 1       | 2021-02-14
3   | 8000  | 2       | 2021-03-01
4   | 400   | 2       | 2021-03-02

// ban_list
user_id | date_from
1       | 2021-03-08
```

Нужно вывести:

1. Вывести уникальные комбинации пользователя и id товара для всех покупок, совершенных пользователями до того, как их забанили. Отсортировать сначала по имени пользователя, потом по SKU
2. Найти пользователей, которые совершили покупок на сумму больше 5000р. Вывести их имена в формате id пользователя | имя | фамилия | сумма покупок

Ответы:

1. Здесь мы проверяем что человек умеет в джоины, distinct, where, order

```sql
SELECT distinct u.id, firstname, lastname, p.item_id
FROM users u
         join purchase p ON u.id = p.user_id
         left join ban_list bl ON u.id = bl.user_id -- не забыть left join, а то в запрос попадут только покупки забаненных пользователей.
WHERE bl.user_id IS NULL    -- пользователь не забанен
   OR bl.date_from > p.date -- или забанен позже, чем совершена покупка
ORDER BY lastname, firstname, u.id, p.item_id -- лучше бы кандидат догадался или спросил, что в сортировке по имени надо сначала ставить фамилию, потом имя.
```

2. Здесь мы проверяем, что человек умеет в HAVING, и знает, чем HAVING отличается от WHERE:

```sql
SELECT u.id, u.firstname, u.lastname, SUM(p.price)
FROM users u
         join purchase p ON u.id = p.user_id
GROUP BY u.id, u.firstname, u.lastname -- и знает, что аггрегирующие функции без group by не будут работать
HAVING SUM(p.price) > 5000;

-- В принципе, вариант тоже является корректным, но все равно надо спросить про HAVING

SELECT *
FROM (SELECT u.id, u.firstname, u.lastname, SUM(p.price) s
      FROM users u
               join purchase p ON u.id = p.user_id
      GROUP BY u.id, u.firstname, u.lastname)
WHERE s > 5000;
```