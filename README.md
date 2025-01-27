# Sorted Set in Golang

Sorted Set is a data-struct inspired by the one from Redis. It allows fast access by key or score.

| Property | Type           | Description                                                           |
|----------|----------------|-----------------------------------------------------------------------|
| `key`    | `string`       | The identifier of the node. It must be unique within the set.         |
| `value`  | `interface {}` | value associated with this node                                       |
| `score`  | `float64`      | score is in order to take the sorted set ordered. It may be repeated. |

Each node in the set is associated with a `key`. While `key`s are unique, `score`s may be repeated. 
Nodes are __taken in order instead of ordered afterwards__, from low score to high score. If scores are the same, the node is ordered by its key in lexicographic order. Each node in the set is associated with __rank__, which represents the position of the node in the sorted set. The __rank__ is 1-based, that is to say, rank 1 is the node with minimum score.

Sorted Set is implemented basing on skip list and hash map internally. With sorted sets you can add, remove, or update nodes in a very fast way (in a time proportional to the logarithm of the number of nodes). You can also get ranges by score or by rank (position) in a very fast way. Accessing the middle of a sorted set is also very fast, so you can use Sorted Sets as a smart list of non repeating nodes where you can quickly access everything you need: nodes in order, fast existence test, fast access to nodes in the middle!

A typical use case of sorted set is a leader board in a massive online game, where every time a new score is submitted you update it using `AddOrUpdate()`. You can easily take the top users using `GetByRankRange()`, you can also, given an user id, return its rank in the listing using `FindRank()`. Using `FindRank()` and `GetByRankRange()` together you can show users with a score similar to a given user. All very quickly.

## Benchmark

### Environment:

- Model : MacBookPro18,3
- OS : MacOS Monterey 12.5.1
- Chip : Apple M1 Pro
- Number of cores: 8


| Test                                 | Result      |
|--------------------------------------|-------------|
| BenchmarkDefaultDecrementInserts-8   | 572.5 ns/op |
| BenchmarkDefaultIncrementInserts-8   | 466.9 ns/op |
| BenchmarkDefaultPermutationInserts-8 | 2582 ns/op  |
| BenchmarkDefaultRandomInserts-8      | 1521 ns/op  |
| BenchmarkRandomSelectByKey-8         | 335.8 ns/op |
| BenchmarkRandomSearchByScore-8       | 1063 ns/op  |
| BenchmarkDelete-8                    | 1481 ns/op  |
| BenchmarkRandomDelete-8              | 1514 ns/op  |

## Documentation

[https://godoc.org/github.com/tunglt1810/sortedset](https://godoc.org/github.com/tunglt1810/sortedset)

Copyright (c) 2016, Jerry.Wang [https://godoc.org/github.com/wangjia184/sortedset](https://godoc.org/github.com/wangjia184/sortedset)