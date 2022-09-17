package sortedset

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func checkOrder(t *testing.T, nodes []*Node, expectedOrder []string) {
	if len(expectedOrder) != len(nodes) {
		t.Errorf("nodes does not contain %d elements", len(expectedOrder))
	}
	for i := 0; i < len(expectedOrder); i++ {
		if nodes[i].Key() != expectedOrder[i] {
			t.Errorf("nodes[%d] is %q, but the expected key is %q", i, nodes[i].Key(), expectedOrder[i])
		}

	}
}

func checkIterByRankRange(t *testing.T, sortedset *SortedSet, start int, end int, expectedOrder []string) {
	var keys []string

	// check nil callback should do nothing
	sortedset.IterFuncByRankRange(start, end, nil)

	sortedset.IterFuncByRankRange(start, end, func(key string, _ interface{}) bool {
		keys = append(keys, key)
		return true
	})
	if len(expectedOrder) != len(keys) {
		t.Errorf("keys does not contain %d elements", len(expectedOrder))
	}
	for i := 0; i < len(expectedOrder); i++ {
		if keys[i] != expectedOrder[i] {
			t.Errorf("keys[%d] is %q, but the expected key is %q", i, keys[i], expectedOrder[i])
		}
	}

	// check return early
	if len(expectedOrder) < 1 {
		return
	}
	// reset data
	keys = []string{}
	var i int
	sortedset.IterFuncByRankRange(start, end, func(key string, _ interface{}) bool {
		keys = append(keys, key)
		i++
		// return early
		return i < len(expectedOrder)-1
	})
	if len(expectedOrder)-1 != len(keys) {
		t.Errorf("keys does not contain %d elements", len(expectedOrder)-1)
	}
	for i := 0; i < len(expectedOrder)-1; i++ {
		if keys[i] != expectedOrder[i] {
			t.Errorf("keys[%d] is %q, but the expected key is %q", i, keys[i], expectedOrder[i])
		}
	}

}

func checkRankRangeIterAndOrder(t *testing.T, sortedset *SortedSet, start int, end int, remove bool, expectedOrder []string) {
	checkIterByRankRange(t, sortedset, start, end, expectedOrder)
	nodes := sortedset.GetByRankRange(start, end, remove)
	checkOrder(t, nodes, expectedOrder)
}

func TestCase1(t *testing.T) {
	sortedset := New()

	sortedset.AddOrUpdate("a", 89, "Kelly")
	sortedset.AddOrUpdate("b", 100, "Staley")
	sortedset.AddOrUpdate("c", 100, "Jordon")
	sortedset.AddOrUpdate("d", -321, "Park")
	sortedset.AddOrUpdate("e", 101, "Albert")
	sortedset.AddOrUpdate("f", 99, "Lyman")
	sortedset.AddOrUpdate("g", 99, "Singleton")
	sortedset.AddOrUpdate("h", 70, "Audrey")

	sortedset.AddOrUpdate("e", 99, "ntrnrt")

	sortedset.Remove("b")

	node := sortedset.GetByRank(3, false)
	if node == nil || node.Key() != "a" {
		t.Error("GetByRank() does not return expected value `a`")
	}

	node = sortedset.GetByRank(-3, false)
	if node == nil || node.Key() != "f" {
		t.Error("GetByRank() does not return expected value `f`")
	}

	// get all nodes since the first one to last one
	checkRankRangeIterAndOrder(t, sortedset, 1, -1, false, []string{"d", "h", "a", "e", "f", "g", "c"})

	// get & remove the 2nd/3rd nodes in reserve order
	checkRankRangeIterAndOrder(t, sortedset, -2, -3, true, []string{"g", "f"})

	// get all nodes since the last one to first one
	checkRankRangeIterAndOrder(t, sortedset, -1, 1, false, []string{"c", "e", "a", "h", "d"})

}

func TestCase2(t *testing.T) {

	// create a new set
	sortedset := New()

	// fill in new node
	sortedset.AddOrUpdate("a", 89, "Kelly")
	sortedset.AddOrUpdate("b", 100, "Staley")
	sortedset.AddOrUpdate("c", 100, "Jordon")
	sortedset.AddOrUpdate("d", -321, "Park")
	sortedset.AddOrUpdate("e", 101, "Albert")
	sortedset.AddOrUpdate("f", 99, "Lyman")
	sortedset.AddOrUpdate("g", 99, "Singleton")
	sortedset.AddOrUpdate("h", 70, "Audrey")

	// update an existing node
	sortedset.AddOrUpdate("e", 99, "ntrnrt")

	// remove node
	sortedset.Remove("b")

	nodes := sortedset.GetByScoreRange(-500, 500, nil)
	checkOrder(t, nodes, []string{"d", "h", "a", "e", "f", "g", "c"})

	nodes = sortedset.GetByScoreRange(500, -500, nil)
	//t.Logf("%v", nodes)
	checkOrder(t, nodes, []string{"c", "g", "f", "e", "a", "h", "d"})

	nodes = sortedset.GetByScoreRange(600, 500, nil)
	checkOrder(t, nodes, []string{})

	nodes = sortedset.GetByScoreRange(500, 600, nil)
	checkOrder(t, nodes, []string{})

	rank := sortedset.FindRank("f")
	if rank != 5 {
		t.Error("FindRank() does not return expected value `5`")
	}

	rank = sortedset.FindRank("d")
	if rank != 1 {
		t.Error("FindRank() does not return expected value `1`")
	}

	nodes = sortedset.GetByScoreRange(99, 100, nil)
	checkOrder(t, nodes, []string{"e", "f", "g", "c"})

	nodes = sortedset.GetByScoreRange(90, 50, nil)
	checkOrder(t, nodes, []string{"a", "h"})

	nodes = sortedset.GetByScoreRange(99, 100, &GetByScoreRangeOptions{
		ExcludeStart: true,
	})
	checkOrder(t, nodes, []string{"c"})

	nodes = sortedset.GetByScoreRange(100, 99, &GetByScoreRangeOptions{
		ExcludeStart: true,
	})
	checkOrder(t, nodes, []string{"g", "f", "e"})

	nodes = sortedset.GetByScoreRange(99, 100, &GetByScoreRangeOptions{
		ExcludeEnd: true,
	})
	checkOrder(t, nodes, []string{"e", "f", "g"})

	nodes = sortedset.GetByScoreRange(100, 99, &GetByScoreRangeOptions{
		ExcludeEnd: true,
	})
	checkOrder(t, nodes, []string{"c"})

	nodes = sortedset.GetByScoreRange(50, 100, &GetByScoreRangeOptions{
		Limit: 2,
	})
	checkOrder(t, nodes, []string{"h", "a"})

	nodes = sortedset.GetByScoreRange(100, 50, &GetByScoreRangeOptions{
		Limit: 2,
	})
	checkOrder(t, nodes, []string{"c", "g"})

	minNode := sortedset.PeekMin()
	if minNode == nil || minNode.Key() != "d" {
		t.Error("PeekMin() does not return expected value `d`")
	}

	minNode = sortedset.PopMin()
	if minNode == nil || minNode.Key() != "d" {
		t.Error("PopMin() does not return expected value `d`")
	}

	nodes = sortedset.GetByScoreRange(-500, 500, nil)
	checkOrder(t, nodes, []string{"h", "a", "e", "f", "g", "c"})

	maxNode := sortedset.PeekMax()
	if maxNode == nil || maxNode.Key() != "c" {
		t.Error("PeekMax() does not return expected value `c`")
	}

	maxNode = sortedset.PopMax()
	if maxNode == nil || maxNode.Key() != "c" {
		t.Error("PopMax() does not return expected value `c`")
	}

	nodes = sortedset.GetByScoreRange(500, -500, nil)
	checkOrder(t, nodes, []string{"g", "f", "e", "a", "h"})
}

func TestSortedSet_GetRandomByScoreRange(t *testing.T) {
	// create a new set
	sortedset := New()

	// fill in new node
	sortedset.AddOrUpdate("a", 89, "Kelly")
	sortedset.AddOrUpdate("b", 100, "Staley")
	sortedset.AddOrUpdate("c", 100, "Jordon")
	sortedset.AddOrUpdate("d", -321, "Park")
	sortedset.AddOrUpdate("e", 101, "Albert")
	sortedset.AddOrUpdate("f", 99, "Lyman")
	sortedset.AddOrUpdate("g", 99, "Singleton")
	sortedset.AddOrUpdate("h", 70, "Audrey")

	nodes := sortedset.GetRandomByScoreRange(99, 100, nil)
	if len(nodes) != 4 {
		t.Errorf("GetRandom with no limit should return 4 nodes")
	}
	for _, node := range nodes {
		if node.score < 99 || node.score > 100 {
			t.Errorf("Selected node should have score from 99 to 100")
		}
	}

	nodes = sortedset.GetRandomByScoreRange(99, 101, &GetByScoreRangeOptions{
		Limit:        10,
		ExcludeStart: true,
		ExcludeEnd:   true,
	})
	if len(nodes) != 2 {
		t.Errorf("GetRandom with exclude start and end index should return 2 nodes")
	}
	for _, node := range nodes {
		if math.Abs(node.score-100) > eps {
			t.Errorf("Selected node should have score equal 100")
		}
	}

	countB := 0
	countC := 0
	for i := 0; i < 1000; i++ {
		nodes = sortedset.GetRandomByScoreRange(99, 101, &GetByScoreRangeOptions{
			Limit:        1,
			ExcludeStart: true,
			ExcludeEnd:   true,
		})
		if len(nodes) != 1 {
			t.Errorf("GetRandom with exclude start and end index should return 1 nodes")
		}
		if nodes[0].key == "b" {
			countB++
		} else if nodes[0].key == "c" {
			countC++
		}
	}
	if countB == 0 || countC == 0 {
		t.Errorf("countB = %d - countC = %d, selected node should be random between \"b\" or \"c\"", countB, countB)
	}
	if countB+countC != 1000 {
		t.Errorf("Total select count should be 1000")
	}
}

func TestBackwardAndForward(t *testing.T) {
	// create a new set
	sortedset := New()

	// fill in new node
	sortedset.AddOrUpdate("a", 99, "Alice")
	sortedset.AddOrUpdate("b", 100, "Bob")
	sortedset.AddOrUpdate("c", 101, "Carol")

	bNode := sortedset.GetByKey("b")

	previous := bNode.Previous()
	if previous == nil {
		t.Errorf("Previous node of \"b\" shoulds not nil")
	}
	if previous.Key() != "a" || previous.Value != "Alice" {
		t.Errorf("Previous node of \"b\" shoulds be \"a\"")
	}

	next := bNode.Next()
	if next == nil {
		t.Errorf("Next node of \"b\" shoulds not nil")
	}
	if next.Key() != "c" || next.Value != "Carol" {
		t.Errorf("Next node of \"b\" shoulds be \"c\"")
	}

}

func BenchmarkDefaultDecrementInserts(b *testing.B) {
	list := New()

	for i := 0; i < b.N; i++ {
		list.AddOrUpdate(fmt.Sprintf("%d", i), float64(i), "")
	}
}

func BenchmarkDefaultIncrementInserts(b *testing.B) {
	list := New()

	for i := 0; i < b.N; i++ {
		list.AddOrUpdate(fmt.Sprintf("%d", b.N-i), float64(b.N-i), "")
	}
}

func BenchmarkDefaultPermutationInserts(b *testing.B) {
	list := New()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rList := r.Perm(b.N)
	for i := 0; i < b.N; i++ {
		score := rList[i]
		list.AddOrUpdate(fmt.Sprintf("%d", score), float64(score), "")
	}
}

func BenchmarkDefaultRandomInserts(b *testing.B) {
	list := New()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < b.N; i++ {
		score := r.Intn(b.N)
		list.AddOrUpdate(fmt.Sprintf("%d", score), float64(score), "")
	}
}

func BenchmarkRandomSelectByKey(b *testing.B) {
	list := New()
	keys := make([]int, 0, b.N)

	for i := 0; i < b.N; i++ {
		keys = append(keys, i)
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Shuffle(b.N, func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	for i := 0; i < b.N; i++ {
		list.AddOrUpdate(fmt.Sprintf("%d", keys[i]), float64(i), "")
	}

	rnd.Shuffle(b.N, func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		list.GetByKey(fmt.Sprintf("%d", keys[i]))
	}
}

func BenchmarkRandomSearchByScore(b *testing.B) {
	list := New()
	keys := make([]int, 0, b.N)

	for i := 0; i < b.N; i++ {
		keys = append(keys, i)
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Shuffle(b.N, func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	for i := 0; i < b.N; i++ {
		list.AddOrUpdate(fmt.Sprintf("%d", keys[i]), float64(i), "")
	}

	rnd.Shuffle(b.N, func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		list.GetByScoreRange(float64(keys[i]), float64(keys[i]), &GetByScoreRangeOptions{
			Limit:        0,
			ExcludeStart: true,
			ExcludeEnd:   true,
		})
	}
}

func BenchmarkDelete(b *testing.B) {
	list := New()
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	keys := rnd.Perm(b.N)

	for i := 0; i < b.N; i++ {
		list.AddOrUpdate(fmt.Sprintf("%d", keys[i]), float64(i), "")
	}

	rnd.Shuffle(b.N, func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		list.Remove(fmt.Sprintf("%d", i))
	}
}

func BenchmarkRandomDelete(b *testing.B) {
	list := New()
	keys := make([]int, 0, b.N)

	for i := 0; i < b.N; i++ {
		keys = append(keys, i)
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Shuffle(b.N, func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	for i := 0; i < b.N; i++ {
		list.AddOrUpdate(fmt.Sprintf("%d", keys[i]), float64(i), "")
	}

	rnd.Shuffle(b.N, func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		list.Remove(fmt.Sprintf("%d", keys[i]))
	}
}
