package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/MakeNowJust/heredoc"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mikelorant/muting/pkg/mutator"
)

type ServerConfig struct {
	Bind        string `mapstructure:"bind"`
	Sources     string `mapstructure:"sources"`
	Target      string `mapstructure:"target"`
	Certificate string `mapstructure:"certificate"`
	Key         string `mapstructure:"key"`
}

var (
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Webhook server.",
		Run: func(cmd *cobra.Command, args []string) {
			doServer()
		},
	}

	serverConfig ServerConfig
)

func init() {
	cobra.OnInitialize(initServerConfig)
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringP("bind", "b", ":6883", "Address to bind")
	serverCmd.Flags().StringP("sources", "s", "", "Source domains")
	serverCmd.Flags().StringP("target", "t", "", "Target domain")
	serverCmd.Flags().StringP("certificate", "c", "/tmp/tls/tls.crt", "Certificate file")
	serverCmd.Flags().StringP("key", "k", "/tmp/tls/tls.key", "Key file")
	// https://github.com/spf13/viper/issues/397
	// serverCmd.MarkFlagRequired("sources")
	// serverCmd.MarkFlagRequired("target")
}

func initServerConfig() {
	viper.SetEnvPrefix("server")
	viper.AutomaticEnv()
	viper.BindPFlag("bind", serverCmd.Flags().Lookup("bind"))
	viper.BindPFlag("sources", serverCmd.Flags().Lookup("sources"))
	viper.BindPFlag("target", serverCmd.Flags().Lookup("target"))
	viper.BindPFlag("certificate", serverCmd.Flags().Lookup("certificate"))
	viper.BindPFlag("key", serverCmd.Flags().Lookup("key"))

	if err := viper.Unmarshal(&serverConfig); err != nil {
		log.Fatal(err)
	}
}

func doServer() {
	fmt.Printf("%+v", &serverConfig)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/health", health)
	e.Any("/mutate", mutate)

	if err := e.StartTLS(serverConfig.Bind, serverConfig.Certificate, serverConfig.Key); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func health(c echo.Context) error {
	return c.String(http.StatusOK, "success")
}

func mutate(c echo.Context) error {
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, "malformed request")
	}

	mutated, err := mutator.Mutate(body, serverConfig.Sources, serverConfig.Target)
	if err != nil {
		if _, ok := err.(*mutator.BadRequest); ok {
			return c.String(http.StatusBadRequest, "bad request")
		} else {
			return c.String(http.StatusInternalServerError, "internal server error")
		}
	}

	return c.JSONBlob(http.StatusOK, mutated)
}

func (c ServerConfig) String() string {
	formatting := heredoc.Doc(`
			Bind: %s
			Sources: %s
			Target: %s
			Certificate: %s
			Key: %s
		`)
	return fmt.Sprintf(formatting, c.Bind, c.Sources, c.Target, c.Certificate, c.Key)
}
