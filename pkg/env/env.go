package env

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/zeromicro/go-zero/core/conf"
)

const DotEnvFile = ".env"

var envRef = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

func MustLoad(path string, v any) {
	if err := LoadDotEnv(); err != nil {
		panic(fmt.Errorf("读取 %s 失败: %w", DotEnvFile, err))
	}
	if err := checkEnvRefs(path); err != nil {
		panic(err)
	}

	conf.MustLoad(path, v, conf.UseEnv())
}

func LoadDotEnv(paths ...string) error {
	if len(paths) == 0 {
		paths = []string{DotEnvFile}
	}

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		for line := range strings.SplitSeq(string(content), "\n") {
			key, value, ok := parseLine(line)
			if !ok {
				continue
			}
			if _, exists := os.LookupEnv(key); exists {
				continue
			}
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func parseLine(line string) (key, value string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}

	key, value, found := strings.Cut(line, "=")
	if !found {
		return "", "", false
	}

	key = strings.TrimSpace(strings.TrimPrefix(key, "export "))
	if key == "" {
		return "", "", false
	}

	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
	}

	return key, value, true
}

func checkEnvRefs(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	seen := make(map[string]struct{})
	var missing []string

	for _, match := range envRef.FindAllStringSubmatch(string(content), -1) {
		name := match[1]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}

		if _, ok := os.LookupEnv(name); !ok {
			missing = append(missing, name)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	sort.Strings(missing)
	return fmt.Errorf("%s 引用了未设置的环境变量: %s\n"+
		"本地开发请复制 .env.example 为 .env 并按需填写；容器内由 docker compose 注入",
		path, strings.Join(missing, ", "))
}
