package db

import (
	"database/sql"
	"fmt"

	// pq postgresql driver
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	blocksmodel "github.com/buidl-labs/filecoin-chain-indexer/model/blocks"
	marketmodel "github.com/buidl-labs/filecoin-chain-indexer/model/market"
	messagemodel "github.com/buidl-labs/filecoin-chain-indexer/model/messages"
	minermodel "github.com/buidl-labs/filecoin-chain-indexer/model/miner"
)

type Store struct {
	db *sql.DB
}

func New(connStr string) (*Store, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &Store{
		db: db,
	}, nil
}

// Conn returns the underlying database connection
func (s *Store) Conn() (*sql.DB, error) {
	return s.db, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	conn := s.db
	return conn.Close()
}

func (s *Store) PersistTransactions(t messagemodel.Transaction) error {
	statement := `INSERT INTO transactions (cid, fromAddr, toAddr, amount) VALUES ($1, $2, $3, $4)`
	err := s.db.QueryRow(statement, t.Cid, t.FromAddr, t.ToAddr, t.Amount)
	if err != nil {
		log.Errorln("Error in inserting Transaction", t.Cid, err)
		return fmt.Errorf("insert transaction")
	}
	return nil
}

func (s *Store) PersistBlockHeaders(bh blocksmodel.BlockHeader) error {
	return nil
}

func (s *Store) PersistMinerInfos(mi minermodel.MinerInfo) error {
	return nil
}

func (s *Store) PersistMinerFunds(mf minermodel.MinerFunds) error {
	return nil
}

func (s *Store) PersistMinerDeadlines(md minermodel.MinerCurrentDeadlineInfo) error {
	return nil
}

func (s *Store) PersistMinerQuality(mq minermodel.MinerQuality) error {
	return nil
}

func (s *Store) PersistMinerSectors(msi minermodel.MinerSectorInfo) error {
	return nil
}

func (s *Store) PersistMinerSectorDeals(msd minermodel.MinerSectorDeal) error {
	return nil
}

func (s *Store) PersistMinerSectorEvents(mse minermodel.MinerSectorEvent) error {
	return nil
}

func (s *Store) PersistMinerSectorPosts(msp minermodel.MinerSectorPost) error {
	return nil
}

func (s *Store) PersistMarketDealProposals(mdp marketmodel.MarketDealProposal) error {
	return nil
}
