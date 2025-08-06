package config

import "github.com/lahiruramesh/dock-route/internal/templates"

type DeployConfig struct {
    AppType       string
    ContainerName string
    ImageName     string
    SourcePath    string
    HostPort      string
    Template      *templates.Template
    DevMode       bool
}

type ProxyConfig struct {
    Port   string
    Domain string
}
