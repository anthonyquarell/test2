package config

import (
	"github.com/caarlos0/env/v9"
	_ "github.com/joho/godotenv/autoload"
)

var Conf = struct {
	Namespace     string `env:"NAMESPACE" envDefault:"example.com"`
	Debug         bool   `env:"DEBUG" envDefault:"false"`
	GrpcPort      string `env:"GRPC_PORT" envDefault:"5050"`
	HttpPort      string `env:"HTTP_PORT" envDefault:"80"`
	HttpCors      bool   `env:"HTTP_CORS" envDefault:"false"`
	WithMetrics   bool   `env:"WITH_METRICS" envDefault:"false"`
	WithTracing   bool   `env:"WITH_TRACING" envDefault:"false"`
	JaegerAddress string `env:"JAEGER_ADDRESS"`
	PgDsn         string `env:"PG_DSN"`

	MdmUrl   string `env:"MDM_URL"`
	MdmToken string `env:"MDM_TOKEN"`

	ComportalUrl      string `env:"COMPORTAL_URL"`
	ComportalUsername string `env:"COMPORTAL_USERNAME"`
	ComportalPassword string `env:"COMPORTAL_PASSWORD"`

	AsbisUrl         string `env:"ASBIS_URL"`
	AsbisUsername    string `env:"ASBIS_USERNAME"`
	AsbisPassword    string `env:"ASBIS_PASSWORD"`
	AsbisP12CertPath string `env:"ASBIS_P12_CERT_PATH"`
	AsbisP12Password string `env:"ASBIS_P12_PASSWORD"`
	AsbisCaCertPath  string `env:"ASBIS_CA_CERT_PATH"`

	MegogoUrl      string `env:"MEGOGO_URL"`
	MegogoUsername string `env:"MEGOGO_USERNAME"`
	MegogoPassword string `env:"MEGOGO_PASSWORD"`
}{}

func init() {
	if err := env.Parse(&Conf); err != nil {
		panic(err)
	}
}
