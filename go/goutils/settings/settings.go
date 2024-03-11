package settings

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"
)

type (
	RateLimiter struct {
		Burst          int `json:"burst"`
		RequestsPerSec int `json:"req_per_sec"`
	}

	Signer struct {
		Domain struct {
			Name              string `json:"name"`
			Version           string `json:"version"`
			ChainId           string `json:"chainId"`
			VerifyingContract string `json:"verifyingContract"`
		} `json:"domain"`
		AccountAddress string `json:"accountAddress"`
		PrivateKey     string `json:"privateKey"`
		DeadlineBuffer int    `json:"deadlineBuffer"`
	}

	Rabbitmq struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Setup    struct {
			Core struct {
				Exchange              string `json:"exchange"`
				CommitPayloadExchange string `json:"commit_payload_exchange"`
				EventDetectorExchange string `json:"event_detector_exchange"`
			} `json:"core"`
			PayloadCommit struct {
				QueueNamePrefix  string `json:"queue_name_prefix"`
				RoutingKeyPrefix string `json:"routing_key_prefix"`
			} `json:"payload_commit"`
			EventDetector struct {
				QueueNamePrefix  string `json:"queue_name_prefix"`
				RoutingKeyPrefix string `json:"routing_key_prefix"`
			} `json:"event_detector"`
		} `json:"setup"`
	}

	IpfsConfig struct {
		URL              string          `json:"url" validate:"required"`
		ReaderURL        string          `json:"reader_url"`
		WriteRateLimiter *RateLimiter    `json:"write_rate_limit,omitempty"`
		ReadRateLimiter  *RateLimiter    `json:"read_rate_limit,omitempty"`
		Timeout          int             `json:"timeout"`
		MaxIdleConns     int             `json:"max_idle_conns"`
		IdleConnTimeout  int             `json:"idle_conn_timeout"`
		ReaderAuthConfig *IPFSAuthConfig `json:"reader_auth_config"`
		WriterAuthConfig *IPFSAuthConfig `json:"writer_auth_config"`
	}

	IPFSAuthConfig struct {
		ProjectApiKey    string `json:"project_api_key"`
		ProjectApiSecret string `json:"project_api_secret"`
	}

	Redis struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Db       int    `json:"db"`
		Password string `json:"password"`
		PoolSize int    `json:"pool_size"`
	}

	RedisReader struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Db       int    `json:"db"`
		Password string `json:"password"`
		PoolSize int    `json:"pool_size"`
	}

	Web3Storage struct {
		URL             string       `json:"url"`
		UploadURLSuffix string       `json:"upload_url_suffix"`
		APIToken        string       `json:"api_token"`
		Timeout         int          `json:"timeout"`
		MaxIdleConns    int          `json:"max_idle_conns"`
		IdleConnTimeout int          `json:"idle_conn_timeout"`
		RateLimiter     *RateLimiter `json:"rate_limit,omitempty"`
	}

	Relayer struct {
		Host     *string `json:"host" validate:"required"`
		Endpoint *string `json:"endpoint" validate:"required"`
	}

	Pruning struct {
		IPFSPinningMaxAge int    `json:"ipfs_pinning_max_age_in_days"`
		LocalDiskMaxAge   int    `json:"local_disk_max_age_in_days"`
		CronFrequency     string `json:"cron_frequency"`
	}

	HTTPClient struct {
		MaxIdleConns        int `json:"max_idle_conns"`
		MaxConnsPerHost     int `json:"max_conns_per_host"`
		MaxIdleConnsPerHost int `json:"max_idle_conns_per_host"`
		IdleConnTimeout     int `json:"idle_conn_timeout"`
		ConnectionTimeout   int `json:"connection_timeout"`
	}

	Reporting struct {
		SlackWebhookURL                string `json:"slack_webhook_url"`
		OffchainConsensusIssueEndpoint string `json:"offchain_consensus_issue_endpoint"`
	}

	Healthcheck struct {
		Port     int    `json:"port"`
		Endpoint string `json:"endpoint"`
	}
)

type SettingsObj struct {
	InstanceId        string       `json:"instance_id" validate:"required"`
	SlotId            int          `json:"slot_id" validate:"required"`
	PoolerNamespace   string       `json:"pooler_namespace" validate:"required"`
	AnchorChainRPCURL string       `json:"anchor_chain_rpc_url" validate:"required"`
	LocalCachePath    string       `json:"local_cache_path" validate:"required"`
	Concurrency       int          `json:"concurrency" validate:"required"`
	WorkerConcurrency int          `json:"worker_concurrency" validate:"required"`
	HttpClient        *HTTPClient  `json:"http_client" validate:"required,dive"`
	Rabbitmq          *Rabbitmq    `json:"rabbitmq" validate:"required,dive"`
	IpfsConfig        *IpfsConfig  `json:"ipfs" validate:"required,dive"`
	Redis             *Redis       `json:"redis" validate:"required,dive"`
	RedisReader       *RedisReader `json:"redis_reader" validate:"required,dive"`
	Web3Storage       *Web3Storage `json:"web3_storage" validate:"required,dive"`
	Signer            *Signer      `json:"signer" validate:"required,dive"`
	Relayer           *Relayer     `json:"relayer" validate:"required,dive"`
	Pruning           *Pruning     `json:"pruning" validate:"required,dive"`
	Reporting         *Reporting   `json:"reporting" validate:"required,dive"`
	Healthcheck       *Healthcheck `json:"healthcheck" validate:"required"`
}

// ParseSettings parses the settings.json file and returns a SettingsObj
func ParseSettings() *SettingsObj {
	log.Debug("parsing settings")

	v := validator.New()

	dir := strings.TrimSuffix(os.Getenv("CONFIG_PATH"), "/")
	settingsFilePath := dir + "/settings.json"

	settingsObj := new(SettingsObj)

	log.Info("reading settings:", settingsFilePath)

	data, err := os.ReadFile(settingsFilePath)
	if err != nil {
		log.Error("cannot read the file:", err)
		panic(err)
	}

	log.Debug("settings json data is", string(data))

	err = json.Unmarshal(data, settingsObj)
	if err != nil {
		log.Error("cannot unmarshal the settings json ", err)
		panic(err)
	}

	err = v.Struct(settingsObj)
	if err != nil {
		log.WithError(err).Fatal("invalid settings object")
	}

	SetDefaults(settingsObj)
	log.Infof("final Settings Object being used %+v", settingsObj)

	err = gi.Inject(settingsObj)
	if err != nil {
		log.Fatal("cannot inject the settings object", err)
	}

	return settingsObj
}

// SetDefaults sets the default values for the settings object
// add default values in this function if required
func SetDefaults(settingsObj *SettingsObj) {
	if settingsObj.Reporting.SlackWebhookURL == "" {
		log.Warning("slack webhook url is not set, errors will not be reported to slack")
	}

	settingsObj.LocalCachePath = strings.TrimSuffix(settingsObj.LocalCachePath, "/")

	// for local testing
	if val, err := strconv.ParseBool(os.Getenv("LOCAL_TESTING")); err == nil && val {
		settingsObj.Redis.Host = "localhost"
		settingsObj.RedisReader.Host = "localhost"
		settingsObj.Rabbitmq.Host = "localhost"
		settingsObj.IpfsConfig.ReaderURL = "/dns/localhost/tcp/5001"
		settingsObj.IpfsConfig.URL = "/dns/localhost/tcp/5001"
	}

	privKey := os.Getenv("PRIVATE_KEY")
	if privKey != "" {
		settingsObj.Signer.PrivateKey = privKey
	}

	ipfsReaderProjectApiKey := os.Getenv("IPFS_READER_PROJECT_API_KEY")
	if ipfsReaderProjectApiKey != "" {
		settingsObj.IpfsConfig.ReaderAuthConfig.ProjectApiKey = ipfsReaderProjectApiKey
	}

	ipfsReaderProjectApiSecret := os.Getenv("IPFS_READER_PROJECT_API_SECRET")
	if ipfsReaderProjectApiSecret != "" {
		settingsObj.IpfsConfig.ReaderAuthConfig.ProjectApiSecret = ipfsReaderProjectApiSecret
	}

	ipfsWriterProjectApiKey := os.Getenv("IPFS_WRITER_PROJECT_API_KEY")
	if ipfsWriterProjectApiKey != "" {
		settingsObj.IpfsConfig.WriterAuthConfig.ProjectApiKey = ipfsWriterProjectApiKey
	}

	ipfsWriterProjectApiSecret := os.Getenv("IPFS_WRITER_PROJECT_API_SECRET")
	if ipfsWriterProjectApiSecret != "" {
		settingsObj.IpfsConfig.WriterAuthConfig.ProjectApiSecret = ipfsWriterProjectApiSecret
	}

	if settingsObj.Healthcheck.Endpoint == "" {
		settingsObj.Healthcheck.Endpoint = "/health"
	}

	if settingsObj.Healthcheck.Port == 0 {
		settingsObj.Healthcheck.Port = 9000
	}
}
