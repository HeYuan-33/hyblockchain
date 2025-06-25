package txpool

import (
	"hash"
	"hyblockchain/statdb"
	"hyblockchain/types"
	"sort"
)

type SortedTxs interface {
	GasPrice() uint64
	Push(tx *types.Transaction)
	Replace(tx *types.Transaction)
	Pop() *types.Transaction
	Nonce() uint64
}

type DefaultSortedTxs []*types.Transaction
type pendingTxs []SortedTxs

func (sorted DefaultSortedTxs) GasPrice() uint64 {
	if len(sorted) == 0 {
		return 0
	}
	return sorted[0].GasPrice()
}

func (sorted *DefaultSortedTxs) Push(tx *types.Transaction) {
	*sorted = append(*sorted, tx)
	// 按 nonce 升序排序，保证同一地址内交易按 nonce 执行
	sort.Slice(*sorted, func(i, j int) bool {
		return (*sorted)[i].Nonce() < (*sorted)[j].Nonce()
	})
}

func (sorted *DefaultSortedTxs) Pop() *types.Transaction {
	if len(*sorted) == 0 {
		return nil
	}
	tx := (*sorted)[0]
	*sorted = (*sorted)[1:]
	return tx
}

func (sorted DefaultSortedTxs) Nonce() uint64 {
	return sorted[len(sorted)-1].Nonce()
}

func (sorted *DefaultSortedTxs) Replace(tx *types.Transaction) {
	for i, t := range *sorted {
		if t.Nonce() == tx.Nonce() {
			if tx.GasPrice() > t.GasPrice() {
				(*sorted)[i] = tx
				sort.Slice(*sorted, func(i, j int) bool {
					return (*sorted)[i].GasPrice() > (*sorted)[j].GasPrice()
				})
			}
			return
		}
	}
}

func (p pendingTxs) Len() int { return len(p) }

func (p pendingTxs) Less(i, j int) bool {
	return p[i].GasPrice() < p[j].GasPrice()
}

func (p pendingTxs) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type DefaultPool struct {
	StatDB   statdb.StatDB
	all      map[hash.Hash]bool
	txs      pendingTxs
	pendings map[types.Address][]SortedTxs
	queued   map[types.Address][]*types.Transaction
}

func (pool *DefaultPool) NewTx(tx *types.Transaction) {
	account := pool.StatDB.Load(tx.From())
	if account.Nonce >= tx.Nonce() {
		return
	}
	blks := pool.pendings[tx.From()]
	expectedNonce := account.Nonce + 1
	for _, blk := range blks {
		expectedNonce = blk.Nonce() + 1
	}

	if tx.Nonce() > expectedNonce {
		pool.addQueue(tx)
	} else if tx.Nonce() == expectedNonce {
		pool.pushpending(blks, tx)
	} else {
		pool.replacePending(blks, tx)
	}
}

/*
	下面是复杂的代码
*/
//	if len(blks) == 0 {
//		if tx.Nonce == account.Nonce+1 {
//			// pending
//		} else {
//			//quene
//			pool.addQueue(tx)
//		}
//	} else {
//		last := blks[len(blks)-1]
//		if tx.Nonce > last.Nonce()+1 {
//			//quene
//			pool.addQueue(tx)
//		} else {
//			//pengding
//		}
//	}
//}

func (pool DefaultPool) replacePending(blks []SortedTxs, tx *types.Transaction) {
	for _, blk := range blks {
		if blk.Nonce() >= tx.Nonce() {
			//替换
			blk.Replace(tx)
			break
		}
	}
}

func (pool DefaultPool) addQueue(tx *types.Transaction) {
	list := pool.queued[tx.From()]
	list = append(list, tx)
	//sort
	sort.Slice(list, func(i, j int) bool {
		return list[i].Nonce() < list[j].Nonce()
	})
	pool.queued[tx.From()] = list
}

func (pool *DefaultPool) pushpending(blks []SortedTxs, tx *types.Transaction) {
	if len(blks) == 0 {
		blk := make(DefaultSortedTxs, 0)
		blk = append(blk, tx)
		pool.pendings[tx.From()] = []SortedTxs{&blk}
		pool.txs = append(pool.txs, &blk)
		sort.Sort(pool.txs)
		return
	}

	for _, blk := range blks {
		for _, t := range *blk.(*DefaultSortedTxs) {
			if t.Nonce() == tx.Nonce() {
				blk.Replace(tx)
				sort.Sort(pool.txs) // 替换后重新排序池内所有块
				return
			}
		}
	}

	last := blks[len(blks)-1]
	if last.GasPrice() <= tx.GasPrice() {
		last.Push(tx)
		sort.Sort(pool.txs)
	} else {
		blk := make(DefaultSortedTxs, 0)
		blk = append(blk, tx)
		blks = append(blks, &blk)
		pool.pendings[tx.From()] = blks
		pool.txs = append(pool.txs, &blk)
		sort.Sort(pool.txs)
	}
}

func (pool *DefaultPool) Pop() *types.Transaction {
	if len(pool.txs) == 0 {
		return nil
	}
	block := pool.txs[len(pool.txs)-1]
	tx := block.Pop()

	if len(*block.(*DefaultSortedTxs)) == 0 {
		pool.txs = pool.txs[:len(pool.txs)-1]
	}

	if tx != nil {
		// 更新账户nonce，表示该交易已执行
		account := pool.StatDB.Load(tx.From())
		if tx.Nonce() > account.Nonce {
			account.Nonce = tx.Nonce()
			pool.StatDB.Store(tx.From(), account)
		}

		pool.promoteQueued(tx.From())
	}

	return tx
}

func (pool *DefaultPool) promoteQueued(addr types.Address) {
	queuedTxs, ok := pool.queued[addr]
	if !ok || len(queuedTxs) == 0 {
		return
	}

	// 先按 nonce 排序
	sort.Slice(queuedTxs, func(i, j int) bool {
		return queuedTxs[i].Nonce() < queuedTxs[j].Nonce()
	})

	account := pool.StatDB.Load(addr)
	expectedNonce := account.Nonce + 1

	var promoteTxs []*types.Transaction
	var remainTxs []*types.Transaction

	for _, tx := range queuedTxs {
		if tx.Nonce() == expectedNonce {
			promoteTxs = append(promoteTxs, tx)
			expectedNonce++
		} else {
			remainTxs = append(remainTxs, tx)
		}
	}

	if len(remainTxs) == 0 {
		delete(pool.queued, addr)
	} else {
		pool.queued[addr] = remainTxs
	}

	if len(promoteTxs) == 0 {
		return
	}

	// 这里推送到 pendings，推荐只用一个区块维护同一地址的连续交易
	blks := pool.pendings[addr]
	var blk *DefaultSortedTxs
	if len(blks) == 0 {
		newBlk := make(DefaultSortedTxs, 0, len(promoteTxs))
		blk = &newBlk
		pool.pendings[addr] = []SortedTxs{blk}
		pool.txs = append(pool.txs, blk)
	} else {
		blk = blks[0].(*DefaultSortedTxs)
	}

	for _, tx := range promoteTxs {
		blk.Push(tx)
	}

	sort.Sort(pool.txs)
}

func (pool *DefaultPool) SetStatRoot(root hash.Hash) {
	pool.StatDB.SetRoot(root)
}

// NewDefaultPool 创建一个默认的交易池
func NewDefaultPool(stat *statdb.MockStatDB) *DefaultPool {
	return &DefaultPool{
		StatDB:   stat,
		all:      make(map[hash.Hash]bool),
		txs:      make([]SortedTxs, 0),
		pendings: make(map[types.Address][]SortedTxs),
		queued:   make(map[types.Address][]*types.Transaction),
	}
}
