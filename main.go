package main

import (
	"fmt"
	"hyblockchain/crypto/secp256k1"
	"hyblockchain/kvstore/leveldb"
	"hyblockchain/mpt"
	"hyblockchain/statdb"
	"hyblockchain/txpool"
	"hyblockchain/types"
	"os"
)

func main() {
	fmt.Println("=== 区块链组件综合测试 ===")

	// 1. 测试 LevelDB
	fmt.Println("\n1. 测试 LevelDB 存储...")
	testLevelDB()

	// 2. 测试 MPT (Merkle Patricia Trie)
	fmt.Println("\n2. 测试 MPT 状态树...")
	testMPT()

	// 3. 测试交易池
	fmt.Println("\n3. 测试交易池...")
	testTxPool()

	fmt.Println("\n=== 所有测试完成 ===")
}

func testLevelDB() {
	// 创建临时数据库目录
	dbPath := "./testdb"
	defer os.RemoveAll(dbPath) // 测试完成后清理

	// 创建 LevelDB 实例
	db, err := leveldb.NewLevelDB(dbPath)
	if err != nil {
		fmt.Printf("创建 LevelDB 失败: %v\n", err)
		return
	}
	defer db.Close()

	// 测试基本操作
	fmt.Println("  - 测试 Put/Get 操作...")
	err = db.Put([]byte("key1"), []byte("value1"))
	if err != nil {
		fmt.Printf("Put 失败: %v\n", err)
		return
	}

	value, err := db.Get([]byte("key1"))
	if err != nil {
		fmt.Printf("Get 失败: %v\n", err)
		return
	}
	fmt.Printf("    获取到值: %s\n", string(value))

	// 测试批量操作
	fmt.Println("  - 测试批量操作...")
	batch := db.Batch()
	batch.Put([]byte("key2"), []byte("value2"))
	batch.Put([]byte("key3"), []byte("value3"))
	batch.Delete([]byte("key1"))

	err = db.Write(batch)
	if err != nil {
		fmt.Printf("批量写入失败: %v\n", err)
		return
	}

	// 验证批量操作结果
	value2, _ := db.Get([]byte("key2"))
	value3, _ := db.Get([]byte("key3"))
	_, err = db.Get([]byte("key1"))

	fmt.Printf("    key2: %s, key3: %s, key1已删除: %v\n",
		string(value2), string(value3), err != nil)

	// 测试迭代器
	fmt.Println("  - 测试迭代器...")
	iter := db.NewIterator([]byte("key"))
	defer iter.Release()

	count := 0
	for iter.Next() {
		fmt.Printf("    迭代器: %s -> %s\n", string(iter.Key()), string(iter.Value()))
		count++
	}
	fmt.Printf("    总共迭代到 %d 个键值对\n", count)
}

func testMPT() {
	// 创建临时数据库目录
	dbPath := "./mpt_testdb"
	defer os.RemoveAll(dbPath)

	// 创建 LevelDB 作为 MPT 的底层存储
	db, err := leveldb.NewLevelDB(dbPath)
	if err != nil {
		fmt.Printf("创建 LevelDB 失败: %v\n", err)
		return
	}
	defer db.Close()

	// 创建 MPT
	mptTree := mpt.NewMPT(db)

	// 测试插入操作
	fmt.Println("  - 测试插入键值对...")
	testData := map[string]string{
		"account1":  "100",
		"account2":  "200",
		"account3":  "300",
		"contract1": "0x1234567890abcdef",
		"contract2": "0xfedcba0987654321",
	}

	for key, value := range testData {
		err := mptTree.Put([]byte(key), []byte(value))
		if err != nil {
			fmt.Printf("插入 %s 失败: %v\n", key, err)
			return
		}
		fmt.Printf("    插入: %s -> %s\n", key, value)
	}

	// 获取根哈希
	rootHash := mptTree.RootHash()
	fmt.Printf("  - MPT 根哈希: %x\n", rootHash)

	// 测试查询操作
	fmt.Println("  - 测试查询操作...")
	for key, expectedValue := range testData {
		value, err := mptTree.Get([]byte(key))
		if err != nil {
			fmt.Printf("查询 %s 失败: %v\n", key, err)
			continue
		}
		fmt.Printf("    查询: %s -> %s (期望: %s)\n", key, string(value), expectedValue)
	}

	// 测试删除操作
	fmt.Println("  - 测试删除操作...")
	err = mptTree.Delete([]byte("account2"))
	if err != nil {
		fmt.Printf("删除 account2 失败: %v\n", err)
		return
	}
	fmt.Println("    删除了 account2")

	// 验证删除结果
	_, err = mptTree.Get([]byte("account2"))
	if err != nil {
		fmt.Println("    account2 已成功删除")
	}

	// 获取新的根哈希
	newRootHash := mptTree.RootHash()
	fmt.Printf("  - 删除后的根哈希: %x\n", newRootHash)
}

func PrintPoolStatus(pool *txpool.DefaultPool) {
	fmt.Println("=== 交易池状态 ===")
	fmt.Println("Pending 交易:")
	for addr, blkList := range pool.Pendings {
		for i, blk := range blkList {
			fmt.Printf("  地址 %x 第 %d 个 pending 块:\n", addr, i)
			if dblk, ok := blk.(*txpool.DefaultSortedTxs); ok {
				for _, tx := range *dblk {
					fmt.Printf("    from=%x nonce=%d gasPrice=%d\n", tx.From(), tx.Nonce(), tx.GasPrice())
				}
			}
		}
	}
	fmt.Println("Queue 交易:")
	for addr, txs := range pool.Queued {
		fmt.Printf("  地址 %x queue:\n", addr)
		for _, tx := range txs {
			fmt.Printf("    from=%x nonce=%d gasPrice=%d\n", tx.From(), tx.Nonce(), tx.GasPrice())
		}
	}
	fmt.Println("=== End ===")
}

func testTxPool() {
	statDB := statdb.NewMockStatDB()
	pool := txpool.NewDefaultPool(statDB)

	priv1, _ := secp256k1.GenerateKey()
	priv2, _ := secp256k1.GenerateKey()
	addr1 := types.PubKeyToAddress(secp256k1.PubKeyFromPrivKey(priv1))
	addr2 := types.PubKeyToAddress(secp256k1.PubKeyFromPrivKey(priv2))

	// 1. 简单添加 + Pop
	tx1, _ := types.NewTransactionWithSigner(addr1, 1, 21000, 100, 10, priv1)
	pool.NewTx(tx1)
	pop1 := pool.Pop()
	fmt.Printf("[1] Pop: nonce=%d, gasPrice=%d\n", pop1.Nonce(), pop1.GasPrice())
	PrintPoolStatus(pool) // 打印状态

	// 2. Nonce 跳跃 → Queue
	tx2, _ := types.NewTransactionWithSigner(addr1, 3, 21000, 100, 20, priv1)
	pool.NewTx(tx2)
	if pool.Pop() == nil {
		fmt.Println("[2] 不能 Pop queue 中 tx2（未解锁）")
	}
	PrintPoolStatus(pool)

	// 3. 补齐 nonce 解锁 queue
	tx3, _ := types.NewTransactionWithSigner(addr1, 2, 21000, 100, 15, priv1)
	pool.NewTx(tx3)

	for i := 0; i < 2; i++ {
		tx := pool.Pop()
		if tx == nil {
			break
		}
		fmt.Printf("[3] Pop: nonce=%d, gasPrice=%d\n", tx.Nonce(), tx.GasPrice())
	}
	PrintPoolStatus(pool)

	// 4. 替换交易（高 Gas）
	txLow, _ := types.NewTransactionWithSigner(addr2, 1, 21000, 100, 10, priv2)
	txHigh, _ := types.NewTransactionWithSigner(addr2, 1, 21000, 100, 50, priv2)
	pool.NewTx(txLow)
	pool.NewTx(txHigh)
	pop := pool.Pop()
	fmt.Printf("[4] Pop: nonce=%d, gasPrice=%d\n", pop.Nonce(), pop.GasPrice())
	PrintPoolStatus(pool)

	// 5. 多地址混合交易
	txA1, _ := types.NewTransactionWithSigner(addr1, 4, 21000, 100, 20, priv1)
	txA2, _ := types.NewTransactionWithSigner(addr1, 5, 21000, 100, 25, priv1)
	txB1, _ := types.NewTransactionWithSigner(addr2, 2, 21000, 100, 15, priv2)
	pool.NewTx(txA1)
	pool.NewTx(txA2)
	pool.NewTx(txB1)

	for i := 0; i < 3; i++ {
		tx := pool.Pop()
		if tx == nil {
			break
		}
		fmt.Printf("[5] Pop混合: from=%x, nonce=%d, gasPrice=%d\n", tx.From(), tx.Nonce(), tx.GasPrice())
	}
	PrintPoolStatus(pool)

	// 6. nonce 回退 or gas price 太低
	txInvalidNonce, _ := types.NewTransactionWithSigner(addr2, 2, 21000, 100, 30, priv2)
	pool.NewTx(txInvalidNonce)

	txWeak, _ := types.NewTransactionWithSigner(addr2, 3, 21000, 100, 10, priv2)
	txStrong, _ := types.NewTransactionWithSigner(addr2, 3, 21000, 100, 50, priv2)
	pool.NewTx(txStrong)
	pool.NewTx(txWeak)
	pop = pool.Pop()
	if pop == nil {
		fmt.Println("[6] Pop: nil，pending 为空或 queue 未解锁")
	} else {
		fmt.Printf("[6] Pop: nonce=%d, gasPrice=%d\n", pop.Nonce(), pop.GasPrice())
	}
	PrintPoolStatus(pool)
}
