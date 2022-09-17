package sortedset

type Level struct {
	forward *Node
	span    int // the number of node between the current node to the forward node
}

// Node in skip list
type Node struct {
	key      string      // unique key of this node
	Value    interface{} // associated data
	score    float64     // score to determine the order of this node in the set
	backward *Node
	level    []Level
}

// Key func return the key of the node
func (node *Node) Key() string {
	return node.key
}

// Score func return the node of the node
func (node *Node) Score() float64 {
	return node.score
}

func (node *Node) Next() *Node {
	return node.level[0].forward
}

func (node *Node) Previous() *Node {
	return node.backward
}
