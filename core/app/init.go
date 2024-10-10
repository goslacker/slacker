package app

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

func init() {
	RegisterListener(func(event BeforeInit) {
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
