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
	first := sorted[0]
	return first.GasPrice()
}

func (sorted DefaultSortedTxs) Push(tx *types.Transaction) {
	sorted = append(sorted, tx)
	//对gas费高的排序
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GasPrice() > sorted[j].GasPrice()
	})
}

func (sorted DefaultSortedTxs) Pop() *types.Transaction {
	if len(sorted) == 0 {
		return nil
	}
	tx := sorted[0]
	copy(sorted[1:], sorted[1:])
	sorted = sorted[:len(sorted)-1]
	return tx
}

func (sorted DefaultSortedTxs) Nonce() uint64 {
	return sorted[len(sorted)-1].Nonce()
}

func (sorted DefaultSortedTxs) Replace(tx *types.Transaction) {
	for i, t := range sorted {
		if t.Nonce() == tx.Nonce() {
			if tx.GasPrice() > t.GasPrice() {
				sorted[i] = tx
			}
			return // 不管是否替换，都 return，避免重复添加
		}
	}
	// 没找到，append 并按 nonce 升序排序
	sorted = append(sorted, tx)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Nonce() < sorted[j].Nonce()
	})
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
	nonce := account.Nonce
	blks := pool.pendings[tx.From()]
	if len(blks) > 0 {
		last := blks[len(blks)-1]
		nonce = last.Nonce()
	}
	if tx.Nonce() > nonce+1 {
		pool.addQueue(tx)
	} else if tx.Nonce() == nonce+1 {
		//push
		pool.pushpending(blks, tx)
	} else {
		//替换
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

func (pool DefaultPool) pushpending(blks []SortedTxs, tx *types.Transaction) {
	if len(blks) == 0 {
		blk := make(DefaultSortedTxs, 0)
		blk = append(blk, tx)
		blks = append(blks, blk)
		pool.pendings[tx.From()] = blks
		pool.txs = append(pool.txs, blk)
		sort.Sort(pool.txs)
	} else {
		last := blks[len(blks)-1]
		if last.GasPrice() <= tx.GasPrice() {
			last.Push(tx)
		} else {
			blk := make(DefaultSortedTxs, 0)
			blk = append(blk, tx)
			blks = append(blks, blk)
			pool.pendings[tx.From()] = blks
			pool.txs = append(pool.txs, blk)
			sort.Sort(pool.txs)
		}
	}
}

func (pool *DefaultPool) Pop() *types.Transaction {
	if len(pool.txs) == 0 {
		return nil
	}
	// 取全池中 gasPrice 最高的 DefaultSortedTxs
	block := pool.txs[len(pool.txs)-1]
	tx := block.Pop()

	// 如果 block 为空，移除
	if len(*block.(*DefaultSortedTxs)) == 0 {
		pool.txs = pool.txs[:len(pool.txs)-1]
	}

	return tx
}

func (pool *DefaultPool) SetStatRoot(root hash.Hash) {
	pool.StatDB.SetRoot(root)
}

func NewDefaultPool(stat statdb.StatDB) *DefaultPool {
	return &DefaultPool{
		StatDB:   stat,
		all:      make(map[hash.Hash]bool),
		txs:      make(pendingTxs, 0),
		pendings: make(map[types.Address][]SortedTxs),
		queued:   make(map[types.Address][]*types.Transaction),
	}
}
