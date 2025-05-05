package env

import (
	"log/slog"
	"os"

	"github.com/stretchr/testify/assert/yaml"
)

func Load(name string, hide ...string) {
	for _, fname := range []string{
		"env",
		".env",
		"env.yaml",
		".env.yaml",
	} {
		bs, err := os.ReadFile(fname)
		if err != nil {
			continue
		}

		envs := map[string]map[string]string{}

		err = yaml.Unmarshal(bs, &envs)
		if err != nil {
			continue
		}

		env, ok := envs[name]
		if !ok {
			continue
		}

		for k, v := range env {
			os.Setenv(k, v)

			printed := false

			for _, h := range hide {
				if k == h {
					slog.Info("Env set", "name", k)

					printed = true

					break
				}
			}

			if !printed {
				slog.Info("Env set", "name", k, "value", v)
			}
		}

		break
	}
}
