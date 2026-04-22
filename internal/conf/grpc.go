package conf

// GRPC 配置
type GRPCConfig struct {
	// gRPC 服务器地址
	Addr string `json:"addr" yaml:"addr"`
	
	// 是否启用 TLS
	EnableTLS bool `json:"enable_tls" yaml:"enable_tls"`
	
	// TLS 证书文件路径
	CertFile string `json:"cert_file" yaml:"cert_file"`
	
	// TLS 私钥文件路径
	KeyFile string `json:"key_file" yaml:"key_file"`
	
	// 最大连接数
	MaxConnections int `json:"max_connections" yaml:"max_connections"`
}
