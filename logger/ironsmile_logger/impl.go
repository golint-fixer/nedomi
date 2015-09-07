package ironsmile_logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ironsmile/nedomi/config"

	"github.com/ironsmile/logger"
)

// New returns configured ironsmile™ logger that is ready to use.
func New(cfg *config.LoggerSection) (*logger.Logger, error) {
	logger := logger.New()
	var s settings
	err := json.Unmarshal(cfg.Settings, &s)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing logger settings for 'ironsmile_logger':\n%s\n", err)
	}

	var errorOutput, debugOutput, logOutput io.Writer

	if s.DebugFile != "" {
		debugOutput, err = os.OpenFile(s.DebugFile, 0, 7770)
		if err != nil {
			return nil, fmt.Errorf("Error while opening file [%s] for debug output:\n%s\n", s.DebugFile, err)
		}
		logger.SetDebugOutput(debugOutput)
	}

	if s.LogFile != "" {
		logOutput, err = os.OpenFile(s.LogFile, 0, 7770)
		if err != nil {
			return nil, fmt.Errorf("Error while opening file [%s] for log output:\n%s\n", s.LogFile, err)
		}
		logger.SetLogOutput(logOutput)
	} else if debugOutput != nil {
		logger.SetLogOutput(debugOutput)
	}

	if s.ErrorFile != "" {
		errorOutput, err = os.OpenFile(s.ErrorFile, 0, 7770)
		if err != nil {
			return nil, fmt.Errorf("Error while opening file [%s] for error output:\n%s\n", s.ErrorFile, err)
		}
		logger.SetErrorOutput(errorOutput)
	} else if logOutput != nil {
		logger.SetErrorOutput(logOutput)
	}

	return logger, nil
}

type settings struct {
	LogFile   string `json:"log"`
	ErrorFile string `json:"error"`
	DebugFile string `json:"debug"`
}
