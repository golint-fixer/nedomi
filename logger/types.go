package logger

// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
//
// If you want to edit it go to types.go.template

import (
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"

	"github.com/ironsmile/nedomi/logger/ironsmile"

	"github.com/ironsmile/nedomi/logger/mock"

	"github.com/ironsmile/nedomi/logger/nillogger"

	"github.com/ironsmile/nedomi/logger/std"
)

type newLoggerFunc func(cfg *config.LoggerSection) (types.Logger, error)

var loggerTypes = map[string]newLoggerFunc{

	"ironsmile": func(cfg *config.LoggerSection) (types.Logger, error) {
		return ironsmile.New(cfg)
	},

	"mock": func(cfg *config.LoggerSection) (types.Logger, error) {
		return mock.New(cfg)
	},

	"nillogger": func(cfg *config.LoggerSection) (types.Logger, error) {
		return nillogger.New(cfg)
	},

	"std": func(cfg *config.LoggerSection) (types.Logger, error) {
		return std.New(cfg)
	},
}
