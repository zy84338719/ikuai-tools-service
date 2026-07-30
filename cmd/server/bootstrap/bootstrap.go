package bootstrap

import (
	"fmt"
	neturl "net/url"
	"net/http"
	"os"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	exportermetrics "github.com/zy84338719/ikuai_exporter/metrics"
	hertzrouter "github.com/zy84338719/ikuai-tools-service/gen/http/router"
	"github.com/zy84338719/ikuai-tools-service/internal/conf"
	"github.com/zy84338719/ikuai-tools-service/internal/ikuai"
	"github.com/zy84338719/ikuai-tools-service/internal/job"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
	"github.com/zy84338719/ikuai-tools-service/internal/repo/db"
	apihandler "github.com/zy84338719/ikuai-tools-service/internal/transport/http/handler"
	"github.com/zy84338719/ikuai-tools-service/internal/transport/http/middleware"
)

// Bootstrap initializes all components and returns the Hertz server instance.
func Bootstrap() (*server.Hertz, error) {
	if err := initConfig(); err != nil {
		return nil, fmt.Errorf("init config failed: %w", err)
	}

	cfg := conf.GlobalConfig

	if err := initLogger(cfg); err != nil {
		return nil, fmt.Errorf("init logger failed: %w", err)
	}

	if err := initDatabase(cfg); err != nil {
		return nil, fmt.Errorf("init database failed: %w", err)
	}

	if err := initIKuaiManager(cfg); err != nil {
		logger.Error(fmt.Sprintf("init ikuai client failed (continuing): %v", err))
	}

	h := initServer(cfg)

	if cfg.Metrics.Enabled {
		registry := initMetrics(cfg)
		startMetricsServer(cfg, registry)
	}

	if err := job.Start(&cfg.Jobs); err != nil {
		logger.Error(fmt.Sprintf("start job scheduler failed: %v", err))
	}

	return h, nil
}

func initConfig() error {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	if err := conf.Init(configPath); err != nil {
		fmt.Printf("Failed to load config from %s: %v\n", configPath, err)
		if err := conf.InitWithDefault(); err != nil {
			return err
		}
	}
	return nil
}

func initLogger(cfg *conf.Config) error {
	return logger.Init(&logger.Config{
		Level:      cfg.Log.Level,
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	})
}

func initDatabase(cfg *conf.Config) error {
	return db.Init(&cfg.Database)
}

func initIKuaiManager(cfg *conf.Config) error {
	return ikuai.Init(&cfg.IKuai)
}

func initServer(cfg *conf.Config) *server.Hertz {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	h := server.New(server.WithHostPorts(addr))

	h.Use(middleware.Recovery())
	h.Use(middleware.Logger())
	h.Use(middleware.CORS())

	hertzrouter.GeneratedRegister(h)
	registerIKuaiRoutes(h)

	return h
}

func registerIKuaiRoutes(h *server.Hertz) {
	v1 := h.Group("/api/v1/ikuai")

	// System
	sys := v1.Group("/system")
	sys.GET("/status", apihandler.GetSystemStatus)
	sys.GET("/interfaces", apihandler.GetInterfaces)
	sys.GET("/devices", apihandler.GetLanDevices)

	// Firewall — custom ISP
	fw := v1.Group("/firewall")
	fw.GET("/custom-isp", apihandler.ListCustomISP)
	fw.POST("/custom-isp", apihandler.AddCustomISP)
	fw.DELETE("/custom-isp/:ids", apihandler.DeleteCustomISP)

	// Firewall — stream domain
	fw.GET("/stream-domain", apihandler.ListStreamDomain)
	fw.POST("/stream-domain", apihandler.AddStreamDomain)
	fw.DELETE("/stream-domain/:ids", apihandler.DeleteStreamDomain)

	// Firewall — ACL
	fw.GET("/acl", apihandler.ListACL)
	fw.POST("/acl", apihandler.AddACL)
	fw.PUT("/acl", apihandler.EditACL)
	fw.DELETE("/acl/:id", apihandler.DeleteACL)

	// Firewall — DNAT (port forwarding)
	fw.GET("/dnat", apihandler.ListDNAT)
	fw.POST("/dnat", apihandler.AddDNAT)
	fw.PUT("/dnat", apihandler.EditDNAT)
	fw.DELETE("/dnat/:id", apihandler.DeleteDNAT)

	// Firewall — IP group
	fw.GET("/ip-group", apihandler.ListIPGroup)
	fw.POST("/ip-group", apihandler.AddIPGroup)
	fw.PUT("/ip-group", apihandler.EditIPGroup)
	fw.DELETE("/ip-group/:ids", apihandler.DeleteIPGroup)

	// Firewall — IPv6 group
	fw.GET("/ipv6-group", apihandler.ListIPv6Group)

	// Firewall — stream IP/port
	fw.GET("/stream-ipport", apihandler.ListStreamIPPort)
	fw.POST("/stream-ipport", apihandler.AddStreamIPPort)
	fw.PUT("/stream-ipport", apihandler.EditStreamIPPort)
	fw.DELETE("/stream-ipport/:ids", apihandler.DeleteStreamIPPort)

	// Firewall — connection limit
	fw.GET("/conn-limit", apihandler.ListConnLimit)
	fw.POST("/conn-limit", apihandler.AddConnLimit)
	fw.PUT("/conn-limit", apihandler.EditConnLimit)
	fw.DELETE("/conn-limit/:id", apihandler.DeleteConnLimit)

	// Sync
	sync := v1.Group("/sync")
	sync.GET("/status", apihandler.GetSyncStatus)
	sync.POST("/custom-isp", apihandler.TriggerCustomISPSync)
	sync.POST("/stream-domain", apihandler.TriggerStreamDomainSync)

	// Network — WAN / LAN
	net := v1.Group("/network")
	net.GET("/wan", apihandler.ListWan)
	net.GET("/lan", apihandler.ListLan)

	// Network — DHCP
	net.GET("/dhcp/leases", apihandler.ListDHCPLeases)
	net.GET("/dhcp/static", apihandler.ListDHCPStatic)
	net.POST("/dhcp/static", apihandler.AddDHCPStatic)
	net.PUT("/dhcp/static", apihandler.EditDHCPStatic)
	net.DELETE("/dhcp/static/:id", apihandler.DeleteDHCPStatic)

	// Network — DNS static
	net.GET("/dns/static", apihandler.ListDNSStatic)
	net.POST("/dns/static", apihandler.AddDNSStatic)
	net.PUT("/dns/static", apihandler.EditDNSStatic)
	net.DELETE("/dns/static/:id", apihandler.DeleteDNSStatic)

	// Network — static routes
	net.GET("/route/static", apihandler.ListRouteStatic)
	net.POST("/route/static", apihandler.AddRouteStatic)
	net.PUT("/route/static", apihandler.EditRouteStatic)
	net.DELETE("/route/static/:id", apihandler.DeleteRouteStatic)

	// VPN — PPTP clients
	vpn := v1.Group("/vpn")
	vpn.GET("/pptp", apihandler.ListPPTPClients)
	vpn.POST("/pptp", apihandler.AddPPTPClient)
	vpn.PUT("/pptp", apihandler.EditPPTPClient)
	vpn.DELETE("/pptp/:id", apihandler.DeletePPTPClient)

	// VPN — L2TP clients
	vpn.GET("/l2tp", apihandler.ListL2TPClients)
	vpn.POST("/l2tp", apihandler.AddL2TPClient)
	vpn.PUT("/l2tp", apihandler.EditL2TPClient)
	vpn.DELETE("/l2tp/:id", apihandler.DeleteL2TPClient)

	// WebUI
	h.GET("/ui", apihandler.WebUI)
}

func initMetrics(cfg *conf.Config) *prometheus.Registry {
	registry := prometheus.NewRegistry()
	m := ikuai.Get()
	if m != nil && m.Client() != nil {
		routerLabel := routerLabelFromCfg(cfg)
		collector := exportermetrics.NewCollector(cfg.Metrics.Namespace, routerLabel, m.Client())
		if err := registry.Register(collector); err != nil {
			logger.Error(fmt.Sprintf("failed to register metrics collector: %v", err))
		}
	}
	return registry
}

// routerLabelFromCfg derives a default "router" label value from the ikuai
// base_url. Returns "default" when the URL cannot be parsed.
func routerLabelFromCfg(cfg *conf.Config) string {
	if u, err := neturl.Parse(cfg.IKuai.BaseURL); err == nil && u.Hostname() != "" {
		return u.Hostname()
	}
	return "default"
}

func startMetricsServer(cfg *conf.Config, registry *prometheus.Registry) {
	metricsPath := cfg.Metrics.Path
	if metricsPath == "" {
		metricsPath = "/metrics"
	}
	port := cfg.Metrics.Port
	if port == 0 {
		port = 9100
	}

	mux := http.NewServeMux()
	mux.Handle(metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<html><head><title>iKuai Exporter</title></head>`+
			`<body><h1>iKuai Prometheus Exporter</h1>`+
			`<p><a href='%s'>Metrics</a></p></body></html>`, metricsPath)
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, port)
	go func() {
		logger.Info(fmt.Sprintf("metrics server listening on %s%s", addr, metricsPath))
		if err := http.ListenAndServe(addr, mux); err != nil {
			logger.Error(fmt.Sprintf("metrics server error: %v", err))
		}
	}()
}

// Cleanup releases all resources gracefully.
func Cleanup() {
	job.Stop()
	if m := ikuai.Get(); m != nil {
		m.Close()
	}
	logger.Sync()
	_ = db.Close()
}
