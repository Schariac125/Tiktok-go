// Package env 把配置里的私密信息从 yaml 挪到环境变量。
//
// go-zero 的 conf 包本身支持 ${VAR} 插值，但直接用有三个不便：
//
//  1. 必须显式传 conf.UseEnv()，否则 ${VAR} 会被当成普通字符串原样留在配置里；
//  2. 底层是 os.ExpandEnv，不认 ${VAR:-默认值} 这种写法，变量没设置时
//     会静默展开成空串——MySQL 密码变空不会在启动时报错，而是等到第一个
//     请求查库才返回 Access denied，问题现场离原因很远；
//  3. 不读 .env，本地开发每次都得先手动 export。
//
// 本包把这三件事一并处理：读 .env、校验占位符、开启插值后再解析 yaml。
package env

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/zeromicro/go-zero/core/conf"
)

// DotEnvFile 是本地开发默认读取的环境变量文件，位于项目根目录。
const DotEnvFile = ".env"

// envRef 匹配 yaml 里的 ${VAR} 占位符。
//
// 只认带花括号的写法：$VAR 与 ${VAR} 对 os.ExpandEnv 是等价的，
// 但前者容易和文件里其他的 $ 混淆，本项目统一用 ${VAR}。
var envRef = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// MustLoad 加载 yaml 配置，支持 ${VAR} 占位符。
//
// 先读 .env（存在的话）注入进程环境，再检查配置里出现的每个 ${VAR}
// 都有对应环境变量，最后交给 conf.MustLoad 解析。
// 任一环节不满足都会 panic——配置错误属于启动期就该暴露的问题，
// 让服务带着空密码跑起来只会把故障推迟到第一个请求上。
func MustLoad(path string, v any) {
	if err := LoadDotEnv(); err != nil {
		panic(fmt.Errorf("读取 %s 失败: %w", DotEnvFile, err))
	}
	if err := checkEnvRefs(path); err != nil {
		panic(err)
	}

	conf.MustLoad(path, v, conf.UseEnv())
}

// LoadDotEnv 把 .env 里的键值写进进程环境。
//
// 文件不存在时返回 nil：容器里环境变量由 compose 注入，没有 .env 是正常的。
// 已经存在的环境变量优先，不会被文件覆盖——这样 CI 或 shell 里临时
// export 的值能盖过文件里的值。
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

// parseLine 解析一行 KEY=VALUE。
// 空行、# 开头的注释、不含 = 的行都返回 ok=false 直接跳过；
// 值两侧成对的引号会被去掉，方便写带空格的值。
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

// checkEnvRefs 检查配置里的 ${VAR} 是否都有对应环境变量，缺失时汇总报错。
//
// 不检查的话 os.ExpandEnv 会把缺失的占位符换成空串，
// 于是 root:${MYSQL_ROOT_PASSWORD}@tcp(...) 变成 root:@tcp(...)，
// 服务照常启动，直到查库那一刻才失败。
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

		// 只要求"设置过"，值为空串也算数——OSS 凭据这类没配也能启动，
		// 在 .env 里写成 OSS_ACCESS_KEY_ID= 即可。
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
