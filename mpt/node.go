package mpt

// NodeType 表示MPT节点的类型
type NodeType byte

const (
	BranchNode   NodeType = 0
	ExtensionNode NodeType = 1
	LeafNode     NodeType = 2
)

// Node 表示MPT中的一个节点
type Node struct {
	Type     NodeType
	Key      []byte
	Value    []byte
	Children [16]*Node
	Hash     []byte
}

// NewBranchNode 创建一个新的分支节点
func NewBranchNode() *Node {
	return &Node{
		Type:     BranchNode,
		Children: [16]*Node{},
	}
}

// NewExtensionNode 创建一个新的扩展节点
func NewExtensionNode(key []byte, child *Node) *Node {
	return &Node{
		Type:     ExtensionNode,
		Key:      key,
		Children: [16]*Node{0: child},
	}
}

// NewLeafNode 创建一个新的叶子节点
func NewLeafNode(key []byte, value []byte) *Node {
	return &Node{
		Type:  LeafNode,
		Key:   key,
		Value: value,
	}
}
