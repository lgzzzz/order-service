package conf

// ConfigCenter 配置中心
type ConfigCenter struct {
	Endpoints []string `json:"endpoints" yaml:"endpoints"`
	Key       string   `json:"key" yaml:"key"`
}

// Bootstrap 启动配置
type Bootstrap struct {
	Kafka        KafkaConfig    `json:"kafka" yaml:"kafka"`
	Database     DatabaseConfig `json:"database" yaml:"database"`
	GRPC         GRPCConfig     `json:"grpc" yaml:"grpc"`
	Registry     RegistryConfig `json:"registry" yaml:"registry"`
	ConfigCenter *ConfigCenter  `json:"config_center" yaml:"config_center"`
	Tracing      TracingConfig  `json:"tracing" yaml:"tracing"`
}

// TracingConfig 链路追踪配置
type TracingConfig struct {
	Enabled     bool    `json:"enabled" yaml:"enabled"`
	Endpoint    string  `json:"endpoint" yaml:"endpoint"`
	SampleRatio float64 `json:"sample_ratio" yaml:"sample_ratio"`
}

// Kafka 配置
type KafkaConfig struct {
	// Kafka 服务器地址列表
	Brokers []string `json:"brokers" yaml:"brokers"`

	// 消费者组 ID
	GroupID string `json:"group_id" yaml:"group_id"`

	// 订单主题
	OrderTopic string `json:"order_topic" yaml:"order_topic"`

	// 消费者数量
	Workers int `json:"workers" yaml:"workers"`

	// 最小提交字节数
	MinBytes int `json:"min_bytes" yaml:"min_bytes"`

	// 最大提交字节数
	MaxBytes int `json:"max_bytes" yaml:"max_bytes"`

	// 读取超时时间（秒）
	ReadTimeout int `json:"read_timeout" yaml:"read_timeout"`
}

// 数据库配置
type DatabaseConfig struct {
	DSN          string `json:"dsn" yaml:"dsn"`
	MaxIdleConns int    `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns int    `json:"max_open_conns" yaml:"max_open_conns"`
}

// Registry 配置
type RegistryConfig struct {
	// etcd 服务器地址列表
	Endpoints []string `json:"endpoints" yaml:"endpoints"`
	// 超时时间（秒）
	Timeout int `json:"timeout" yaml:"timeout"`
}

// GRPC 配置
type GRPCConfig struct {
	Addr           string `json:"addr" yaml:"addr"`
	Timeout        int    `json:"timeout" yaml:"timeout"`
	MaxConnections int    `json:"max_connections" yaml:"max_connections"`
}

// Config 应用配置 (别名，用于兼容)
type Config = Bootstrap
