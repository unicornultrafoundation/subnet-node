package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/unicornultrafoundation/subnet-node/internal/api"
	"github.com/unicornultrafoundation/subnet-node/repo/snrepo"
	"gopkg.in/yaml.v2"
)

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

func ensureDefaultConfig(configPath string) error {
	// Get directory and create default config path
	dir := filepath.Dir(configPath)
	defaultConfigPath := filepath.Join(dir, "config_default.yaml")

	// Check if default config already exists
	if _, err := os.Stat(defaultConfigPath); os.IsNotExist(err) {
		// Read current config
		content, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %v", err)
		}

		// Write to default config file
		if err := os.WriteFile(defaultConfigPath, content, 0644); err != nil {
			return fmt.Errorf("failed to create default config file: %v", err)
		}
	}
	return nil
}

func resetConfig(configPath string) error {
	defaultConfigPath := filepath.Join(filepath.Dir(configPath), "config_default.yaml")

	// Check if default config exists
	if _, err := os.Stat(defaultConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("no default config file found at %s", defaultConfigPath)
	}

	// Create backup of current config
	if err := backupConfig(configPath); err != nil {
		return err
	}

	// Read default config content
	content, err := os.ReadFile(defaultConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read default config file: %v", err)
	}

	// Reset config file to default values
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		return fmt.Errorf("failed to reset config file: %v", err)
	}

	return nil
}

func backupConfig(configPath string) error {
	// Read current config
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// Create backup file path
	backupPath := filepath.Join(filepath.Dir(configPath), "config_backup.yaml")

	// Write to backup file
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}

	return nil
}

func EditConfig(dataPath string, args []string) error {
	// Create a new flagset for edit-config subcommand
	editFlags := flag.NewFlagSet("edit-config", flag.ExitOnError)

	var setFlags configSetFlag
	showFlag := editFlags.Bool("show", false, "Show current configuration")
	resetFlag := editFlags.Bool("reset", false, "Reset configuration to default values")
	editFlags.Var(&setFlags, "set", "Set configuration value (format: key=value, can be used multiple times)")

	if err := editFlags.Parse(args); err != nil {
		return err
	}

	// Check if repo is initialized
	if !snrepo.IsInitialized(dataPath) {
		return fmt.Errorf("subnet node is not initialized. Please run 'subnet --datadir %s init' first", dataPath)
	}

	// Set config file path explicitly
	configFile := filepath.Join(dataPath, "config.yaml")

	// Ensure default config exists
	if err := ensureDefaultConfig(configFile); err != nil {
		return err
	}

	r, err := snrepo.Open(dataPath, &configFile)
	if err != nil {
		return fmt.Errorf("failed to open repo: %v", err)
	}
	defer r.Close()

	// Create config API instance
	configAPI := api.NewConfigAPI(r)

	// Handle reset configuration
	if *resetFlag {
		if err := resetConfig(configFile); err != nil {
			return err
		}

		fmt.Println("Configuration has been reset to default values")
		fmt.Printf("Previous configuration backed up to: %s\n", filepath.Join(dataPath, "config_backup.yaml"))

		// Show the reset configuration
		yamlBytes, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read config: %v", err)
		}

		fmt.Println("\nReset configuration:")
		fmt.Println(string(yamlBytes))
		return nil
	}

	// Handle show configuration
	if *showFlag {
		yamlBytes, err := yaml.Marshal(r.Config().Settings)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %v", err)
		}

		fmt.Println("Current configuration:")
		fmt.Println(string(yamlBytes))
		return nil
	}

	// Handle set configuration values
	if len(setFlags) > 0 {
		// Create backup before modification
		if err := backupConfig(configFile); err != nil {
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
		fmt.Printf("Previous configuration backed up to: %s\n", filepath.Join(dataPath, "config_backup.yaml"))

		// Show the updated configuration
		yamlBytes, err := yaml.Marshal(r.Config().Settings)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %v", err)
		}

		fmt.Println("\nUpdated configuration:")
		fmt.Println(string(yamlBytes))
	}

	// If no flags specified, print usage
	if !*showFlag && !*resetFlag && len(setFlags) == 0 {
		fmt.Println("Usage of edit-config:")
		fmt.Println("  Show current configuration:")
		fmt.Println("    subnet --datadir ./.data edit-config --show")
		fmt.Println("\n  Update configuration values:")
		fmt.Println("    subnet --datadir ./.data edit-config --set key=value [--set key2=value2 ...]")
		fmt.Println("\n  Reset to default configuration:")
		fmt.Println("    subnet --datadir ./.data edit-config --reset")
		fmt.Println("\n  Examples:")
		fmt.Println("    subnet --datadir ./.data edit-config --set addresses.api=/ip4/0.0.0.0/tcp/8080")
		fmt.Println("    subnet --datadir ./.data edit-config --set server.port=9090 --set debug=true")
	}

	return nil
}

func parseValue(value string) interface{} {
	// Try to parse as boolean
	switch strings.ToLower(value) {
	case "true", "yes", "1":
		return true
	case "false", "no", "0":
		return false
	}

	// Return as string
	return value
}
