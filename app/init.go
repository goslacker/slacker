package app

import (
	"github.com/spf13/viper"
	"log/slog"
	"os"
)

func init() {
	RegisterListener(func(event *BeforeInit) {
		opts := &slog.HandlerOptions{
			AddSource: true,
		}

		c, err := Resolve[*viper.Viper]()
		if err != nil {
			slog.Error("init slog failed", "err", err)
			return
		}

		if c.GetString("mode") == ModeLocal || c.GetBool("showDebugLog") {
			opts.Level = slog.LevelDebug
		}

		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))
	})
}
