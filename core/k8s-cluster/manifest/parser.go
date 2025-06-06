package manifest

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Parser represents an SDL parser
type Parser struct {
	// Add any parser-specific configuration here
}

// NewParser creates a new SDL parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile parses an SDL file from the given path
func (p *Parser) ParseFile(path string) (*SDL, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return p.Parse(file)
}

// Parse parses SDL data from a reader
func (p *Parser) Parse(r io.Reader) (*SDL, error) {
	// Read the entire file content
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Try parsing as YAML first
	sdl := &SDL{}
	if err := yaml.Unmarshal(data, sdl); err == nil {
		return sdl, nil
	} else {
		fmt.Printf("YAML parsing failed: %v\n", err)
	}

	// If YAML parsing fails, try JSON
	if err := json.Unmarshal(data, sdl); err != nil {
		return nil, fmt.Errorf("failed to parse SDL file: %w", err)
	}

	return sdl, nil
}

// ParseYAML parses SDL data from YAML format
func (p *Parser) ParseYAML(data []byte) (*SDL, error) {
	sdl := &SDL{}
	if err := yaml.Unmarshal(data, sdl); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return sdl, nil
}

// ParseJSON parses SDL data from JSON format
func (p *Parser) ParseJSON(data []byte) (*SDL, error) {
	sdl := &SDL{}
	if err := json.Unmarshal(data, sdl); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return sdl, nil
}

// ValidateAndParse parses and validates an SDL file
func (p *Parser) ValidateAndParse(path string) (*SDL, error) {
	sdl, err := p.ParseFile(path)
	if err != nil {
		return nil, err
	}

	if err := sdl.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SDL: %w", err)
	}

	return sdl, nil
}
