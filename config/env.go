package config

import (
	"crypto/x509"
	"encoding/base64"
	"log"
	"os"
	"strings"

	gormcrypto "github.com/pkasila/gorm-crypto"
	"github.com/pkasila/gorm-crypto/algorithms"
	"github.com/pkasila/gorm-crypto/serialization"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config/telemetry"
	"github.com/soluixdeveloper/ces-utilities/v2/cesdatabase/cesgorm"
	cesenv "github.com/soluixdeveloper/ces-utilities/v2/cesenv"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	"github.com/soluixdeveloper/ces-utilities/v2/cesresponse"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type EnvConfigApp struct {
	ProjectName      string `mapstructure:"PROJECT_NAME" validate:"required"`
	ProjectModule    string `mapstructure:"PROJECT_MODULE" validate:"required"`
	AppPort          string `mapstructure:"APP_PORT" validate:"required"`
	LogLevel         string `mapstructure:"LOG_LEVEL" validate:"required"`
	DBProvider       string `mapstructure:"DB_PROVIDER" validate:"required"`
	RedisUse         bool   `mapstructure:"REDIS_USE"`
	LocalLanguage    string `mapstructure:"LOCALE_LANGUAGE" validate:"required"`
	FormatTimestamp  string `mapstructure:"FORMAT_TIMESTAMP"`
	BackendCORS      string `mapstructure:"Backend_CORS"`
	Environment      string `mapstructure:"ENVIRONMENT"`
	ModuleVersion    string `mapstructure:"MODULE_VERSION"`
	RETENTIONLOG     int    `mapstructure:"RETENTION_LOG"`
	LogMaskDepth     int    `mapstructure:"LOG_MASK_DEPTH"`
	LogMaskThreshold int    `mapstructure:"LOG_MASK_THRESHOLD"`
	RunMigration     bool   `mapstructure:"RUN_MIGRATION"`
}

type EnvConfigPostgre struct {
	PostgreSQLHost                 string `mapstructure:"POSTGRE_SQL_HOST" validate:"required"`
	PostgreSQLPort                 string `mapstructure:"POSTGRE_SQL_PORT" validate:"required"`
	PostgreSQLUser                 string `mapstructure:"POSTGRE_SQL_USER" validate:"required"`
	PostgreSQLPassword             string `mapstructure:"POSTGRE_SQL_PASSWORD" validate:"required"`
	PostgreSQLDBName               string `mapstructure:"POSTGRE_SQL_DB_NAME" validate:"required"`
	PostgreMaxOpenConnection       int    `mapstructure:"POSTGRE_SQL_MAX_OPEN_CONNECTION" validate:"required"`
	PostgreMaxIdleConnection       int    `mapstructure:"POSTGRE_SQL_MAX_IDLE_CONNECTION" validate:"required"`
	PostgreSQLOrchestratorUser     string `mapstructure:"POSTGRE_SQL_ORCHESTRATOR_USER"`
	PostgreSQLOrchestratorPassword string `mapstructure:"POSTGRE_SQL_ORCHESTRATOR_PASSWORD"`
}

type EnvConfigArangoDB struct {
	DbUsername string `mapstructure:"ARANGO_USERNAME"`
	DbPassword string `mapstructure:"ARANGO_PASSWORD"`
	DbName     string `mapstructure:"ARANGO_DB_NAME"`
	DbURL      string `mapstructure:"ARANGO_URL"`
}

type EnvConfigRedis struct {
	RedisURL             string `mapstructure:"REDIS_URL"`
	RedisPort            string `mapstructure:"REDIS_PORT"`
	RedisPassword        string `mapstructure:"REDIS_PASSWORD"`
	CacheDuration        int64  `mapstructure:"CACHE_DURATION"`
	RedisWithoutPassword bool   `mapstructure:"REDIS_WITHOUT_PASSWORD"`
}

type EnvConfigRabbitMQ struct {
	RabbitMQURL          string `mapstructure:"RABBITMQURL"`
	RabbitMQHost         string `mapstructure:"RABBIT_MQ_HOST"`
	RabbitMQPort         string `mapstructure:"RABBIT_MQ_PORT"`
	RabbitMQUsername     string `mapstructure:"RABBIT_MQ_USERNAME"`
	RabbitMQPassword     string `mapstructure:"RABBIT_MQ_PASSWORD"`
	RabbitMQStreamMaxAge string `mapstructure:"RABBITMQ_STREAM_MAX_AGE"`
	RabbitMQTimeout      int    `mapstructure:"RABBIT_MQ_TIMEOUT"`
}

type EnvConfigCrypto struct {
	CryptoRSA string `mapstructure:"CRYPTO_RSA"`
}

type EnvConfigElasticSearch struct {
	ElasticSearchURL      string `mapstructure:"ELASTICSEARCH_URL"`
	ElasticSearchPort     string `mapstructure:"ELASTICSEARCH_PORT"`
	ElasticSearchUsername string `mapstructure:"ELASTICSEARCH_USERNAME"`
	ElasticSearchPassword string `mapstructure:"ELASTICSEARCH_PASSWORD"`
	ElasticSearchAPIKey   string `mapstructure:"ELASTICSEARCH_API_KEY"`
}

type EnvConfigTelemetry struct {
	TempoEndpoint    string `mapstructure:"TEMPO_ENDPOINT"`
	PrometheusPort   string `mapstructure:"PROMETHEUS_PORT"`
	TelemetryEnabled bool   `mapstructure:"TELEMETRY_ENABLED"`
}

var (
	AppConfig           EnvConfigApp
	PostgreConfig       EnvConfigPostgre
	ArangoDBConfig      EnvConfigArangoDB
	RedisConfig         EnvConfigRedis
	RabbitMQConfig      EnvConfigRabbitMQ
	CryptoConfig        EnvConfigCrypto
	ElasticSearchConfig EnvConfigElasticSearch
	TelemetryConfig     EnvConfigTelemetry
	// Global variables
	PostgreDB              *gorm.DB
	WofkflowConfigurations = make(map[string]model.YamlWorkload)
	RequestTypes           = make(map[string]string)
	Endpoints              []model.Endpoints
)

func InitConfig() {
	// initiate viper config env
	envPath := []string{
		"../../../.",
		"../.",
		".",
	}
	envModel := map[string]interface{}{
		"app-config":           &AppConfig,
		"postgre-config":       &PostgreConfig,
		"arangodb-config":      &ArangoDBConfig,
		"redis-config":         &RedisConfig,
		"rabbit-mq-config":     &RabbitMQConfig,
		"crypto-config":        &CryptoConfig,
		"elasticsearch-config": &ElasticSearchConfig,
		"telemetry-config":     &TelemetryConfig,
	}
	cesenv.InitEnv("config", "env", envPath, envModel, false)
	if RabbitMQConfig.RabbitMQTimeout < 1 {
		RabbitMQConfig.RabbitMQTimeout = 60
	}
	// initiate logger
	ceslogger.InitLogger(
		AppConfig.LogLevel,
		AppConfig.Environment,
		AppConfig.ProjectName,
		AppConfig.ProjectModule,
		AppConfig.ModuleVersion)

	// define user and password
	postgreUser := PostgreConfig.PostgreSQLUser
	postgrePassword := PostgreConfig.PostgreSQLPassword
	if PostgreConfig.PostgreSQLOrchestratorUser != "" {
		postgreUser = PostgreConfig.PostgreSQLOrchestratorUser
	}
	if PostgreConfig.PostgreSQLOrchestratorPassword != "" {
		postgrePassword = PostgreConfig.PostgreSQLOrchestratorPassword
	}

	// auth database
	postgreDBConfig := cesgorm.NewDatabaseConfig(
		PostgreConfig.PostgreSQLHost,
		PostgreConfig.PostgreSQLPort,
		postgreUser,
		postgrePassword,
		PostgreConfig.PostgreSQLDBName,
		AppConfig.DBProvider,
		// cesgorm.SetMaxOpenConnections(PostgreConfig.PostgreMaxOpenConnection),
		// cesgorm.SetMaxIdleConnections(PostgreConfig.PostgreMaxIdleConnection),
	)
	// config:
	// - SkipDefaultTransaction: https://gorm.io/docs/transactions.html#Disable-Default-Transaction
	// - NamingStrategy: special case for table mybcas_parameter.
	postgreDB := postgreDBConfig.AuthDatabase(&gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			NameReplacer: strings.NewReplacer("MyBCAS", "mybcas"),
		},
		// Logger:                 logger.Default.LogMode(logger.Info),
	})

	// init RSA
	initRSA()

	// Jika RUN_MIGRATION tidak di set, var RunMigration set default ke true agar migration tetap berjalan saat service di start.
	// Jika RUN_MIGRATION di set dan value nya false, migration tidak akan berjalan tapi service lain tetap dijalankan.
	if _, isSet := os.LookupEnv("RUN_MIGRATION"); !isSet {
		AppConfig.RunMigration = true
	}

	logging := ceslogger.NewLogger("")
	if AppConfig.RunMigration {
		logging.LogInfo("RunMigration: starting auto migrate")
		// auto migrate tables using models
		postgreDB.AutoMigrate(map[string]interface{}{
			"workflow_audit":         model.WorkflowAudit{},
			"workflow_configuration": model.WorkflowConfiguration{},
			"workflow_history":       model.WorkflowHistory{},
			"workflow_state":         model.WorkflowState{},
			"workflow":               model.Workflow{},
		})
		logging.LogInfo("RunMigration: auto migrate completed")
	} else {
		logging.LogInfo("RunMigration: skipped (RUN_MIGRATION=false)")
	}

	// Register GORM tracing if telemetry is enabled
	if TelemetryConfig.TelemetryEnabled {
		logging := ceslogger.Logger{}
		if err := telemetry.RegisterGormTracing(postgreDB.DB, PostgreConfig.PostgreSQLDBName); err != nil {
			logging.LogError("Failed to register GORM tracing", err.Error())
		} else {
			logging.LogInfo("GORM tracing registered successfully")
		}
	}

	PostgreDB = postgreDB.DB
	cesresponse.NewErrorMessageSource(
		cesresponse.AddSourceDatabaseGorm(PostgreDB),
		cesresponse.SetGetSequence("postgre"),
		cesresponse.OverwriteDefaultSource(),
	)

}

func initRSA() {
	logging := ceslogger.Logger{}
	// REMOVE VALUE UNUSE FROM CRYPTORSA
	cryptoRSA := strings.ReplaceAll(CryptoConfig.CryptoRSA, "-----BEGIN RSA PRIVATE KEY-----", "")
	cryptoRSA = strings.ReplaceAll(cryptoRSA, "-----END RSA PRIVATE KEY-----", "")
	cryptoRSA = strings.ReplaceAll(cryptoRSA, " ", "")

	// DECODE CRYPTO RSA ENV
	base64Data := []byte(cryptoRSA)
	lenCryptoRSA := make([]byte, base64.StdEncoding.DecodedLen(len(base64Data)))
	valueCryptoRSA, err := base64.StdEncoding.Decode(lenCryptoRSA, base64Data)
	if err != nil {
		// Handle error
		logging.LogError("base64.StdEncoding.Decode(d, base64Data)", err)
		log.Fatal()
	}

	// PARSE CRYPTO RSA PRIVATE KEY
	lenCryptoRSA = lenCryptoRSA[:valueCryptoRSA]
	keyCryptoRSA, err := x509.ParsePKCS1PrivateKey(lenCryptoRSA)
	if err != nil {
		// Handle error
		logging.LogError("x509.ParsePKCS1PrivateKey(lenCryptoRSA)", err)
		log.Fatal()
	}

	// SET CTYPTO RSA PUBLIC KEY
	publicKey := &keyCryptoRSA.PublicKey

	// Use privateKey and publicKey to initialize gormcrypto
	gormcrypto.Init(algorithms.NewRSA(keyCryptoRSA, publicKey), serialization.NewJSON())
}
