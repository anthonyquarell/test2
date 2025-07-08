package app

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mechta-market/e-product/internal/errs"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/goccy/go-json"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/mechta-market/e-product/internal/config"
)

func GrpcGatewayCreateHandler(muxHook func(*runtime.ServeMux) error) (http.Handler, error) {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
			var repBody []byte

			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.NotFound {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`service path not found`))
					return
				} else if st.Code() == codes.InvalidArgument && len(st.Details()) > 0 {
					var marshalErr error
					repBody, marshalErr = marshaler.Marshal(st.Details()[0])
					if marshalErr != nil {
						slog.Error("GRPC_GW: ErrorHandler: Failed to marshal", "error", marshalErr)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
			}

			if len(repBody) == 0 {
				//runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
				obj := map[string]string{
					"code":    errs.ServiceNA.Error(),
					"message": err.Error(),
				}
				repBody, err = json.Marshal(obj)
				if err != nil {
					slog.Error("GRPC_GW: ErrorHandler: Failed to marshal", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			//slog.Error("GRPC_GW: ErrorHandler", "error", err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, err = io.Copy(w, bytes.NewReader(repBody))
			if err != nil {
				slog.Error("GRPC_GW: ErrorHandler: Failed to write response", "error", err)
			}
			//_, _ = w.Write(repBody)
		}),
	)

	if muxHook != nil {
		err := muxHook(mux)
		if err != nil {
			return nil, fmt.Errorf("grpc-gateway: muxHook: %w", err)
		}
	}

	// add health check handler
	err := mux.HandlePath(http.MethodGet, "/healthcheck", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		w.WriteHeader(http.StatusOK)
	})
	if err != nil {
		return nil, fmt.Errorf("grpc-gateway: register healthcheck handler: %w", err)
	}

	// add docs handler
	docFS := http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs")))
	err = mux.HandlePath(http.MethodGet, "/docs/{any}", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		docFS.ServeHTTP(w, r)
	})
	if err != nil {
		return nil, fmt.Errorf("grpc-gateway: register docs handler: %w", err)
	}
	err = mux.HandlePath(http.MethodGet, "/docs/{any}/{any}", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		docFS.ServeHTTP(w, r)
	})
	if err != nil {
		return nil, fmt.Errorf("grpc-gateway: register docs handler: %w", err)
	}

	// add metrics handler
	if err = mux.HandlePath(http.MethodGet, "/metrics", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		promhttp.Handler().ServeHTTP(w, r)
	}); err != nil {
		return nil, fmt.Errorf("grpc-gateway: register metrics handler: %w", err)
	}

	handler := http.Handler(mux)

	// add cors middleware
	if config.Conf.HttpCors {
		handler = cors.New(cors.Options{
			AllowOriginFunc: func(origin string) bool { return true },
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodPut,
				http.MethodPost,
				http.MethodDelete,
			},
			AllowedHeaders: []string{
				"Accept",
				"Content-Type",
				"X-Requested-With",
				"Authorization",
			},
			AllowCredentials: true,
			MaxAge:           604800,
		}).Handler(handler)
	}

	// add recover middleware
	handler = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				// use always new err instance in defer
				if err := recover(); err != nil {
					slog.Error(
						"Recovered from panic",
						slog.Any("error", err),
						slog.Any("recovery_stacktrace", string(debug.Stack())),
					)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			h.ServeHTTP(w, r)
		})
	}(handler)

	return handler, nil
}
