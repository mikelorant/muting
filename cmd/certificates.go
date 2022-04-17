package cmd

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mikelorant/muting/pkg/certificates"
	"github.com/mikelorant/muting/pkg/mutationconfig"
)

type CertificatesConfig struct {
    Name 			string `mapstructure:"name"`
		Namespace	string `mapstructure:"namespace"`
		Service		string `mapstructure:"service"`
		Output		string `mapstructure:"output"`
}

var (
	certificatesCmd = &cobra.Command{
		Use:   "certificates",
		Short: "Generate webhook certificates",
		Run: func(cmd *cobra.Command, args []string) {
			doCertificates()
		},
	}

	certificatesConfig CertificatesConfig
)

func init() {
	cobra.OnInitialize(initCertificatesConfig)
	rootCmd.AddCommand(certificatesCmd)
	certificatesCmd.Flags().StringP("name", "n", "muting", "Mutation configuration name")
	certificatesCmd.Flags().StringP("namespace", "", "default", "Webhook namespace")
	certificatesCmd.Flags().StringP("service", "s", "muting", "Webhook service")
	certificatesCmd.Flags().StringP("output", "o", "/tmp/tls", "Output directory")
	viper.SetEnvPrefix("cert")
	viper.AutomaticEnv()
	viper.BindPFlag("name", certificatesCmd.Flags().Lookup("name"))
	viper.BindPFlag("namespace", certificatesCmd.Flags().Lookup("namespace"))
	viper.BindPFlag("service", certificatesCmd.Flags().Lookup("service"))
	viper.BindPFlag("output", certificatesCmd.Flags().Lookup("output"))
}

func initCertificatesConfig() {
	if err := viper.Unmarshal(&certificatesConfig); err != nil {
			log.Fatal(err)
	}
}

func doCertificates() {
	fmt.Printf("%+v", &certificatesConfig)

	commonName := certificatesConfig.Service + "." + certificatesConfig.Namespace + ".svc"
	dnsNames := []string{
		certificatesConfig.Service,
		certificatesConfig.Service + "." + certificatesConfig.Namespace,
		certificatesConfig.Service + "." + certificatesConfig.Namespace + ".svc",
	}

	// Used for local debugging.
	// commonName := "host.minikube.internal"
	// dnsNames := []string {
	// 	commonName,
	// 	certificatesConfig.Service,
	// 	certificatesConfig.Service+"."+certificatesConfig.Namespace,
	// 	certificatesConfig.Service+"."+certificatesConfig.Namespace+".svc",
	// }

	caConfig := certificates.GenerateCA()
	serverConfig := certificates.GenerateServerCertificate(commonName, dnsNames, caConfig)

	err := certificates.WriteCertificates(certificatesConfig.Output, caConfig, serverConfig)
	if err != nil {
		log.Panic(err)
	}

	client := mutationconfig.CreateClient()
	mutateConfig := mutationconfig.GenerateMutationConfig(certificatesConfig.Name, certificatesConfig.Namespace, certificatesConfig.Service, caConfig.CAPEM)
	err = mutationconfig.ApplyMutationConfig(client, certificatesConfig.Name, mutateConfig)
	if err != nil {
		log.Panic(err)
	}
}

func (c CertificatesConfig) String() string {
		formatting := heredoc.Doc(`
			Name: %s
			Namespace: %s
			Service: %s
			Output: %s
		`)
		return fmt.Sprintf(formatting, c.Name, c.Namespace, c.Service, c.Output)
}
