# Nimble - A lightweight Redis clone created for educational purposes.

[![Golang](https://img.shields.io/badge/Golang-%2300ADD8.svg?&logo=go&logoColor=white)](#)

# 📚 Commands

Full list of supported commands grouped by category.

## 🔤 STRING

| Command | Description | Example |
| :---: | :---: | :---: |
| SET | Save string value by key | `SET key value` |
| MSET | Save multiple string values | `MSET key1 value1 key2 value2` |
| GET | Get value by key | `GET key` |
| MGET | Get multiple values | `MGET key1 key2` |

---

## ⚙️ GENERAL

| Command | Description | Example |
| :---: | :---: | :---: |
| DEL | Delete keys | `DEL key1 key2` |
| COPY | Copy value to another key | `COPY source_key target_key` |
| RENAME | Rename key | `RENAME old_key new_key` |
| TYPE | Get key data type | `TYPE key` |
| KEYS | Find keys by pattern | `KEYS user:???` `KEYS user:*` |
| LIST | Get all keys | `LIST` |
| LISTLEN | Get total number of keys | `LISTLEN` |
| EXISTS | Check if keys exist (count) | `EXISTS key1 key2` |
| LEXISTS | Check if keys exist (list) | `LEXISTS key1 key2` |

---

## 🔢 NUMERIC

| Command | Description | Example |
| :---: | :---: | :---: |
| INCR | Increment value by 1 | `INCR key` |
| INCRBY | Increment value by N | `INCRBY key 5` |
| DECR | Decrement value by 1 | `DECR key` |
| DECRBY | Decrement value by N | `DECRBY key 5` |
| MUL | Multiply value by N | `MUL key 2` |
| DIV | Divide value by N | `DIV key 2` |

---

## 📋 LIST

| Command | Description | Example |
| :---: | :---: | :---: |
| ESET | Create empty list | `ESET mylist` |
| LSET | Set values by index | `LSET mylist 0 value` |
| LGET | Get values by index | `LGET mylist 0` |
| LCLEAR | Clear list | `LCLEAR mylist` |
| LLEN | Get list length | `LLEN mylist` |
| SPUSH | Push to start | `SPUSH mylist value` |
| EPUSH | Push to end | `EPUSH mylist value` |
| SPOP | Pop from start | `SPOP mylist 1` |
| EPOP | Pop from end | `EPOP mylist 1` |
| SRANGE | Get range | `SRANGE mylist 0 5` |
| CONTAINS | Check values (count) | `CONTAINS mylist value` |
| LCONTAINS | Check values (list) | `LCONTAINS mylist value` |
| INDEXOF | Get index of value | `INDEXOF mylist value` |

---

## 🧩 HASH

| Command | Description | Example |
| :---: | :---: | :---: |
| HSET | Set hash fields | `HSET user name John` |
| HGET | Get hash fields | `HGET user name` |
| HDEL | Delete hash fields | `HDEL user name` |
| HLEN | Number of fields | `HLEN user` |
| HKEYS | Get field names | `HKEYS user` |
| HVALUES | Get field values | `HVALUES user` |
| HCONTAINS | Check fields (count) | `HCONTAINS user name` |
| LHCONTAINS | Check fields (list) | `LHCONTAINS user name` |

---

## 🟢 SET

| Command | Description | Example |
| :---: | :---: | :---: |
| SADD | Add elements | `SADD myset value` |
| SREM | Remove elements | `SREM myset value` |
| SLEN | Number of elements | `SLEN myset` |
| SMEMBERS | Get all elements | `SMEMBERS myset` |
| SCONTAINS | Check elements (count) | `SCONTAINS myset value` |
| LSCONTAINS | Check elements (list) | `LSCONTAINS myset value` |

---

## 📊 ZSET

| Command | Description | Example |
| :---: | :---: | :---: |
| ZADD | Add element with score | `ZADD leaderboard user1 100` |
| ZREM | Remove element | `ZREM leaderboard user1` |
| ZRANGE | Get by score range | `ZRANGE leaderboard 0 100` |
| SCORE | Get element score | `SCORE leaderboard user1` |
| LSCORE | Get multiple scores | `LSCORE leaderboard user1 user2` |
| SLEN | Number of elements | `SLEN leaderboard` |
| SMEMBERS | Get all elements | `SMEMBERS leaderboard` |
| SCONTAINS | Check elements (count) | `SCONTAINS leaderboard user1` |
| LSCONTAINS | Check elements (list) | `LSCONTAINS leaderboard user1` |

---

## ⏳ EXPIRATION

| Command | Description | Example |
| :---: | :---: | :---: |
| TTK | Set time to delete (in seconds) | `TTK key 60` |
| TTL | Get remaining time | `TTL key` |

---

## 🖥 SERVER

| Command | Description | Example |
| :---: | :---: | :---: |
| MODE | Change server mode | `MODE READONLY` |
| PING | Check server status | `PING` |

---

## 🧹 CONSOLE

| Command | Description | Example |
| :---: | :---: | :---: |
| CLS | Clear console output | `CLS` |
