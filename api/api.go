package api

import (
	"github.com/labstack/echo/v4"
	"libp2p-test-node/config"
	_ "net/http/pprof"

	"github.com/labstack/echo/v4/middleware"
)

func InitializeEchoServer(cfg config.NodeConfig) {

	// Echo instance
	e := echo.New()
	e.Use(middleware.Recover())

	baseGroup := e.Group("")
	baseGroup.GET("/connected-peers", func(c echo.Context) error {

		if cfg.Host == nil {
			return c.JSON(200, "Host not initialized")
		}

		if cfg.Host.Network() == nil {
			return c.JSON(200, "Network not initialized")
		}

		if cfg.Host.Network().Peers() == nil {
			return c.JSON(200, "Peers not initialized")
		}

		peers := cfg.Host.Network().Peers()
		return c.JSON(200, peers)
	})

	baseGroup.GET("/get-eth-prices", func(c echo.Context) error {

		db := cfg.DB
		var message config.Message
		db.Model(&config.Message{}).Find(&message)
		return c.JSON(200, message)
	})

	// Start server
	e.Logger.Print(e.Start("0.0.0.0:" + cfg.Api.Port)) // configuration
}
