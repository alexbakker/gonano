package config

type Config struct {
	Addr      string   `json:"addr"`
	AddrRPC   string   `json:"addr_rpc"`
	AddrPprof string   `json:"addr_pprof"`
	Peers     []string `json:"peers"`
}
