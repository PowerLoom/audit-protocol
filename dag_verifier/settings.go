package main

import (
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"
)

type _DagVerifierSettings_ struct {
	SlackNotifyURL               string `json:"slack_notify_URL"`
	RunIntervalSecs              int    `json:"run_interval_secs"`
	SuppressNotificationTimeSecs int64  `json:"suppress_notification_for_secs"`
}

type SettingsObj struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	WebhookListener struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"webhook_listener"`
	IpfsURL          string `json:"ipfs_url"`
	SnapshotInterval int    `json:"snapshot_interval"`
	Rlimit           struct {
		FileDescriptors int `json:"file_descriptors"`
	} `json:"rlimit"`
	RPCMatic          string `json:"rpc_matic"`
	ContractAddresses struct {
		IuniswapV2Factory string `json:"iuniswap_v2_factory"`
		IuniswapV2Router  string `json:"iuniswap_v2_router"`
		IuniswapV2Pair    string `json:"iuniswap_v2_pair"`
		USDT              string `json:"USDT"`
		DAI               string `json:"DAI"`
		USDC              string `json:"USDC"`
		WETH              string `json:"WETH"`
		MAKER             string `json:"MAKER"`
		WETHUSDT          string `json:"WETH-USDT"`
	} `json:"contract_addresses"`
	MetadataCache string `json:"metadata_cache"`
	DagTableName  string `json:"dag_table_name"`
	Seed          string `json:"seed"`
	Redis         struct {
		Host     string      `json:"host"`
		Port     int         `json:"port"`
		Db       int         `json:"db"`
		Password interface{} `json:"password"`
	} `json:"redis"`
	RedisReader struct {
		Host     string      `json:"host"`
		Port     int         `json:"port"`
		Db       int         `json:"db"`
		Password interface{} `json:"password"`
	} `json:"redis_reader"`
	TableNames struct {
		APIKeys           string `json:"api_keys"`
		AccountingRecords string `json:"accounting_records"`
		RetreivalsSingle  string `json:"retreivals_single"`
		RetreivalsBulk    string `json:"retreivals_bulk"`
	} `json:"table_names"`
	CleanupServiceInterval   int    `json:"cleanup_service_interval"`
	AuditContract            string `json:"audit_contract"`
	AppName                  string `json:"app_name"`
	PowergateClientAddr      string `json:"powergate_client_addr"`
	MaxIpfsBlocks            int    `json:"max_ipfs_blocks"`
	MaxPendingPayloadCommits int    `json:"max_pending_payload_commits"`
	BlockStorage             string `json:"block_storage"`
	PayloadStorage           string `json:"payload_storage"`
	ContainerHeight          int    `json:"container_height"`
	BloomFilterSettings      struct {
		MaxElements int         `json:"max_elements"`
		ErrorRate   float64     `json:"error_rate"`
		Filename    interface{} `json:"filename"`
	} `json:"bloom_filter_settings"`
	PayloadCommitInterval      int      `json:"payload_commit_interval"`
	PruningServiceInterval     int      `json:"pruning_service_interval"`
	RetrievalServiceInterval   int      `json:"retrieval_service_interval"`
	DealWatcherServiceInterval int      `json:"deal_watcher_service_interval"`
	BackupTargets              []string `json:"backup_targets"`
	MaxPayloadCommits          int      `json:"max_payload_commits"`
	UnpinMode                  string   `json:"unpin_mode"`
	MaxPendingEvents           int      `json:"max_pending_events"`
	IpfsTimeout                int      `json:"ipfs_timeout"`
	SpanExpireTimeout          int      `json:"span_expire_timeout"`
	APIKey                     string   `json:"api_key"`
	AiohtttpTimeouts           struct {
		SockRead    int `json:"sock_read"`
		SockConnect int `json:"sock_connect"`
		Connect     int `json:"connect"`
	} `json:"aiohtttp_timeouts"`
	DagVerifierSettings _DagVerifierSettings_ `json:"dag_verifier"`
}

func ParseSettings(settingsFile string) *SettingsObj {
	var settingsObj SettingsObj
	log.Info("Reading Settings:", settingsFile)
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		log.Error("Cannot read the file:", err)
		panic(err)
	}

	log.Debug("Settings json data is", string(data))
	err = json.Unmarshal(data, &settingsObj)
	if err != nil {
		log.Error("Cannot unmarshal the settings json ", err)
		panic(err)
	}
	SetDefaults(&settingsObj)
	log.Infof("Final Settings Object being used %+v", settingsObj)
	return &settingsObj
	//log.Info("Settings for namespace", settingsObj.Development.Namespace)
}

func SetDefaults(settings *SettingsObj) {
	if settings.DagVerifierSettings.RunIntervalSecs == 0 {
		settings.DagVerifierSettings.RunIntervalSecs = 300
	}
	if settings.DagVerifierSettings.SlackNotifyURL == "" {
		log.Warnf("Slack Notification URL is not set, any issues observed by this service will not be notified.")
	}
	if settings.DagVerifierSettings.SuppressNotificationTimeSecs == 0 {
		settings.DagVerifierSettings.SuppressNotificationTimeSecs = 1800
	}
}
