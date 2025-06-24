package mpt

// NodeType 表示MPT节点的类型
type NodeType byte

const (
	BranchNode    NodeType = 0
	ExtensionNode NodeType = 1
	LeafNode      NodeType = 2
)

// Node 表示MPT中的一个节点
type Node struct {
	Type     NodeType  // 节点类型
	Key      []byte    // 压缩路径（用于扩展/叶子节点）
	Value    []byte    // 存储值（仅叶子节点用）
	Children [16]*Node // 子节点数组（仅分支节点使用）
	Hash     []byte    // 当前节点的哈希（默认为 nil，需外部生成）
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
