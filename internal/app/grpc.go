package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net"
	"os"
	"time"

	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"

	"github.com/mechta-market/e-product/internal/config"
	"github.com/mechta-market/e-product/internal/errs"
	"github.com/mechta-market/e-product/pkg/proto/common"
)

type GrpcServer struct {
	name   string
	server *grpc.Server
}

func NewGrpcServer(name string, register func(*grpc.Server)) *GrpcServer {
	interceptors := make([]grpc.UnaryServerInterceptor, 0, 4)

	// ctx without cancel
	interceptors = append(interceptors, GrpcInterceptorCtxWithoutCancel())

	// error
	interceptors = append(interceptors, GrpcInterceptorError())

	// tracing
	if config.Conf.WithTracing {
		interceptors = append(interceptors, GrpcInterceptorTracing())
	}

	// metrics
	if config.Conf.WithMetrics {
		interceptors = append(interceptors, GrpcInterceptorMetrics())
	}

	// server
	server := grpc.NewServer(
		grpc.MaxSendMsgSize(math.MaxUint32),
		grpc.MaxRecvMsgSize(math.MaxUint32),
		grpc.ChainUnaryInterceptor(interceptors...),
	)

	// register handlers
	if register != nil {
		register(server)
	}

	// register grpc reflection
	reflection.Register(server)

	return &GrpcServer{
		name:   name,
		server: server,
	}
}

func (s *GrpcServer) Start() error {
	lis, err := net.Listen("tcp", ":"+config.Conf.GrpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen grpc: %w", err)
	}
	go func() {
		err = s.server.Serve(lis)
		if err != nil {
			slog.Error(s.name + "-grpc-server stopped: " + err.Error())
			os.Exit(1)
		}
	}()
	slog.Info(s.name + "-grpc-server started " + lis.Addr().String())
	return nil
}

func (s *GrpcServer) Stop() {
	s.server.GracefulStop()
}

func GrpcInterceptorCtxWithoutCancel() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		return handler(context.WithoutCancel(ctx), req)
	}
}

func GrpcInterceptorTracing() grpc.UnaryServerInterceptor {
	tracer := opentracing.GlobalTracer()

	return otgrpc.OpenTracingServerInterceptor(
		tracer,
		otgrpc.SpanDecorator(func(ctx context.Context, span opentracing.Span, method string, req, resp any, err error) {
			if err != nil {
				span.SetTag("error", true)
			}
		}),
	)
}

func GrpcInterceptorMetrics() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()

		h, err := handler(ctx, req)

		responseStatus := "ok"
		if err != nil {
			responseStatus = "error"
		}

		metricRequestCounter.WithLabelValues("grpc", info.FullMethod, responseStatus).Inc()
		metricResponseDuration.WithLabelValues("grpc", info.FullMethod, responseStatus).Observe(time.Since(start).Seconds())

		return h, err
	}
}

func GrpcInterceptorError() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		h, err := handler(ctx, req)
		if err == nil {
			return h, nil
		}

		var ei protoadapt.MessageV1
		errStr := err.Error()

		var errBase errs.Err
		if errors.As(err, &errBase) { // constant.Err
			ei = &common.ErrorRep{
				Code:    errBase.Error(),
				Message: errStr,
			}
		} else {
			var errFull errs.ErrFull
			if errors.As(err, &errFull) { // constant.ErrFull
				ei = &common.ErrorRep{
					Code:    errFull.Err.Error(),
					Message: errFull.Desc,
					Fields:  errFull.Fields,
				}
			}
		}
		if ei == nil {
			ei = &common.ErrorRep{
				Code:    errs.ServiceNA.Error(),
				Message: errStr,
			}
		}

		slog.Info(
			"GRPC handler error",
			slog.String("error", errStr),
			slog.String("method", info.FullMethod),
		)

		st := status.New(codes.InvalidArgument, errStr)
		st, err = st.WithDetails(ei)
		if err != nil {
			slog.Error(
				"error while creating status with details",
				slog.String("error", errStr),
				slog.String("method", info.FullMethod),
			)
			st = status.New(codes.InvalidArgument, errStr)
		}

		return h, st.Err()
	}
}
