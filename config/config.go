package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/labstack/gommon/log"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"gorm.io/gorm"
)

type NodeConfig struct {
	Common struct {
		PostgresDsn string `env:"DB_DSN" envDefault:"libp2p.db"`
		Version     string `env:"VERSION" envDefault:""`
		Commit      string `env:"COMMIT" envDefault:""`
	}

	PubSub struct {
		TopicName string `env:"TOPIC_NAME" envDefault:"eth-price"`
	}
	Peer struct {
		ListenPort     string `env:"LISTEN_PORT" envDefault:"4001"`
		ListenAddr     string `env:"LISTEN_ADDR" envDefault:""`
		BootstrapPeers string `env:"BOOTSTRAP_PEERS" envDefault:""`
	}

	Api struct {
		Port string `env:"API_PORT" envDefault:"8080"`
	}

	DB   *gorm.DB
	Host host.Host
	DHT  *dht.IpfsDHT
}

func InitConfig() NodeConfig {
	godotenv.Load() // load from environment OR .env file if it exists
	var cfg NodeConfig

	if err := env.Parse(&cfg); err != nil {
		log.Fatal("error parsing config: %+v\n", err)
	}
	db, err := initDb(cfg.Common.PostgresDsn)
	if err != nil {
		log.Fatal("error parsing config: %+v\n", err)
	}
	cfg.DB = db
	log.Debug("config parsed successfully")
	return cfg
}
