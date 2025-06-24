package maker

import (
	"hyblockchain/blockchain"
	"hyblockchain/statdb"
	"hyblockchain/statemachine"
	"hyblockchain/txpool"
	"hyblockchain/types"
	"math/big"
	"time"
)

type BlockProducerConfig struct {
	Duration   time.Duration
	Difficulty big.Int
	MaxTx      int64
	Coinbase   types.Address
}

type BlockProducer struct {
	txpool txpool.TxPool
	statdb statdb.StatDB
	config BlockProducerConfig

	chain blockchain.Blockchain
	m     statemachine.IMachine

	header *blockchain.Header
	block  *blockchain.Body

	interupt chan bool
}

func (producer BlockProducer) NewBlock() {
	producer.header = blockchain.NewHeader(producer.chain.Current)
	// new Body
	// producer.statdb =
}

func (producer BlockProducer) pack() {
	t := time.After(producer.config.Duration)
	for {
		select {
		case <-producer.interupt:
			break
		case <-t:
			break
		// TODO 数量
		default:
			tx := producer.txpool.Pop()
			producer.m.Execute(producer.statdb, *tx)

		}
	}
}

func (producer BlockProducer) Interupt() {
	producer.interupt <- true
}

func Seal() {

}
