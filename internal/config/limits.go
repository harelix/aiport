package config

// ProxyLimits are independent of telemetry availability and response length.
type ProxyLimits struct {
	MaxConcurrent           int   `yaml:"max_concurrent" json:"max_concurrent"`
	MaxRequestBytes         int64 `yaml:"max_request_bytes" json:"max_request_bytes"`
	ConnectTimeoutMS        int   `yaml:"connect_timeout_ms" json:"connect_timeout_ms"`
	ResponseHeaderTimeoutMS int   `yaml:"response_header_timeout_ms" json:"response_header_timeout_ms"`
	StreamIdleTimeoutMS     int   `yaml:"stream_idle_timeout_ms" json:"stream_idle_timeout_ms"`
	RequestReadTimeoutMS    int   `yaml:"request_read_timeout_ms" json:"request_read_timeout_ms"`
}

func DefaultProxyLimits() ProxyLimits { return ProxyLimits{32, 8 << 20, 10000, 30000, 120000, 30000} }
