package models

const MAX_TOKEN_PRICE_HISTORY_INDEX int = 300

type TokenSummarySnapshotMeta struct {
	Cid                          string  `json:"cid"`
	TxHash                       string  `json:"txHash"`
	TxStatus                     int     `json:"txStatus"`
	PrevTxHash                   string  `json:"prevTxHash,omitempty"`
	DAGHeight                    int     `json:"dagHeight"`
	BeginBlockHeight24h          int64   `json:"begin_block_height_24h"`
	BeginBlockheightTimeStamp24h float64 `json:"begin_block_timestamp_24h"`
	BeginBlockHeight7d           int64   `json:"begin_block_height_7d"`
	BeginBlockheightTimeStamp7d  float64 `json:"begin_block_timestamp_7d"`
}

type AuditProtocolErrorResp struct {
	Error string `json:"error"`
}

type AuditProtocolBlockHeightResp struct {
	Height int64 `json:"height"`
}

type AuditProtocolCommitPayloadReq struct {
	ProjectId       string      `json:"projectId"`
	Payload         _TokensData `json:"payload"`
	Web3Storage     bool        `json:"web3Storage"`
	SkipAnchorProof bool        `json:"skipAnchorProof"`
}

type _TokensData struct {
	TokensData []*TokenData `json:"data"`
}

type AuditProtocolCommitPayloadResp struct {
	TentativeHeight int    `json:"tentativeHeight"`
	CommitID        string `json:"commitId"`
}

type IPLDLink struct {
	LinkData string `json:"/"`
}

type AuditProtocolBlockResp struct {
	Data struct {
		Cid  IPLDLink `json:"cid"`
		Type string   `json:"type"`
	} `json:"data"`
	Height    int      `json:"height"`
	PrevCid   IPLDLink `json:"prevCid"`
	Timestamp int      `json:"timestamp"`
	TxHash    string   `json:"txHash"`
}

const (
	SNAPSHOT_COMMIT_PENDING = 1
	TX_ACK_PENDING          = 2
	TX_CONFIRMATION_PENDING = 3
	TX_CONFIRMED            = 4
)

type AuditProtocolBlockHeightStatusResp struct {
	ProjectId   string `json:"project_id"`
	BlockHeight int    `json:"block_height"`
	PayloadCid  string `json:"payload_cid"`
	TxHash      string `json:"tx_hash"`
	Status      int    `json:"status"`
}

type TokenPriceHistory struct {
	Timestamp   float64 `json:"timeStamp"`
	Price       float64 `json:"price"`
	BlockHeight int     `json:"blockHeight"`
}

type BlockHeightConfirmationPayload struct {
	CommitID  string `json:"commitID"`
	ProjectID string `json:"projectId"`
}

type TokenData struct {
	ContractAddress       string  `json:"contractAddress"`
	BlockHeight           int     `json:"block_height"`
	BlockTimestamp        int     `json:"block_timestamp"`
	Name                  string  `json:"name"`
	Symbol                string  `json:"symbol"`
	Price                 float64 `json:"price"`
	Liquidity             float64 `json:"liquidity"`
	LiquidityUSD          float64 `json:"liquidityUSD"`
	TradeVolume24h        float64 `json:"tradeVolume_24h"`
	TradeVolumeUSD24h     float64 `json:"tradeVolumeUSD_24h"`
	TradeVolume7d         float64 `json:"tradeVolume_7d"`
	TradeVolumeUSD7d      float64 `json:"tradeVolumeUSD_7d"`
	PriceChangePercent24h float64 `json:"priceChangePercent_24h"`
}

type PairSummarySnapshot struct {
	Data []*TokenPairLiquidityProcessedData `json:"data"`
}

type DAGBlockRange struct {
	HeadBlockCid string `json:"head_block_cid"`
	TailBlockCid string `json:"tail_block_cid"`
}

type TokenPairLiquidityProcessedData struct {
	ContractAddress         string        `json:"contractAddress"`
	Name                    string        `json:"name"`
	Liquidity               string        `json:"liquidity"`
	Volume24h               string        `json:"volume_24h"`
	Volume7d                string        `json:"volume_7d"`
	CidVolume24h            DAGBlockRange `json:"cid_volume_24h"`
	CidVolume7d             DAGBlockRange `json:"cid_volume_7d"`
	Fees24h                 string        `json:"fees_24h"`
	BlockHeight             int           `json:"block_height"`
	BlockTimestamp          int           `json:"block_timestamp"`
	DeltaToken0Reserves     float64       `json:"deltaToken0Reserves"`
	DeltaToken1Reserves     float64       `json:"deltaToken1Reserves"`
	DeltaTime               float64       `json:"deltaTime"`
	LatestTimestamp         float64       `json:"latestTimestamp"`
	EarliestTimestamp       float64       `json:"earliestTimestamp"`
	Token0Liquidity         float64       `json:"token0Liquidity"`
	Token1Liquidity         float64       `json:"token1Liquidity"`
	Token0LiquidityUSD      float64       `json:"token0LiquidityUSD"`
	Token1LiquidityUSD      float64       `json:"token1LiquidityUSD"`
	Token0TradeVolume24h    float64       `json:"token0TradeVolume_24h"`
	Token1TradeVolume24h    float64       `json:"token1TradeVolume_24h"`
	Token0TradeVolumeUSD24h float64       `json:"token0TradeVolumeUSD_24h"`
	Token1TradeVolumeUSD24h float64       `json:"token1TradeVolumeUSD_24h"`
	Token0TradeVolume7d     float64       `json:"token0TradeVolume_7d"`
	Token1TradeVolume7d     float64       `json:"token1TradeVolume_7d"`
	Token0TradeVolumeUSD7d  float64       `json:"token0TradeVolumeUSD_7d"`
	Token1TradeVolumeUSD7d  float64       `json:"token1TradeVolumeUSD_7d"`
}
