package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unicornultrafoundation/subnet-node/common/fsutil"
	"github.com/unicornultrafoundation/subnet-node/internal/api"
	"github.com/unicornultrafoundation/subnet-node/repo/snrepo"
)

var dataPath string

type configValue struct {
	key   string
	value string
}

type configSetFlag []configValue

func (f *configSetFlag) String() string {
	return fmt.Sprint(*f)
}

func (f *configSetFlag) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, expected key=value, got %s", value)
	}
	*f = append(*f, configValue{key: parts[0], value: parts[1]})
	return nil
}

func backupConfig(expPath string) error {
	// Read current config
	content, err := os.ReadFile(filepath.Join(expPath, "config.yaml"))
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	backupPath := filepath.Join(expPath, "config_backup.yaml")

	// Write to backup file
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}

	return nil
}

func EditConfig(dataPath string, setFlags configSetFlag) error {
	// Check if repo is initialized
	if !snrepo.IsInitialized(dataPath) {
		return fmt.Errorf("subnet node is not initialized. Please run 'subnet --datadir %s init' first", dataPath)
	}

	expPath, err := fsutil.ExpandHome(filepath.Clean(dataPath))
	if err != nil {
		return fmt.Errorf("failed to expand data path: %v", err)
	}

	configFile := filepath.Join(expPath, "config.yaml")

	r, err := snrepo.Open(dataPath, &configFile)
	if err != nil {
		return fmt.Errorf("failed to open repo: %v", err)
	}
	defer r.Close()

	// Create config API instance
	configAPI := api.NewConfigAPI(r)

	// Only handle set configuration values
	if len(setFlags) == 0 {
		return fmt.Errorf("no configuration values provided to set")
	}

	// Create backup before modification
	if err := backupConfig(expPath); err != nil {
		return err
	}

	updates := make(map[string]interface{})
	for _, setValue := range setFlags {
		updates[setValue.key] = parseValue(setValue.value)
	}

	if err := configAPI.Update(context.Background(), updates); err != nil {
		return fmt.Errorf("failed to update config: %v", err)
	}

	fmt.Println("Configuration updated successfully")
	fmt.Printf("Previous configuration backed up to: %s\n", filepath.Join(expPath, "config_backup.yaml"))

	return nil
}

func parseValue(value string) interface{} {
	// Try to parse as boolean
	switch strings.ToLower(value) {
	case "true", "yes":
		return true
	case "false", "no":
		return false
	}

	// Try to parse as integer
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}

	// Try to parse as float
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	// Return as string
	return value
}

var editConfigCmd = &cobra.Command{
	Use:   "set",
	Short: "Edit node configuration",
	Long:  "Edit the node configuration by setting key-value pairs or showing the current configuration.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var setFlags configSetFlag
		for _, arg := range args {
			if err := setFlags.Set(arg); err != nil {
				return err
			}
		}
		return EditConfig(dataPath, setFlags)
	},
}

var showConfigCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Display the current node configuration in YAML format.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		expPath, err := fsutil.ExpandHome(filepath.Clean(dataPath))
		if err != nil {
			return fmt.Errorf("failed to expand data path: %v", err)
		}

		configFile := filepath.Join(expPath, "config.yaml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("configuration file does not exist at %s", configFile)
		}

		content, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read config file: %v", err)
		}

		fmt.Println("Current configuration:")
		fmt.Println(string(content))
		return nil
	},
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage node configuration",
		Long:  "Commands to manage and edit the node configuration.",
	}

	cmd.PersistentFlags().StringVar(&dataPath, "datadir", "~/.subnet-node", "Path to data directory")

	cmd.AddCommand(editConfigCmd)
	cmd.AddCommand(showConfigCmd)

	cmd.Flags().StringVar(&dataPath, "datadir", "~/.subnet-node", "Path to data directory")
	return cmd
}
