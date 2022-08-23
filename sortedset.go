package sortedset

import (
	"math"
	"math/rand"
	"time"
)

const (
	SkiplistMaxLevel  = 32   /* Should be enough for 2^32 elements */
	SkiplistLevelRate = 0.25 /* Skiplist P = 1/4 */
	eps               = 0.00001
)

type SortedSet struct {
	header *Node
	tail   *Node
	length int
	level  int
	dict   map[string]*Node
	r      *rand.Rand
}

func createNode(level int, score float64, key string, value interface{}) *Node {
	node := Node{
		score: score,
		key:   key,
		Value: value,
		level: make([]Level, level),
	}
	return &node
}

// Returns a random level for the new skiplist node we are going to create.
// The return value of this function is between 1 and SkiplistMaxLevel
// (both inclusive), with a power-law-alike distribution where higher
// levels are less likely to be returned.
func (set *SortedSet) randomLevel() int {
	level := 1
	for float64(set.r.Int31()&0xFFFF) < SkiplistLevelRate*0xFFFF {
		level += 1
	}
	if level < SkiplistMaxLevel {
		return level
	}

	return SkiplistMaxLevel
}

func (set *SortedSet) insertNode(score float64, key string, value interface{}) *Node {
	var update [SkiplistMaxLevel]*Node
	var rank [SkiplistMaxLevel]int

	x := set.header
	for i := set.level - 1; i >= 0; i-- {
		/* store rank that is crossed to reach the insert position */
		if set.level-1 == i {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(math.Abs(x.level[i].forward.score-score) < eps && // score is the same but the key is different
					x.level[i].forward.key < key)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}

	/* we assume the key is not already inside, since we allow duplicated
	 * scores, and the re-insertion of score and redis object should never
	 * happen since the caller of Insert() should test in the hash table
	 * if the element is already inside or not. */
	level := set.randomLevel()

	if level > set.level { // add a new level
		for i := set.level; i < level; i++ {
			rank[i] = 0
			update[i] = set.header
			update[i].level[i].span = set.length
		}
		set.level = level
	}

	x = createNode(level, score, key, value)
	for i := 0; i < level; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		/* update span covered by update[i] as x is inserted here */
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	/* increment span for untouched levels */
	for i := level; i < set.level; i++ {
		update[i].level[i].span++
	}

	if update[0] == set.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		set.tail = x
	}
	set.length++
	return x
}

/* Internal function used by delete, DeleteByScore and DeleteByRank */
func (set *SortedSet) deleteNode(x *Node, update [SkiplistMaxLevel]*Node) {
	for i := 0; i < set.level; i++ {
		if update[i].level[i].forward == x {
			update[i].level[i].span += x.level[i].span - 1
			update[i].level[i].forward = x.level[i].forward
		} else {
			update[i].level[i].span -= 1
		}
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		set.tail = x.backward
	}
	for set.level > 1 && set.header.level[set.level-1].forward == nil {
		set.level--
	}
	set.length--
	delete(set.dict, x.key)
}

/* Delete an element with matching score/key from the skiplist. */
func (set *SortedSet) delete(score float64, key string) bool {
	var update [SkiplistMaxLevel]*Node

	x := set.header
	for i := set.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(math.Abs(x.level[i].forward.score-score) < eps &&
					x.level[i].forward.key < key)) {
			x = x.level[i].forward
		}
		update[i] = x
	}
	/* We may have multiple elements with the same score, what we need
	 * is to find the element with both the right score and object. */
	x = x.level[0].forward
	if x != nil && math.Abs(score-x.score) < eps && x.key == key {
		set.deleteNode(x, update)
		// free x
		return true
	}
	return false /* not found */
}

// New Create a new SortedSet
func New() *SortedSet {
	sortedSet := SortedSet{
		header: createNode(SkiplistMaxLevel, math.Inf(-1), "", nil),
		level:  1,
		dict:   make(map[string]*Node),
		r:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return &sortedSet
}

// GetCount Get the number of elements
func (set *SortedSet) GetCount() int {
	return set.length
}

// PeekMin get the element with minimum score, nil if the set is empty
//
// Time complexity of this method is : O(1)
func (set *SortedSet) PeekMin() *Node {
	return set.header.level[0].forward
}

// PopMin get and remove the element with minimal score, nil if the set is empty
//
// Time complexity of this method is : O(log(N))
func (set *SortedSet) PopMin() *Node {
	x := set.header.level[0].forward
	if x != nil {
		set.Remove(x.key)
	}
	return x
}

// PeekMax get the element with maximum score, nil if the set is empty
//
// Time Complexity : O(1)
func (set *SortedSet) PeekMax() *Node {
	return set.tail
}

// PopMax get and remove the element with maximum score, nil if the set is empty
//
// Time complexity of this method is : O(log(N))
func (set *SortedSet) PopMax() *Node {
	x := set.tail
	if x != nil {
		set.Remove(x.key)
	}
	return x
}

// AddOrUpdate Add an element into the sorted set with specific key / value / score.
// if the element is added, this method returns true; otherwise false means updated
//
// Time complexity of this method is : O(log(N))
func (set *SortedSet) AddOrUpdate(key string, score float64, value interface{}) bool {
	var newNode *Node = nil

	found := set.dict[key]
	if found != nil {
		// score does not change, only update value
		if math.Abs(found.score-score) < eps {
			found.Value = value
		} else { // score changes, delete and re-insert
			set.delete(found.score, found.key)
			newNode = set.insertNode(score, key, value)
		}
	} else {
		newNode = set.insertNode(score, key, value)
	}

	if newNode != nil {
		set.dict[key] = newNode
	}
	return found == nil
}

// Remove Delete element specified by key
//
// Time complexity of this method is : O(log(N))
func (set *SortedSet) Remove(key string) *Node {
	found := set.dict[key]
	if found != nil {
		set.delete(found.score, found.key)
		return found
	}
	return nil
}

type GetByScoreRangeOptions struct {
	Limit        int  // limit the max nodes to return
	ExcludeStart bool // exclude start value, so it search in interval (start, end] or (start, end)
	ExcludeEnd   bool // exclude end value, so it search in interval [start, end) or (start, end)
}

// GetByScoreRange Get the nodes whose score within the specific range
//
// If options is nil, it searches in interval [minScore, maxScore] without any limit by default
//
// Time complexity of this method is : O(log(N))
func (set *SortedSet) GetByScoreRange(minScore float64, maxScore float64, options *GetByScoreRangeOptions) []*Node {

	// prepare parameters
	var limit = int((^uint(0)) >> 1)
	if options != nil && options.Limit > 0 {
		limit = options.Limit
	}

	excludeStart := options != nil && options.ExcludeStart
	excludeEnd := options != nil && options.ExcludeEnd
	reverse := minScore > maxScore
	if reverse {
		minScore, maxScore = maxScore, minScore
		excludeStart, excludeEnd = excludeEnd, excludeStart
	}

	//////////////////////////
	var nodes []*Node

	//determine if out of range
	if set.length == 0 {
		return nodes
	}
	//////////////////////////

	if reverse { // search from maxScore to minScore
		x := set.header

		if excludeEnd {
			for i := set.level - 1; i >= 0; i-- {
				for x.level[i].forward != nil &&
					x.level[i].forward.score < maxScore {
					x = x.level[i].forward
				}
			}
		} else {
			for i := set.level - 1; i >= 0; i-- {
				for x.level[i].forward != nil &&
					x.level[i].forward.score <= maxScore {
					x = x.level[i].forward
				}
			}
		}

		for x != nil && limit > 0 {
			if excludeStart {
				if x.score <= minScore {
					break
				}
			} else {
				if x.score < minScore {
					break
				}
			}

			next := x.backward

			nodes = append(nodes, x)
			limit--

			x = next
		}
	} else {
		// search from minScore to maxScore
		x := set.header
		if excludeStart {
			for i := set.level - 1; i >= 0; i-- {
				for x.level[i].forward != nil &&
					x.level[i].forward.score <= minScore {
					x = x.level[i].forward
				}
			}
		} else {
			for i := set.level - 1; i >= 0; i-- {
				for x.level[i].forward != nil &&
					x.level[i].forward.score < minScore {
					x = x.level[i].forward
				}
			}
		}

		/* Current node is the last with score < or <= minScore. */
		x = x.level[0].forward

		for x != nil && limit > 0 {
			if excludeEnd {
				if x.score >= maxScore {
					break
				}
			} else {
				if x.score > maxScore {
					break
				}
			}

			next := x.level[0].forward

			nodes = append(nodes, x)
			limit--

			x = next
		}
	}

	return nodes
}

// sanitizeIndexes return start, end, and reverse flag
func (set *SortedSet) sanitizeIndexes(start int, end int) (int, int, bool) {
	if start < 0 {
		start = set.length + start + 1
	}
	if end < 0 {
		end = set.length + end + 1
	}
	if start <= 0 {
		start = 1
	}
	if end <= 0 {
		end = 1
	}

	reverse := start > end
	if reverse { // swap start and end
		start, end = end, start
	}
	return start, end, reverse
}

func (set *SortedSet) findNodeByRank(start int, remove bool) (traversed int, x *Node, update [SkiplistMaxLevel]*Node) {
	x = set.header
	for i := set.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			traversed+x.level[i].span < start {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		if remove {
			update[i] = x
		} else {
			if traversed+1 == start {
				break
			}
		}
	}
	return
}

// GetByRankRange Get nodes within specific rank range [start, end]
// Note that the rank is 1-based integer. Rank 1 means the first node; Rank -1 means the last node;
//
// If start is greater than end, the returned array is in reserved order
// If remove is true, the returned nodes are removed
//
// Time complexity of this method is : O(log(N))
func (set *SortedSet) GetByRankRange(start int, end int, remove bool) []*Node {
	start, end, reverse := set.sanitizeIndexes(start, end)

	var nodes []*Node

	traversed, x, update := set.findNodeByRank(start, remove)

	traversed++
	x = x.level[0].forward
	for x != nil && traversed <= end {
		next := x.level[0].forward

		nodes = append(nodes, x)

		if remove {
			set.deleteNode(x, update)
		}

		traversed++
		x = next
	}

	if reverse {
		for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
			nodes[i], nodes[j] = nodes[j], nodes[i]
		}
	}
	return nodes
}

// GetByRank Get node by rank.
// Note that the rank is 1-based integer. Rank 1 means the first node; Rank -1 means the last node;
//
// If remove is true, the returned nodes are removed
// If node is not found at specific rank, nil is returned
//
// Time complexity of this method is : O(log(N))
func (set *SortedSet) GetByRank(rank int, remove bool) *Node {
	nodes := set.GetByRankRange(rank, rank, remove)
	if len(nodes) == 1 {
		return nodes[0]
	}
	return nil
}

// GetByKey Get node by key
//
// If node is not found, nil is returned
// Time complexity : O(1)
func (set *SortedSet) GetByKey(key string) *Node {
	return set.dict[key]
}

// FindRank Find the rank of the node specified by key
// Please note that the rank is 1-based integer. Rank 1 means the first node
//
// If the node is not found, 0 is returned. Otherwise rank(> 0) is returned
//
// Time complexity of this method is : O(log(N))
func (set *SortedSet) FindRank(key string) int {
	var rank = 0
	node := set.dict[key]
	if node != nil {
		x := set.header
		for i := set.level - 1; i >= 0; i-- {
			for x.level[i].forward != nil &&
				(x.level[i].forward.score < node.score ||
					(math.Abs(x.level[i].forward.score-node.score) < eps &&
						x.level[i].forward.key <= node.key)) {
				rank += x.level[i].span
				x = x.level[i].forward
			}

			if x.key == key {
				return rank
			}
		}
	}
	return 0
}

// IterFuncByRankRange apply fn to node within specific rank range [start, end]
// or until fn return false
//
// Note that the rank is 1-based integer. Rank 1 means the first node; Rank -1 means the last node;
// If start is greater than end, apply fn in reserved order
// If fn is nil, this function return without doing anything
func (set *SortedSet) IterFuncByRankRange(start int, end int, fn func(key string, value interface{}) bool) {
	if fn == nil {
		return
	}

	start, end, reverse := set.sanitizeIndexes(start, end)
	traversed, x, _ := set.findNodeByRank(start, false)
	var nodes []*Node

	x = x.level[0].forward
	for x != nil && traversed < end {
		next := x.level[0].forward

		if reverse {
			nodes = append(nodes, x)
		} else if !fn(x.key, x.Value) {
			return
		}

		traversed++
		x = next
	}

	if reverse {
		for i := len(nodes) - 1; i >= 0; i-- {
			if !fn(nodes[i].key, nodes[i].Value) {
				return
			}
		}
	}
}
