package main

type DagPayload struct {
	PayloadCid string `json:payloadCid`
	Height     int64  `json:height`
	Data       DagPayloadData
}

type DagPayloadData struct {
	Contract string `json:"contract"`
	/* Commenting out payload Data, to keep it generic.
	Token0Reserves map[string]float64  `json:"token0Reserves"`
	Token1Reserves map[string]float64 `json:"token1Reserves"`*/
	ChainHeightRange struct {
		Begin int64 `json:"begin"`
		End   int64 `json:"end"`
	} `json:"chainHeightRange"`
	BroadcastID string  `json:"broadcast_id"`
	Timestamp   float64 `json:"timestamp"`
}

type DagChainBlock struct {
	Data struct {
		Cid  string `json:"cid"`
		Type string `json:"type"`
	} `json:"data"`
	Height    int64  `json:"height"`
	PrevCid   string `json:"prevCid"`
	Timestamp int64  `json:"timestamp"`
	TxHash    string `json:"txHash"`
}
