package mpt

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"hyblockchain/kvstore"
)

// MPT 表示一个Merkle Patricia Trie
type MPT struct {
	root      *Node
	db        kvstore.KVStore
	emptyRoot *Node // 唯一的空节点实例，保证空树哈希稳定
}

// NewMPT 创建新的MPT，初始化空节点并递归提交数据库
func NewMPT(db kvstore.KVStore) *MPT {
	emptyRoot := NewBranchNode()
	mpt := &MPT{
		root:      emptyRoot,
		emptyRoot: emptyRoot,
		db:        db,
	}
	// 提交空节点到数据库，保证空节点hash和数据存在
	mpt.commitNode(emptyRoot)
	return mpt
}

// Put 在MPT中存储键值对
func (m *MPT) Put(key, value []byte) error {
	nibbles := bytesToNibbles(key)
	newRoot, err := m.insert(m.root, nibbles, value)
	if err != nil {
		return err
	}
	m.root = newRoot
	return m.commit()
}

// Get 从MPT中获取值
func (m *MPT) Get(key []byte) ([]byte, error) {
	nibbles := bytesToNibbles(key)
	return m.get(m.root, nibbles)
}

// Delete 从MPT中删除键值对
func (m *MPT) Delete(key []byte) error {
	nibbles := bytesToNibbles(key)
	newRoot, err := m.delete(m.root, nibbles)
	if err != nil {
		return err
	}
	if newRoot == nil || newRoot == m.emptyRoot {
		// 空节点统一使用唯一实例
		newRoot = m.emptyRoot
	}
	m.root = newRoot
	return m.commit()
}

// RootHash 获取MPT的根哈希
func (m *MPT) RootHash() []byte {
	if m.root == nil {
		return nil
	}
	return m.root.Hash
}

// insert 在MPT中插入或更新节点
func (m *MPT) insert(n *Node, key []byte, value []byte) (*Node, error) {
	if n == nil || n == m.emptyRoot {
		return NewLeafNode(key, value), nil
	}

	switch n.Type {
	case LeafNode:
		common := commonPrefix(n.Key, key)
		if common == len(n.Key) && common == len(key) {
			n.Value = value
			return n, nil
		}
		branch := NewBranchNode()
		if common == 0 {
			if len(n.Key) > 0 {
				branch.Children[n.Key[0]] = NewLeafNode(n.Key[1:], n.Value)
			} else {
				branch.Value = n.Value
			}
			if len(key) > 0 {
				branch.Children[key[0]] = NewLeafNode(key[1:], value)
			} else {
				branch.Value = value
			}
			return branch, nil
		}

		extension := NewExtensionNode(key[:common], branch)
		if common < len(n.Key) {
			branch.Children[n.Key[common]] = NewLeafNode(n.Key[common+1:], n.Value)
		} else {
			branch.Value = n.Value
		}
		if common < len(key) {
			branch.Children[key[common]] = NewLeafNode(key[common+1:], value)
		} else {
			branch.Value = value
		}
		return extension, nil

	case ExtensionNode:
		common := commonPrefix(n.Key, key)
		if common == len(n.Key) {
			child, err := m.insert(n.Children[0], key[common:], value)
			if err != nil {
				return nil, err
			}
			n.Children[0] = child
			return n, nil
		}

		branch := NewBranchNode()
		if common < len(n.Key) {
			child, err := m.insert(n.Children[0], n.Key[common+1:], n.Children[0].Value)
			if err != nil {
				return nil, err
			}
			branch.Children[n.Key[common]] = child
		} else {
			branch.Value = n.Children[0].Value
		}

		if common < len(key) {
			branch.Children[key[common]] = NewLeafNode(key[common+1:], value)
		} else {
			branch.Value = value
		}

		if common == 0 {
			return branch, nil
		}
		return NewExtensionNode(key[:common], branch), nil

	case BranchNode:
		if len(key) == 0 {
			n.Value = value
			return n, nil
		}
		child, err := m.insert(n.Children[key[0]], key[1:], value)
		if err != nil {
			return nil, err
		}
		n.Children[key[0]] = child
		return n, nil
	}

	return nil, errors.New("unknown node type")
}

// get 从MPT中获取值
func (m *MPT) get(n *Node, key []byte) ([]byte, error) {
	if n == nil || n == m.emptyRoot {
		return nil, errors.New("key not found")
	}

	switch n.Type {
	case LeafNode:
		if bytes.Equal(n.Key, key) {
			return n.Value, nil
		}
		return nil, errors.New("key not found")

	case ExtensionNode:
		if len(key) < len(n.Key) || !bytes.Equal(n.Key, key[:len(n.Key)]) {
			return nil, errors.New("key not found")
		}
		return m.get(n.Children[0], key[len(n.Key):])

	case BranchNode:
		if len(key) == 0 {
			if n.Value == nil {
				return nil, errors.New("key not found")
			}
			return n.Value, nil
		}
		return m.get(n.Children[key[0]], key[1:])
	}

	return nil, errors.New("unknown node type")
}

// delete 从MPT中删除节点，空节点统一返回 m.emptyRoot
func (m *MPT) delete(n *Node, key []byte) (*Node, error) {
	if n == nil || n == m.emptyRoot {
		return m.emptyRoot, nil
	}

	switch n.Type {
	case LeafNode:
		if bytes.Equal(n.Key, key) {
			return m.emptyRoot, nil
		}
		return n, nil

	case ExtensionNode:
		if len(key) < len(n.Key) || !bytes.Equal(n.Key, key[:len(n.Key)]) {
			return n, nil
		}
		child, err := m.delete(n.Children[0], key[len(n.Key):])
		if err != nil {
			return nil, err
		}
		if child == m.emptyRoot {
			return m.emptyRoot, nil
		}
		if child.Type == ExtensionNode {
			return NewExtensionNode(append(n.Key, child.Key...), child.Children[0]), nil
		}
		return NewExtensionNode(n.Key, child), nil

	case BranchNode:
		if len(key) == 0 {
			n.Value = nil
		} else {
			child, err := m.delete(n.Children[key[0]], key[1:])
			if err != nil {
				return nil, err
			}
			n.Children[key[0]] = child
		}

		nonNilChildren := 0
		lastChildIndex := -1
		for i, child := range n.Children {
			if child != nil && child != m.emptyRoot {
				nonNilChildren++
				lastChildIndex = i
			}
		}

		if nonNilChildren == 0 && n.Value == nil {
			return m.emptyRoot, nil
		}

		if nonNilChildren == 1 && n.Value == nil {
			child := n.Children[lastChildIndex]
			if child.Type == ExtensionNode {
				return NewExtensionNode(append([]byte{byte(lastChildIndex)}, child.Key...), child.Children[0]), nil
			}
			return NewLeafNode(append([]byte{byte(lastChildIndex)}, child.Key...), child.Value), nil
		}
		return n, nil
	}

	return nil, errors.New("unknown node type")
}

// commit 将MPT的更改提交到数据库，先递归提交子节点，保证哈希更新
func (m *MPT) commit() error {
	return m.commitNode(m.root)
}

// commitNode 递归提交节点到数据库
func (m *MPT) commitNode(n *Node) error {
	if n == nil {
		return nil
	}

	// 先提交所有子节点
	for _, child := range n.Children {
		if child != nil {
			if err := m.commitNode(child); err != nil {
				return err
			}
		}
	}

	// 计算当前节点的哈希并存储
	n.Hash = m.hashNode(n)
	return m.db.Put(n.Hash, m.serializeNode(n))
}

// hashNode 计算节点的哈希
func (m *MPT) hashNode(n *Node) []byte {
	data := m.serializeNode(n)
	hash := sha256.Sum256(data)
	return hash[:]
}

// serializeNode 序列化节点，保证顺序和格式稳定
func (m *MPT) serializeNode(n *Node) []byte {
	var buf bytes.Buffer
	buf.WriteByte(byte(n.Type))

	switch n.Type {
	case LeafNode:
		buf.Write(n.Key)
		buf.Write(n.Value)

	case ExtensionNode:
		buf.Write(n.Key)
		if len(n.Children) > 0 && n.Children[0] != nil {
			buf.Write(n.Children[0].Hash)
		} else {
			buf.Write(make([]byte, 32))
		}

	case BranchNode:
		for _, child := range n.Children {
			if child != nil {
				buf.Write(child.Hash)
			} else {
				buf.Write(make([]byte, 32))
			}
		}
		if n.Value != nil {
			buf.Write(n.Value)
		} else {
			buf.Write([]byte{})
		}
	}

	return buf.Bytes()
}

// bytesToNibbles 将字节数组转换成nibbles
func bytesToNibbles(b []byte) []byte {
	nibbles := make([]byte, len(b)*2)
	for i, v := range b {
		nibbles[i*2] = v >> 4
		nibbles[i*2+1] = v & 0x0f
	}
	return nibbles
}

// commonPrefix 计算两个字节数组的最长共同前缀长度
func commonPrefix(a, b []byte) int {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	return i
}
