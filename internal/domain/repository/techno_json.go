package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Medzoner/gomedz/pkg/logger"
	"github.com/Medzoner/medzoner-go/internal/config"
)

// TechnoJSONRepository is an implementation of TechnoRepository backed by JSON files.
type TechnoJSONRepository struct {
	RootPath string
}

// NewTechnoJSONRepository creates a new TechnoJSONRepository.
func NewTechnoJSONRepository(config config.Config) *TechnoJSONRepository {
	return &TechnoJSONRepository{
		RootPath: string(config.RootPath),
	}
}

// FetchStack reads and returns the stacks data from JSON file.
func (m *TechnoJSONRepository) FetchStack(ctx context.Context) (map[string]any, error) {
	jsonFile, err := os.Open(m.RootPath + "internal/resources/data/jobs/stacks.json")
	if err != nil {
		return nil, fmt.Errorf("error during open json file: %w", err)
	}
	defer func() {
		if cerr := jsonFile.Close(); cerr != nil {
			logger.Error(ctx, "jsonFile close error.", cerr)
		}
	}()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("error during read json file: %w", err)
	}

	result := make(map[string]any)
	if err = json.Unmarshal(byteValue, &result); err != nil {
		return nil, fmt.Errorf("error during unmarshal json: %w", err)
	}

	return result, nil
}
