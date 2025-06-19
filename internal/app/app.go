package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/mechta-market/e-product/internal/constant"
	domainKeyServiceP "github.com/mechta-market/e-product/internal/domain/key"
	domainKeyRepoDbP "github.com/mechta-market/e-product/internal/domain/key/repo/pg"
	handlerGrpcP "github.com/mechta-market/e-product/internal/handler/grpc"
	serviceMdmP "github.com/mechta-market/e-product/internal/service/mdm"
	serviceMdmRepoP "github.com/mechta-market/e-product/internal/service/mdm/repo"
	serviceAsbP "github.com/mechta-market/e-product/internal/service/provider/asbis"
	serviceAsbRepoP "github.com/mechta-market/e-product/internal/service/provider/asbis/repo"
	serviceCptP "github.com/mechta-market/e-product/internal/service/provider/comportal"
	serviceCptRepoP "github.com/mechta-market/e-product/internal/service/provider/comportal/repo"
	serviceMegogoP "github.com/mechta-market/e-product/internal/service/provider/megogo"
	serviceMegogoRepoP "github.com/mechta-market/e-product/internal/service/provider/megogo/repo"
	usecaseKeyP "github.com/mechta-market/e-product/internal/usecase/key"
	e_product_v1 "github.com/mechta-market/e-product/pkg/proto/e_product"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/mechta-market/e-product/internal/config"
)

type App struct {
	globalTracerCloser io.Closer

	pgpool *pgxpool.Pool

	grpcServer *GrpcServer
	httpServer *http.Server

	ctx       context.Context
	ctxCancel context.CancelFunc

	exitCode int
}

func (a *App) Init() {
	var err error

	a.ctx, a.ctxCancel = context.WithCancel(context.Background())

	var keyService *domainKeyServiceP.Service
	var mdmService *serviceMdmP.Service
	var cptService *serviceCptP.Service
	var asbisService *serviceAsbP.Service
	var megogoService *serviceMegogoP.Service

	var keyUsecase *usecaseKeyP.Usecase

	var handlerGrpcKey *handlerGrpcP.Key

	// logger
	{
		if !config.Conf.Debug {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			slog.SetDefault(logger)
		}
	}

	// globalTracer
	{
		if config.Conf.WithTracing && config.Conf.JaegerAddress != "" {
			slog.Info("tracing enabled")
			_, a.globalTracerCloser, err = tracerInitGlobal(config.Conf.JaegerAddress, constant.ServiceName)
			errCheck(err, "tracerInitGlobal")
		}
	}

	// pgpool
	{
		pgConf, err := pgxpool.ParseConfig(config.Conf.PgDsn)
		errCheck(err, "pgxpool.ParseConfig")

		pgConf.MaxConns = 10
		pgConf.MinConns = 2
		pgConf.MaxConnLifetime = 3 * time.Minute
		pgConf.MaxConnIdleTime = time.Minute
		pgConf.HealthCheckPeriod = 15 * time.Second

		a.pgpool, err = pgxpool.NewWithConfig(context.Background(), pgConf)
		errCheck(err, "pgxpool.NewWithConfig")
	}

	// http client connections
	{
	}

	// mdm
	{
		var repo serviceMdmP.RepoI
		repo = serviceMdmRepoP.New(config.Conf.MdmUrl, config.Conf.MdmToken)
		mdmService = serviceMdmP.New(repo)
	}

	// comportal
	{
		var repo serviceCptP.RepoI
		repo = serviceCptRepoP.New(config.Conf.ComportalUrl, config.Conf.ComportalUsername, config.Conf.ComportalPassword)
		cptService = serviceCptP.New(repo)
	}

	// asbis
	{
		var repo serviceAsbP.RepoI
		repo = serviceAsbRepoP.New(config.Conf.AsbisUrl, config.Conf.AsbisUsername,
			config.Conf.AsbisPassword, config.Conf.AsbisP12CertPath, config.Conf.AsbisP12Password, config.Conf.AsbisCaCertPath)
		asbisService = serviceAsbP.New(repo)
	}

	// megogo
	{
		var repo serviceMegogoP.RepoI
		repo = serviceMegogoRepoP.New(config.Conf.MegogoUrl, config.Conf.MegogoUsername, config.Conf.MegogoPassword)
		megogoService = serviceMegogoP.New(repo)
	}

	// provider manager
	providers := make(map[string]usecaseKeyP.ProviderServiceI, 3)

	if cptService != nil {
		providers[constant.ProviderComportal] = cptService
		slog.Info("Comportal provider registered successfully",
			"provider_id", constant.ProviderComportal)
	}

	if asbisService != nil {
		providers[constant.ProviderASBIS] = asbisService
		slog.Info("ASBIS provider registered successfully",
			"provider_id", constant.ProviderASBIS)
	}

	if megogoService != nil {
		providers[constant.ProviderMegogo] = megogoService
		slog.Info("Megogo provider registered successfully",
			"provider_id", constant.ProviderMegogo)
	}

	// key
	{
		repoShop := domainKeyRepoDbP.New(a.pgpool)
		keyService = domainKeyServiceP.New(repoShop)
		keyUsecase = usecaseKeyP.New(keyService, mdmService, providers)
		handlerGrpcKey = handlerGrpcP.NewKey(keyUsecase)
	}

	// grpc server
	{
		a.grpcServer = NewGrpcServer("main", func(server *grpc.Server) {
			e_product_v1.RegisterKeyServer(server, handlerGrpcKey)
		})
	}

	// http-gw server
	{
		var handler http.Handler

		handler, err = GrpcGatewayCreateHandler(func(mux *runtime.ServeMux) error {
			opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

			var conn *grpc.ClientConn
			conn, err = grpc.NewClient("localhost:"+config.Conf.GrpcPort, opts...)
			errCheck(err, "grpc.Dial")

			// register grpc handlers
			handlers := []func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error{
				e_product_v1.RegisterKeyHandler,
			}
			for _, h := range handlers {
				err = h(context.Background(), mux, conn)
				if err != nil {
					return fmt.Errorf("grpc-gateway: register grpc-handler: %w", err)
				}
			}

			// http handlers
			httpHandlers := []struct {
				method  string
				path    string
				handler runtime.HandlerFunc
			}{
				{
					"GET", "/tst", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
						slog.Error("test error", "error", errors.New("test error"))
					},
				},
				// examples:
				// {"POST", "/route/register", handlerHttpRouteRegister.Register},
				// {"GET", "/route/{id}/link", handlerHttpRouteRegister.GetLink},
				// {"POST", "/ord_shop_change", handlerHttpOrdShopChange.Set},
			}
			for _, h := range httpHandlers {
				err = mux.HandlePath(h.method, h.path, h.handler)
				if err != nil {
					return fmt.Errorf("grpc-gateway: register http-handler: %w", err)
				}
			}

			return nil
		})
		errCheck(err, "grpcGatewayCreateHandler")

		// server
		a.httpServer = &http.Server{
			Addr:              ":" + config.Conf.HttpPort,
			Handler:           handler,
			ReadHeaderTimeout: 2 * time.Second,
			ReadTimeout:       time.Minute,
			MaxHeaderBytes:    300 * 1024,
		}
	}
}

func (a *App) PreStartHook() {
	slog.Info("PreStartHook")
}

func (a *App) Start() {
	slog.Info("Starting")

	// grpc server
	{
		err := a.grpcServer.Start()
		errCheck(err, "grpcServer.Start")
	}

	// http-gw server
	{
		go func() {
			err := a.httpServer.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				// errCheck(err, "http-server stopped")
			}
		}()
		slog.Info("http-server started " + a.httpServer.Addr)
	}
}

func (a *App) Listen() {
	signalCtx, signalCtxCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer signalCtxCancel()

	// wait signal
	<-signalCtx.Done()
}

func (a *App) Stop() {
	slog.Info("Shutting down...")

	// stop context
	a.ctxCancel()

	// http-gw server
	{
		ctx, ctxCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer ctxCancel()

		if err := a.httpServer.Shutdown(ctx); err != nil {
			slog.Error("http-server shutdown error", "error", err)
			a.exitCode = 1
		}
	}

	// grpc server
	a.grpcServer.Stop()
}

func (a *App) WaitJobs() {
	slog.Info("waiting jobs")
}

func (a *App) Exit() {
	slog.Info("Exit")

	if a.globalTracerCloser != nil {
		_ = a.globalTracerCloser.Close()
	}

	// flush stdout

	os.Exit(a.exitCode)
}

func errCheck(err error, msg string) {
	if err != nil {
		if msg != "" {
			err = fmt.Errorf("%s: %w", msg, err)
		}
		slog.Error(err.Error())
		os.Exit(1)
	}
}
