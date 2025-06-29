package txpool

import (
	"fmt"
	"sort"
	"testing"

	"hyblockchain/crypto/secp256k1"
	"hyblockchain/statdb"
	"hyblockchain/types"
)

// 格式化地址为 hex 字符串
func formatAddr(addr types.Address) string {
	return fmt.Sprintf("0x%x", addr[:])
}

// 打印交易池状态（排序 nonce）
func (pool *DefaultPool) PrintState(t *testing.T) {
	t.Log("== Pending Transactions ==")
	for addr, blks := range pool.Pendings {
		t.Logf("  [%s]", formatAddr(addr))

		for i, blk := range blks {
			txs := *blk.(*DefaultSortedTxs)

			// ✅ 排序 nonce
			sort.Slice(txs, func(i, j int) bool {
				return txs[i].Nonce() < txs[j].Nonce()
			})

			for _, tx := range txs {
				t.Logf("    Block[%d]: nonce=%d, gasPrice=%d", i, tx.Nonce(), tx.GasPrice())
			}
		}
	}

	t.Log("== Queued Transactions ==")
	for addr, txs := range pool.Queued {
		t.Logf("  [%s]", formatAddr(addr))

		// ✅ 同样排序 nonce
		sort.Slice(txs, func(i, j int) bool {
			return txs[i].Nonce() < txs[j].Nonce()
		})

		for _, tx := range txs {
			t.Logf("    Queued: nonce=%d, gasPrice=%d", tx.Nonce(), tx.GasPrice())
		}
	}
}

func TestTxPoolAdvancedMoreCases(t *testing.T) {
	priv1, _ := secp256k1.GenerateKey()
	pub1 := secp256k1.PubKeyFromPrivKey(priv1)
	addr1 := types.PubKeyToAddress(pub1)

	priv2, _ := secp256k1.GenerateKey()
	pub2 := secp256k1.PubKeyFromPrivKey(priv2)
	addr2 := types.PubKeyToAddress(pub2)

	db := statdb.NewMockStatDB()
	db.Store(addr1, &types.Account{Amount: 100, Nonce: 0})
	db.Store(addr2, &types.Account{Amount: 100, Nonce: 0})

	pool := NewDefaultPool(db)

	// 1. addr1 连续nonce，gasPrice递减
	txA1, _ := types.NewTransactionWithSigner(addr1, 1, 21000, 10, 12, priv1)
	txA2, _ := types.NewTransactionWithSigner(addr1, 2, 21000, 9, 9, priv1)
	txA3, _ := types.NewTransactionWithSigner(addr1, 3, 21000, 8, 8, priv1)

	pool.NewTx(txA1)
	pool.NewTx(txA2)
	pool.NewTx(txA3)
	t.Log("Added addr1 continuous nonce transactions with descending gasPrice")

	// 2. addr1 加入跳nonce交易（queued）
	txA5, _ := types.NewTransactionWithSigner(addr1, 5, 21000, 7, 7, priv1)
	pool.NewTx(txA5)
	t.Log("Added addr1 tx nonce=5 (queued due to nonce gap)")

	// 3. addr2 连续nonce加入
	txB2, _ := types.NewTransactionWithSigner(addr2, 1, 21000, 15, 11, priv2)
	txB1, _ := types.NewTransactionWithSigner(addr2, 2, 21000, 20, 10, priv2)
	pool.NewTx(txB2)
	pool.NewTx(txB1)
	t.Log("Added addr2 txs nonce=2 and nonce=1 out of order")

	// 4. addr2 加入高gasPrice交易nonce=3，应该晋升
	txB3, _ := types.NewTransactionWithSigner(addr2, 3, 21000, 25, 7, priv2)
	pool.NewTx(txB3)
	t.Log("Added addr2 tx nonce=3 with high gasPrice")
	txB4, _ := types.NewTransactionWithSigner(addr2, 6, 21000, 25, 19, priv2)
	pool.NewTx(txB4)
	// 5. addr1 弹出交易，测试晋升机制
	t.Log("=== 状态 before Pop ===")
	pool.PrintState(t)

	// 弹出交易直到空
	for i := 1; i <= 10; i++ {
		tx := pool.Pop()
		if tx == nil {
			t.Logf("Pop %d: nil (no more transactions)", i)
			break
		}
		t.Logf("Pop %d: [%s] nonce=%d, gasPrice=%d", i, formatAddr(tx.From()), tx.Nonce(), tx.GasPrice())
	}

	t.Log("=== 状态 after Pop ===")
	pool.PrintState(t)

	// 6. 测试 Pop 空池行为
	txNil := pool.Pop()
	if txNil != nil {
		t.Errorf("Expected nil on empty pool Pop, got %v", txNil)
	} else {
		t.Log("Pop on empty pool correctly returns nil")
	}
}
