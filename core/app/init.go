package app

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

func init() {
	RegisterListener(func(event BeforeInit) (err error) {
		opts := &slog.HandlerOptions{
			AddSource: true,
		}

		c, err := Resolve[*viper.Viper]()
		if err != nil {
			err = fmt.Errorf("get viper when init slog failed: %w", err)
			slog.Error(err.Error())
			return
		}

		if c.GetString("mode") == ModeLocal || c.GetBool("showDebugLog") {
			opts.Level = slog.LevelDebug
		}

		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))
		return
	})
}
