/*
Copyright © 2024 The Single Tunnel Authors - singletunnel@spotai.co
*/

package clntagent

import (
	api "github.com/spotai/singletunnel-api/coord"
)

type Config struct {
	Client          api.ClientInfo
	CsAddr          string
	Tls             *Tls `yaml:",omitempty"`
	Enable          bool
	OpenVpnPath     string
	SudoForOpenVpn  bool
	WireguardKeyDir string // non-empty when wireguard is enabled
}

type Tls struct {
	CertFile   string
	KeyFile    string
	Insecure   bool   `yaml:",omitempty"`
	CaFile     string `yaml:",omitempty"`
	ServerName string `yaml:",omitempty"`
}
