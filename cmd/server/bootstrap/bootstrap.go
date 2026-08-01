package bootstrap

import (
	"context"
	"fmt"
	neturl "net/url"
	"net/http"
	"os"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	exportermetrics "github.com/zy84338719/ikuai_exporter/metrics"
	hertzrouter "github.com/zy84338719/ikuai-tools-service/gen/http/router"
	routersvc "github.com/zy84338719/ikuai-tools-service/internal/app/router"
	"github.com/zy84338719/ikuai-tools-service/internal/conf"
	"github.com/zy84338719/ikuai-tools-service/internal/ikuai"
	"github.com/zy84338719/ikuai-tools-service/internal/job"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
	"github.com/zy84338719/ikuai-tools-service/internal/repo/db"
	"github.com/zy84338719/ikuai-tools-service/internal/repo/redis"
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

	if err := initRedis(cfg); err != nil {
		// Redis is optional for small deployments — warn but continue.
		logger.Error(fmt.Sprintf("init redis failed (continuing without cache): %v", err))
	}

	if err := initIKuaiManager(cfg); err != nil {
		logger.Error(fmt.Sprintf("init ikuai client failed (continuing): %v", err))
	}

	// Hydrate the router registry from the DB (routers added via the API).
	loadRoutersFromDB()

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

// initRedis connects to Redis. It returns an error on failure but the caller
// treats it as non-fatal so the service still runs without a cache backend.
func initRedis(cfg *conf.Config) error {
	return redis.Init(&cfg.Redis)
}

// loadRoutersFromDB seeds the connection registry with every enabled router
// stored in the database, so routers configured through the API survive a
// restart. The legacy single-router config block (ikuai.Init) already
// registered a "default" entry when a token was present.
func loadRoutersFromDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	svc := routersvc.NewService()
	rs, err := svc.AllForManager(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("load routers from DB failed: %v", err))
		return
	}
	specs := make([]ikuai.RouterSpec, 0, len(rs))
	for _, r := range rs {
		// Skip "default" — already registered from the config block to avoid
		// shadowing it with a stale DB copy on first run.
		if r.Name == "default" {
			continue
		}
		specs = append(specs, ikuai.RouterSpec{
			Name: r.Name, BaseURL: r.BaseURL, Token: r.Token,
			Insecure: r.Insecure, Timeout: r.Timeout,
		})
	}
	ikuai.GetRegistry().LoadAll(specs)
}

func initServer(cfg *conf.Config) *server.Hertz {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	h := server.New(server.WithHostPorts(addr))

	h.Use(middleware.Recovery())
	h.Use(middleware.Logger())              // injects X-Request-ID into context
	h.Use(middleware.CORS(cfg.Server.CORSOrigins))
	h.Use(middleware.Auth(cfg.Auth.APIKey)) // no-op when API key is empty
	h.Use(middleware.Audit())               // records mutating requests

	hertzrouter.GeneratedRegister(h)
	registerIKuaiRoutes(h)
	registerAuthRoutes(h, cfg)

	return h
}

// registerAuthRoutes wires the login endpoint. With the API-key model login is
// a simple key-verification that returns the key back for the client to use as
// a Bearer token; it exists so the WebUI can have a login flow.
func registerAuthRoutes(h *server.Hertz, cfg *conf.Config) {
	if cfg.Auth.APIKey == "" {
		return
	}
	h.POST("/api/v1/auth/login", func(ctx context.Context, c *app.RequestContext) {
		var req struct {
			APIKey string `json:"api_key"`
		}
		if err := c.BindAndValidate(&req); err != nil {
			resp.BadRequest(c, err.Error())
			return
		}
		if req.APIKey != cfg.Auth.APIKey {
			resp.Unauthorized(c, "invalid api key")
			return
		}
		resp.Success(c, map[string]string{"token": req.APIKey})
	})
}

func registerIKuaiRoutes(h *server.Hertz) {
	// ── Router instance management (NOT scoped by :router_id) ──
	routers := h.Group("/api/v1/routers")
	routers.GET("", apihandler.ListRouters)
	routers.GET("/:name", apihandler.GetRouter)
	routers.POST("", apihandler.CreateRouter)
	routers.PUT("/:name", apihandler.UpdateRouter)
	routers.DELETE("/:name", apihandler.DeleteRouter)

	// ── Per-router resources (scoped by :router_id) ──
	v1 := h.Group("/api/v1/ikuai/:router_id")

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

	// Logs (read-only)
	logs := v1.Group("/log")
	logs.GET("/notice", apihandler.ListLogNotice)
	logs.GET("/system", apihandler.ListLogSystem)
	logs.GET("/auth", apihandler.ListLogAuth)
	logs.GET("/dhcp", apihandler.ListLogDhcp)
	logs.GET("/pppoe", apihandler.ListLogPppoe)
	logs.GET("/arp", apihandler.ListLogArp)
	logs.GET("/ddns", apihandler.ListLogDdns)
	logs.GET("/web-activity", apihandler.ListLogWebActivity)
	logs.GET("/wireless", apihandler.ListLogWireless)

	// System ops — backup
	bk := v1.Group("/system/backup")
	bk.GET("", apihandler.ListBackup)
	bk.POST("", apihandler.CreateBackup)
	bk.DELETE("/:id", apihandler.DeleteBackup)
	v1.POST("/system/backup/restore", apihandler.RestoreBackup)
	// System ops — upgrade
	v1.GET("/system/upgrade", apihandler.ListUpgrade)
	v1.POST("/system/upgrade/check", apihandler.CheckUpgrade)
	v1.POST("/system/upgrade/start", apihandler.StartUpgrade)
	v1.GET("/system/upgrade/status", apihandler.GetUpgradeStatus)
	// System ops — reboot schedules
	rb := v1.Group("/system/reboot")
	rb.GET("", apihandler.ListRebootSchedules)
	rb.POST("", apihandler.AddRebootSchedule)
	rb.PUT("", apihandler.EditRebootSchedule)
	rb.DELETE("/:id", apihandler.DeleteRebootSchedule)
	// System ops — remote access / disks / admin accounts
	v1.GET("/system/remote-access", apihandler.GetRemoteAccess)
	v1.PUT("/system/remote-access", apihandler.UpdateRemoteAccess)
	v1.GET("/system/disks", apihandler.GetDisks)
	wa := v1.Group("/system/web-admin")
	wa.GET("", apihandler.ListWebAdminAccounts)
	wa.POST("", apihandler.AddWebAdminAccount)
	wa.DELETE("/:id", apihandler.DeleteWebAdminAccount)

	// Network — DMZ
	dmz := net.Group("/dmz")
	dmz.GET("", apihandler.ListDMZ)
	dmz.POST("", apihandler.AddDMZ)
	dmz.PUT("", apihandler.EditDMZ)
	dmz.DELETE("/:id", apihandler.DeleteDMZ)
	// Network — NAT
	nat := net.Group("/nat")
	nat.GET("", apihandler.ListNAT)
	nat.POST("", apihandler.AddNAT)
	nat.PUT("", apihandler.EditNAT)
	nat.DELETE("/:id", apihandler.DeleteNAT)
	// Network — QoS by IP
	qip := net.Group("/qos/ip")
	qip.GET("", apihandler.ListQosIP)
	qip.POST("", apihandler.AddQosIP)
	qip.PUT("", apihandler.EditQosIP)
	qip.DELETE("/:id", apihandler.DeleteQosIP)
	// Network — QoS by MAC
	qmac := net.Group("/qos/mac")
	qmac.GET("", apihandler.ListQosMac)
	qmac.POST("", apihandler.AddQosMac)
	qmac.PUT("", apihandler.EditQosMac)
	qmac.DELETE("/:id", apihandler.DeleteQosMac)
	// Network — VLAN
	vlan := net.Group("/vlan")
	vlan.GET("", apihandler.ListVLAN)
	vlan.POST("", apihandler.AddVLAN)
	vlan.PUT("", apihandler.EditVLAN)

	// Security — MAC filter rules
	mac := fw.Group("/mac-rules")
	mac.GET("", apihandler.ListMacRules)
	mac.POST("", apihandler.AddMacRules)
	mac.PUT("", apihandler.EditMacRules)
	mac.DELETE("/:id", apihandler.DeleteMacRules)

	// Routing — load balance (multi-WAN)
	lb := v1.Group("/routing/load-balance")
	lb.GET("", apihandler.ListLoadBalance)
	lb.POST("", apihandler.AddLoadBalance)
	lb.PUT("", apihandler.EditLoadBalance)
	lb.DELETE("/:id", apihandler.DeleteLoadBalance)
	// Routing — app-protocol rules
	ap := v1.Group("/routing/app-protocols")
	ap.GET("", apihandler.ListAppProtocols)
	ap.POST("", apihandler.AddAppProtocols)
	ap.PUT("", apihandler.EditAppProtocols)
	ap.DELETE("/:id", apihandler.DeleteAppProtocols)

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
	_ = redis.Close()
	logger.Sync()
	_ = db.Close()
}
