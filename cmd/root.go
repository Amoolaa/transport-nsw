package cmd

import (
	"log/slog"
	"os"
	"transport-nsw-exporter/internal/config"
	"transport-nsw-exporter/internal/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var rootCmd = &cobra.Command{
	Use:   "transport-nsw-exporter",
	Short: "prometheus exporter for transport NSW data",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		var config config.Config
		if err := viper.Unmarshal(&config); err != nil {
			logger.Error("unable to decode config into struct", "error", err)
			os.Exit(1)
		}

		if len(config.Collectors.Carpark.FacilityIDs) == 0 {
			logger.Error("no car park facility IDs provided in config")
			os.Exit(1)
		}

		if os.Getenv("TRANSPORT_NSW_API_TOKEN") == "" {
			logger.Error("TRANSPORT_NSW_API_TOKEN env var must be set")
			os.Exit(1)
		}

		addr := viper.GetString("web.listen-address")

		s := server.New(addr, logger, config)
		if err := s.Run(); err != nil {
			logger.Error("error in server", "error", err)
		}
	},
}

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config-file", "", "path to config file")
	rootCmd.MarkPersistentFlagRequired("config-file")

	rootCmd.PersistentFlags().String("web.listen-address", ":8080", "Address on which to expose metrics")
	viper.BindPFlag("web.listen-address", rootCmd.PersistentFlags().Lookup("web.listen-address"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigType("yaml")
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AutomaticEnv()
	}

	if err := viper.ReadInConfig(); err == nil {
		slog.Info("Using config file", "path", viper.ConfigFileUsed())
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("error while executing", "error", err)
		os.Exit(1)
	}
}
