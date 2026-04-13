package types

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	LabelSandboxContainer = "sandbox_container"
	LabelSSHPort          = "sandbox_ssh_port"
	LabelSSHUsername      = "sandbox_ssh_username"
	LabelSSHPassword      = "sandbox_ssh_password"
)

const (
	SandboxLabelPrefix = "sandbox_"
	TraefikLabelPrefix = "traefik."
)

func IsEphemeralLabel(key string) bool {
	return strings.HasPrefix(key, SandboxLabelPrefix) || strings.HasPrefix(key, TraefikLabelPrefix)
}

func SandboxBaseLabels() map[string]string {
	return map[string]string{LabelSandboxContainer: "true"}
}

func SSHLabels(port int, username, password string) map[string]string {
	return map[string]string{
		LabelSSHPort:     strconv.Itoa(port),
		LabelSSHUsername: username,
		LabelSSHPassword: password,
	}
}

type TraefikLabelConfig struct {
	ContainerName string
	Hostname      string
	InternalPort  int
	Network       string
	Enable        bool
	Entrypoints   string
	CertResolver  string
	Middlewares   string
}

func TraefikLabels(cfg TraefikLabelConfig) map[string]string {
	if !cfg.Enable {
		return nil
	}

	router := "traefik.http.routers." + cfg.ContainerName
	service := "traefik.http.services." + cfg.ContainerName

	labels := map[string]string{
		"traefik.enable":                      "true",
		router + ".rule":                      fmt.Sprintf("Host(`%s`)", cfg.Hostname),
		router + ".service":                   cfg.ContainerName,
		service + ".loadbalancer.server.port": strconv.Itoa(cfg.InternalPort),
	}

	if cfg.Network != "" {
		labels["traefik.docker.network"] = cfg.Network
	}
	if cfg.Entrypoints != "" {
		labels[router+".entrypoints"] = cfg.Entrypoints
	}
	if cfg.CertResolver != "" {
		labels[router+".tls"] = "true"
		labels[router+".tls.certresolver"] = cfg.CertResolver
	}
	if cfg.Middlewares != "" {
		labels[router+".middlewares"] = cfg.Middlewares
	}

	return labels
}

func MergeLabels(maps ...map[string]string) map[string]string {
	size := 0
	for _, m := range maps {
		size += len(m)
	}
	out := make(map[string]string, size)
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}
