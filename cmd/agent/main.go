package main

import (
	"os"

	"os-artificer/saber/internal/agent"
	"os-artificer/saber/pkg/logger"

	"github.com/spf13/cobra"
)

func main() {
	if os.Getenv("SABER_AGENT_SUPERVISOR") == "1" {
		if err := agent.RunSupervisor(); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
		return
	}

	rootCmd := &cobra.Command{
		Use:          "Agent",
		Short:        "Saber Agent",
		SilenceUsage: true,
		RunE:         agent.Run,
	}

	rootCmd.PersistentFlags().StringVarP(&agent.ConfigFilePath, "config", "c", "./etc/agent.yaml", "")
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(agent.StartCmd)
	rootCmd.AddCommand(agent.StopCmd)
	rootCmd.AddCommand(agent.RestartCmd)
	rootCmd.AddCommand(agent.ReloadCmd)
	rootCmd.AddCommand(agent.VersionCmd)
	rootCmd.AddCommand(agent.HealthCheckCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Errorf("failed to start agent server. errmsg:%s", err.Error())
		return
	}
}
