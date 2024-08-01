package app

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/goslacker/slacker/container"
	"github.com/spf13/viper"
)

const (
	ModeLocal   = "local"
	ModeDevelop = "develop"
	ModeStaging = "staging"
	ModeRelease = "release"
)

type configOpt struct {
	config string
	path   string
}

func WithPath(path string) func(*configOpt) {
	return func(opt *configOpt) {
		opt.path = path
	}
}

func WithContent(content string) func(*configOpt) {
	return func(opt *configOpt) {
		opt.config = content
	}
}

func LoadConfig(opts ...func(*configOpt)) (err error) {
	m := &configOpt{}
	for _, opt := range opts {
		opt(m)
	}

	viper.SetConfigType("yaml")

	if m.path != "" {
		m.config = readConfig(m.path)
	}

	err = viper.ReadConfig(strings.NewReader(m.config))
	if err != nil {
		return
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viper.AutomaticEnv()
	err = container.Bind[*viper.Viper](viper.GetViper)
	if err != nil {
		return
	}

	if viper.GetString("mode") == "" {
		viper.Set("mode", ModeLocal)
	}
	setLog(viper.GetViper())

	return
}

func setLog(c *viper.Viper) {
	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	if c.GetString("mode") == ModeLocal || c.GetBool("showDebugLog") {
		opts.Level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))
}

func readConfig(path string) (result string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return
	}

	return string(content)
}
