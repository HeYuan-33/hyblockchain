package statdb

import (
	"hash"
	"hyblockchain/types"
)

// MockStatDB 是一个内存实现的 StatDB 模拟对象
type MockStatDB struct {
	accounts map[types.Address]types.Account
	root     hash.Hash
}

func NewMockStatDB() *MockStatDB {
	return &MockStatDB{
		accounts: make(map[types.Address]types.Account),
	}
}

// Load 返回指定地址的账户信息，找不到时返回空账户(Nonce=0)
func (db *MockStatDB) Load(addr types.Address) *types.Account {
	acc, ok := db.accounts[addr]
	if !ok {
		return &types.Account{}
	}
	return &acc
}

// Store 设置指定地址的账户信息
func (db *MockStatDB) Store(addr types.Address, account *types.Account) {
	if account == nil {
		delete(db.accounts, addr)
		return
	}
	db.accounts[addr] = *account
}

// SetRoot 设置状态树根哈希（模拟用）
func (db *MockStatDB) SetRoot(root hash.Hash) {
	db.root = root
}

// SetStatRoot 是 SetRoot 的别名，兼容接口
func (db *MockStatDB) SetStatRoot(root hash.Hash) {
	db.SetRoot(root)
}
