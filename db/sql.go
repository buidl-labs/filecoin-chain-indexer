package db

import (
	"database/sql"
	"fmt"
	"strconv"

	// pq postgresql driver
	"github.com/go-pg/pg/v10"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	blocksmodel "github.com/buidl-labs/filecoin-chain-indexer/model/blocks"
	marketmodel "github.com/buidl-labs/filecoin-chain-indexer/model/market"
	messagemodel "github.com/buidl-labs/filecoin-chain-indexer/model/messages"
	minermodel "github.com/buidl-labs/filecoin-chain-indexer/model/miner"
	powermodel "github.com/buidl-labs/filecoin-chain-indexer/model/power"
)

type Store struct {
	db *sql.DB
	DB *pg.DB
}

func New(connStr string) (*Store, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	opt, err := pg.ParseURL(connStr)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	pgdb := pg.Connect(opt)

	return &Store{
		db: db,
		DB: pgdb,
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

func (s *Store) CloseGoPg() error {
	return s.DB.Close()
}

func (s *Store) PersistBatch(m []interface{}, modelName string) error {
	s.db.Begin()
	for _, i := range m {
		fmt.Println(i)
		switch modelName {
		// case "txn":
		// 	if txn, ok := i.(messagemodel.Transaction); ok {
		// 		s.PersistTransactions(txn)
		// 	} else {
		// 		fmt.Println(ok)
		// 	}
		// case "miner_info":
		// 	if minerInfo, ok := i.(minermodel.MinerInfo); ok {
		// 		s.PersistMinerInfos(minerInfo)
		// 	} else {
		// 		fmt.Println(ok)
		// 	}
		// case "miner_quality":
		// 	if minerQuality, ok := i.(minermodel.MinerQuality); ok {
		// 		s.PersistMinerQuality(minerQuality)
		// 	} else {
		// 		fmt.Println(ok)
		// 	}
		// case "miner_sector_info":
		// 	if minerSectorInfo, ok := i.(minermodel.MinerSectorInfo); ok {
		// 		s.PersistMinerSectors(minerSectorInfo)
		// 	} else {
		// 		fmt.Println(ok)
		// 	}
		case "miner_sector_fault":
		default:
		}
	}
	return nil
}

func (s *Store) PersistTransactions(txns []messagemodel.Transaction) error {
	if len(txns) == 0 {
		log.Info("no txns")
		return nil
	}
	query := "INSERT INTO transactions (" +
		"cid, height, sender, receiver, " +
		"amount, type, gas_fee_cap, gas_premium, " +
		"gas_limit, size_bytes, nonce, method, method_name, " +
		"state_root, exit_code, gas_used, parent_base_fee, " +
		"base_fee_burn, over_estimation_burn, miner_penalty, " +
		"miner_tip, refund, gas_refund, gas_burned, actor_name)" +
		" VALUES "
	valueArgs := []interface{}{}
	cols := 25
	c := 1
	for _, txn := range txns {
		// query += "($" + strconv.Itoa(c) + ", $" + strconv.Itoa(c+1) + ", $" + strconv.Itoa(c+2) + ", $" + strconv.Itoa(c+3) + ", $" + strconv.Itoa(c+4) + ", $" + strconv.Itoa(c+5) + ", $" + strconv.Itoa(c+6) + ", $" + strconv.Itoa(c+7) + ", $" + strconv.Itoa(c+8) + ", $" + strconv.Itoa(c+9) + ", $" + strconv.Itoa(c+10) + ", $" + strconv.Itoa(c+11) +
		// ", $" + strconv.Itoa(c+12) + ", $" + strconv.Itoa(c+13) + ", $" + strconv.Itoa(c+14) + ", $" + strconv.Itoa(c+15) + ", $" + strconv.Itoa(c+16) + ", $" + strconv.Itoa(c+17) + ", $" + strconv.Itoa(c+18) + ", $" + strconv.Itoa(c+19) + ", $" + strconv.Itoa(c+20) + ", $" + strconv.Itoa(c+21) + ", $" + strconv.Itoa(c+22) + ", $" + strconv.Itoa(c+23) + ", $" + strconv.Itoa(c+24) + "),"
		query += generateQuery(cols, c)
		valueArgs = append(valueArgs, txn.Cid)
		valueArgs = append(valueArgs, txn.Height)
		valueArgs = append(valueArgs, txn.Sender)
		valueArgs = append(valueArgs, txn.Receiver)
		valueArgs = append(valueArgs, txn.Amount)
		valueArgs = append(valueArgs, txn.Type)
		valueArgs = append(valueArgs, txn.GasFeeCap)
		valueArgs = append(valueArgs, txn.GasPremium)
		valueArgs = append(valueArgs, txn.GasLimit)
		valueArgs = append(valueArgs, txn.SizeBytes)
		valueArgs = append(valueArgs, txn.Nonce)
		valueArgs = append(valueArgs, txn.Method)
		valueArgs = append(valueArgs, txn.MethodName)
		valueArgs = append(valueArgs, txn.StateRoot)
		valueArgs = append(valueArgs, txn.ExitCode)
		valueArgs = append(valueArgs, txn.GasUsed)
		valueArgs = append(valueArgs, txn.ParentBaseFee)
		valueArgs = append(valueArgs, txn.BaseFeeBurn)
		valueArgs = append(valueArgs, txn.OverEstimationBurn)
		valueArgs = append(valueArgs, txn.MinerPenalty)
		valueArgs = append(valueArgs, txn.MinerTip)
		valueArgs = append(valueArgs, txn.Refund)
		valueArgs = append(valueArgs, txn.GasRefund)
		valueArgs = append(valueArgs, txn.GasBurned)
		valueArgs = append(valueArgs, txn.ActorName)
		c += cols
	}
	query = query[0 : len(query)-1]
	stmt, err := s.db.Prepare(query)
	if err != nil {
		log.Println("prep txns", err)
	}
	res, err := stmt.Exec(valueArgs...)
	if err != nil {
		log.Println("insert txns", err)
	}
	log.Info("res txns", res)
	return nil
}

func (s *Store) PersistBlockHeaders(bhs []blocksmodel.BlockHeader) error {
	if len(bhs) == 0 {
		log.Info("no bhs")
		return nil
	}
	query := "INSERT INTO block_headers (" +
		"height, cid, miner_id, parent_weight, parent_state_root, " +
		"parent_base_fee, fork_signaling, win_count, timestamp)" +
		" VALUES "
	valueArgs := []interface{}{}
	cols := 9
	c := 1
	for _, mi := range bhs {
		// query += "($" + strconv.Itoa(c) + ", $" + strconv.Itoa(c+1) + ", $" + strconv.Itoa(c+2) + ", $" + strconv.Itoa(c+3) + ", $" + strconv.Itoa(c+4) + ", $" + strconv.Itoa(c+5) + ", $" + strconv.Itoa(c+6) + ", $" + strconv.Itoa(c+7) + ", $" + strconv.Itoa(c+8) + "),"
		query += generateQuery(cols, c)
		valueArgs = append(valueArgs, mi.Height)
		valueArgs = append(valueArgs, mi.Cid)
		valueArgs = append(valueArgs, mi.MinerID)
		valueArgs = append(valueArgs, mi.ParentWeight)
		valueArgs = append(valueArgs, mi.ParentStateRoot)
		valueArgs = append(valueArgs, mi.ParentBaseFee)
		valueArgs = append(valueArgs, mi.ForkSignaling)
		valueArgs = append(valueArgs, mi.WinCount)
		valueArgs = append(valueArgs, mi.Timestamp)
		c += cols
	}
	query = query[0 : len(query)-1]
	stmt, err := s.db.Prepare(query)
	if err != nil {
		log.Println("prep bhs", err)
	}
	res, err := stmt.Exec(valueArgs...)
	if err != nil {
		log.Println("insert bhs", err)
	}
	log.Info("res bhs", res)
	return nil
}

func (s *Store) PersistMinerInfos(mis []*minermodel.MinerInfo) error {
	if len(mis) == 0 {
		log.Info("no minerinfos")
		return nil
	}
	query := "INSERT INTO miner_infos (" +
		"miner_id, address, peer_id, owner_id, worker_id, height, " +
		"state_root, storage_ask_price, min_piece_size, max_piece_size)" +
		" VALUES "
	valueArgs := []interface{}{}
	cols := 10
	c := 1
	for _, mi := range mis {
		// query += "($" + strconv.Itoa(c) + ", $" + strconv.Itoa(c+1) + ", $" + strconv.Itoa(c+2) + ", $" + strconv.Itoa(c+3) + ", $" + strconv.Itoa(c+4) + ", $" + strconv.Itoa(c+5) + ", $" + strconv.Itoa(c+6) + ", $" + strconv.Itoa(c+7) + ", $" + strconv.Itoa(c+8) + "),"
		query += generateQuery(cols, c)
		valueArgs = append(valueArgs, mi.MinerID)
		valueArgs = append(valueArgs, mi.Address)
		valueArgs = append(valueArgs, mi.PeerID)
		valueArgs = append(valueArgs, mi.OwnerID)
		valueArgs = append(valueArgs, mi.WorkerID)
		valueArgs = append(valueArgs, mi.Height)
		valueArgs = append(valueArgs, mi.StateRoot)
		valueArgs = append(valueArgs, mi.StorageAskPrice)
		valueArgs = append(valueArgs, mi.MinPieceSize)
		valueArgs = append(valueArgs, mi.MaxPieceSize)
		c += cols
	}
	query = query[0 : len(query)-1]
	stmt, err := s.db.Prepare(query)
	if err != nil {
		log.Println("prep minerinfos", err)
	}
	res, err := stmt.Exec(valueArgs...)
	if err != nil {
		log.Println("insert minerinfos", err)
	}
	log.Info("res minerinfos", res)
	return nil
}

func (s *Store) PersistMinerFunds(mfs []*minermodel.MinerFund) error {
	if len(mfs) == 0 {
		log.Info("no minerfunds")
		return nil
	}
	query := "INSERT INTO miner_funds (" +
		"miner_id, height, state_root, locked_funds, initial_pledge, " +
		"pre_commit_deposits, available_balance)" +
		" VALUES "
	valueArgs := []interface{}{}
	cols := 7
	c := 1
	for _, mq := range mfs {
		// query += "($" + strconv.Itoa(c) + ", $" + strconv.Itoa(c+1) + ", $" + strconv.Itoa(c+2) + ", $" + strconv.Itoa(c+3) + ", $" + strconv.Itoa(c+4) + ", $" + strconv.Itoa(c+5) + ", $" + strconv.Itoa(c+6) + "),"
		query += generateQuery(cols, c)
		valueArgs = append(valueArgs, mq.MinerID)
		valueArgs = append(valueArgs, mq.Height)
		valueArgs = append(valueArgs, mq.StateRoot)
		valueArgs = append(valueArgs, mq.LockedFunds)
		valueArgs = append(valueArgs, mq.InitialPledge)
		valueArgs = append(valueArgs, mq.PreCommitDeposits)
		valueArgs = append(valueArgs, mq.AvailableBalance)
		c += cols
	}
	query = query[0 : len(query)-1]
	stmt, err := s.db.Prepare(query)
	if err != nil {
		log.Println("prep minerfunds", err)
	}
	res, err := stmt.Exec(valueArgs...)
	if err != nil {
		log.Println("insert minerfunds", err)
	}
	log.Info("res funds", res)
	return nil
}

func (s *Store) PersistMinerDeadlines(mds []*minermodel.MinerCurrentDeadlineInfo) error {
	if len(mds) == 0 {
		log.Info("no minerdeadlines")
		return nil
	}
	query := "INSERT INTO miner_current_deadline_infos (" +
		"height, miner_id, state_root, deadline_index, period_start, " +
		"open, close, challenge, fault_cutoff)" +
		" VALUES "
	valueArgs := []interface{}{}
	cols := 9
	c := 1
	for _, mq := range mds {
		// query += "($" + strconv.Itoa(c) + ", $" + strconv.Itoa(c+1) + ", $" + strconv.Itoa(c+2) + ", $" + strconv.Itoa(c+3) + ", $" + strconv.Itoa(c+4) + ", $" + strconv.Itoa(c+5) + ", $" + strconv.Itoa(c+6) + ", $" + strconv.Itoa(c+7) + ", $" + strconv.Itoa(c+8) + "),"
		query += generateQuery(cols, c)
		valueArgs = append(valueArgs, mq.Height)
		valueArgs = append(valueArgs, mq.MinerID)
		valueArgs = append(valueArgs, mq.StateRoot)
		valueArgs = append(valueArgs, mq.DeadlineIndex)
		valueArgs = append(valueArgs, mq.PeriodStart)
		valueArgs = append(valueArgs, mq.Open)
		valueArgs = append(valueArgs, mq.Close)
		valueArgs = append(valueArgs, mq.Challenge)
		valueArgs = append(valueArgs, mq.FaultCutoff)
		c += cols
	}
	query = query[0 : len(query)-1]
	stmt, err := s.db.Prepare(query)
	if err != nil {
		log.Println("prep minerdeadlines", err)
	}
	res, err := stmt.Exec(valueArgs...)
	if err != nil {
		log.Println("insert minerdeadlines", err)
	}
	log.Info("res minerdeadlines", res)
	return nil
}

func (s *Store) PersistMinerQuality(mqs []minermodel.MinerQuality) error {
	if len(mqs) == 0 {
		log.Info("no mqs")
		return nil
	}
	query := "INSERT INTO miner_quality (" +
		"miner_id, quality_adj_power, raw_byte_power, win_count, data_stored, " +
		"blocks_mined, mining_efficiency, faulty_sectors, height)" +
		" VALUES "
	valueArgs := []interface{}{}
	cols := 9
	c := 1
	for _, mq := range mqs {
		// query += "($" + strconv.Itoa(c) + ", $" + strconv.Itoa(c+1) + ", $" + strconv.Itoa(c+2) + ", $" + strconv.Itoa(c+3) + ", $" + strconv.Itoa(c+4) + ", $" + strconv.Itoa(c+5) + ", $" + strconv.Itoa(c+6) + ", $" + strconv.Itoa(c+7) + ", $" + strconv.Itoa(c+8) + "),"
		query += generateQuery(cols, c)
		valueArgs = append(valueArgs, mq.MinerID)
		valueArgs = append(valueArgs, mq.QualityAdjPower)
		valueArgs = append(valueArgs, mq.RawBytePower)
		valueArgs = append(valueArgs, mq.WinCount)
		valueArgs = append(valueArgs, mq.DataStored)
		valueArgs = append(valueArgs, mq.BlocksMined)
		valueArgs = append(valueArgs, mq.MiningEfficiency)
		valueArgs = append(valueArgs, mq.FaultySectors)
		valueArgs = append(valueArgs, mq.Height)
		c += cols
	}
	query = query[0 : len(query)-1]
	stmt, err := s.db.Prepare(query)
	if err != nil {
		log.Println("prep minerqual", err)
	}
	res, err := stmt.Exec(valueArgs...)
	if err != nil {
		log.Println("insert minerqual", err)
	}
	log.Info("res minerqual", res)
	return nil
}

func (s *Store) PersistPowerActorClaims(pacs []*powermodel.PowerActorClaim) error {
	if len(pacs) == 0 {
		log.Info("no pacs")
		return nil
	}
	query := "INSERT INTO power_actor_claims (" +
		"miner_id, height, state_root, raw_byte_power, quality_adj_power)" +
		" VALUES "
	valueArgs := []interface{}{}
	cols := 5
	c := 1
	for _, mq := range pacs {
		// query += "($" + strconv.Itoa(c) + ", $" + strconv.Itoa(c+1) + ", $" + strconv.Itoa(c+2) + ", $" + strconv.Itoa(c+3) + ", $" + strconv.Itoa(c+4) + ", $" + strconv.Itoa(c+5) + ", $" + strconv.Itoa(c+6) + ", $" + strconv.Itoa(c+7) + ", $" + strconv.Itoa(c+8) + "),"
		query += generateQuery(cols, c)
		valueArgs = append(valueArgs, mq.MinerID)
		valueArgs = append(valueArgs, mq.Height)
		valueArgs = append(valueArgs, mq.StateRoot)
		valueArgs = append(valueArgs, mq.RawBytePower)
		valueArgs = append(valueArgs, mq.QualityAdjPower)
		c += cols
	}
	query = query[0 : len(query)-1]
	stmt, err := s.db.Prepare(query)
	if err != nil {
		log.Println("prep pac", err)
	}
	res, err := stmt.Exec(valueArgs...)
	if err != nil {
		log.Println("insert pac", err)
	}
	log.Info("res pac", res)
	return nil
}

func (s *Store) PersistMinerSectors(msis []*minermodel.MinerSectorInfo) error {
	if len(msis) == 0 {
		log.Info("no minersectorinfos")
		return nil
	}
	query := "INSERT INTO miner_sector_infos (" +
		"miner_id, sector_id, state_root, sealed_cid, activation_epoch, " +
		"expiration_epoch, deal_weight, verified_deal_weight, initial_pledge, " +
		"expected_day_reward, expected_storage_pledge, height)" +
		" VALUES "
	valueArgs := []interface{}{}
	cols := 12
	c := 1
	for _, msi := range msis {
		// query += "($" + strconv.Itoa(c) + ", $" + strconv.Itoa(c+1) + ", $" + strconv.Itoa(c+2) + ", $" + strconv.Itoa(c+3) + ", $" + strconv.Itoa(c+4) + ", $" + strconv.Itoa(c+5) + ", $" + strconv.Itoa(c+6) + ", $" + strconv.Itoa(c+7) + ", $" + strconv.Itoa(c+8) + ", $" + strconv.Itoa(c+9) + ", $" + strconv.Itoa(c+10) + ", $" + strconv.Itoa(c+11) + "),"
		query += generateQuery(cols, c)
		valueArgs = append(valueArgs, msi.MinerID)
		valueArgs = append(valueArgs, msi.SectorID)
		valueArgs = append(valueArgs, msi.StateRoot)
		valueArgs = append(valueArgs, msi.SealedCID)
		valueArgs = append(valueArgs, msi.ActivationEpoch)
		valueArgs = append(valueArgs, msi.ExpirationEpoch)
		valueArgs = append(valueArgs, msi.DealWeight)
		valueArgs = append(valueArgs, msi.VerifiedDealWeight)
		valueArgs = append(valueArgs, msi.InitialPledge)
		valueArgs = append(valueArgs, msi.ExpectedDayReward)
		valueArgs = append(valueArgs, msi.ExpectedStoragePledge)
		valueArgs = append(valueArgs, msi.Height)
		c += cols
	}
	query = query[0 : len(query)-1]
	log.Info("Query", query)
	stmt, err := s.db.Prepare(query)
	if err != nil {
		log.Println("prep minersectorinfos", err)
	}
	log.Info("VARGS", valueArgs)
	res, err := stmt.Exec(valueArgs...)
	if err != nil {
		log.Println("insert minersectorinfos", err)
	}
	log.Info("res minersectorinfos", res)
	return nil
}

func (s *Store) PersistMinerSectorDeals(msds []minermodel.MinerSectorDeal) error {
	return nil
}

func (s *Store) PersistMinerSectorEvents(mses []minermodel.MinerSectorEvent) error {
	return nil
}

func (s *Store) PersistMinerSectorPosts(msps []minermodel.MinerSectorPost) error {
	return nil
}

func (s *Store) PersistMinerSectorFaults(msfs []*minermodel.MinerSectorFault) error {
	if len(msfs) == 0 {
		log.Info("no msfaults")
		return nil
	}
	query := "INSERT INTO miner_sector_faults (" +
		"height, miner_id, sector_id)" +
		" VALUES "
	valueArgs := []interface{}{}
	cols := 3
	c := 1
	for _, mq := range msfs {
		// query += "($" + strconv.Itoa(c) + ", $" + strconv.Itoa(c+1) + ", $" + strconv.Itoa(c+2) + "),"
		query += generateQuery(cols, c)
		valueArgs = append(valueArgs, mq.Height)
		valueArgs = append(valueArgs, mq.MinerID)
		valueArgs = append(valueArgs, mq.SectorID)
		c += cols
	}
	query = query[0 : len(query)-1]
	stmt, err := s.db.Prepare(query)
	if err != nil {
		log.Println("prep msfaults", err)
	}
	res, err := stmt.Exec(valueArgs...)
	if err != nil {
		log.Println("insert msfaults", err)
	}
	log.Info("res msfaults", res)
	return nil
}

func (s *Store) PersistMarketDealProposals(mdps []marketmodel.MarketDealProposal) error {
	if len(mdps) == 0 {
		log.Info("no marketdealprops")
		return nil
	}
	query := "INSERT INTO market_deal_proposals (" +
		"deal_id, state_root, piece_cid, padded_piece_size, unpadded_piece_size, " +
		"is_verified, client_id, provider_id, start_epoch, end_epoch, " +
		"storage_price_per_epoch, provider_collateral, " +
		"client_collateral, label, height)" +
		" VALUES "
	valueArgs := []interface{}{}
	cols := 15
	c := 1
	for _, msi := range mdps {
		// query += "($" + strconv.Itoa(c) + ", $" + strconv.Itoa(c+1) + ", $" + strconv.Itoa(c+2) + ", $" + strconv.Itoa(c+3) + ", $" + strconv.Itoa(c+4) + ", $" + strconv.Itoa(c+5) + ", $" + strconv.Itoa(c+6) + ", $" + strconv.Itoa(c+7) + ", $" + strconv.Itoa(c+8) + ", $" + strconv.Itoa(c+9) + ", $" + strconv.Itoa(c+10) + ", $" + strconv.Itoa(c+11) + ", $" + strconv.Itoa(c+12) + ", $" + strconv.Itoa(c+13) + ", $" + strconv.Itoa(c+14) + "),"
		query += generateQuery(cols, c)
		valueArgs = append(valueArgs, msi.DealID)
		valueArgs = append(valueArgs, msi.StateRoot)
		valueArgs = append(valueArgs, msi.PieceCID)
		valueArgs = append(valueArgs, msi.PaddedPieceSize)
		valueArgs = append(valueArgs, msi.UnpaddedPieceSize)
		valueArgs = append(valueArgs, msi.IsVerified)
		valueArgs = append(valueArgs, msi.ClientID)
		valueArgs = append(valueArgs, msi.ProviderID)
		valueArgs = append(valueArgs, msi.StartEpoch)
		valueArgs = append(valueArgs, msi.EndEpoch)
		// valueArgs = append(valueArgs, msi.SlashedEpoch)
		valueArgs = append(valueArgs, msi.StoragePricePerEpoch)
		valueArgs = append(valueArgs, msi.ProviderCollateral)
		valueArgs = append(valueArgs, msi.ClientCollateral)
		valueArgs = append(valueArgs, msi.Label)
		valueArgs = append(valueArgs, msi.Height)
		c += cols
	}
	query = query[0 : len(query)-1]
	stmt, err := s.db.Prepare(query)
	if err != nil {
		log.Println("prep marketdealprops", err)
	}
	res, err := stmt.Exec(valueArgs...)
	if err != nil {
		log.Println("insert marketdealprops", err)
	}
	log.Info("res marketdealprops", res)
	return nil
}

// func (s *Store) UpdateWinCount() {
// 	// minermodel.MinerQuality{}
// 	rows, err := s.db.Query("SELECT miner_id, SUM(win_count) from block_headers GROUP BY miner_id")
// 	fmt.Println("rows", rows)
// }

func generateQuery(n int, c int) string {
	query := "($" + strconv.Itoa(c)
	for i := 1; i < n; i++ {
		query += ", $" + strconv.Itoa(c+i)
	}
	query += "),"
	return query
}
