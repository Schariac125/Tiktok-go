package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// .env 的解析规则：注释与空行跳过，值两侧成对的引号去掉
func TestParseLine(t *testing.T) {
	cases := []struct {
		line      string
		wantKey   string
		wantValue string
		wantOK    bool
	}{
		{"KEY=value", "KEY", "value", true},
		{"  KEY  =  value  ", "KEY", "value", true},
		{"KEY=", "KEY", "", true},
		{`KEY="has space"`, "KEY", "has space", true},
		{"KEY='single'", "KEY", "single", true},
		{"export KEY=value", "KEY", "value", true},

		{"", "", "", false},
		{"   ", "", "", false},
		{"# 注释", "", "", false},
		{"#KEY=value", "", "", false},
		{"没有等号", "", "", false},
		{"=value", "", "", false},
	}

	for _, c := range cases {
		key, value, ok := parseLine(c.line)
		if ok != c.wantOK {
			t.Errorf("parseLine(%q) ok = %v, 期望 %v", c.line, ok, c.wantOK)
			continue
		}
		if !ok {
			continue
		}
		if key != c.wantKey || value != c.wantValue {
			t.Errorf("parseLine(%q) = (%q, %q), 期望 (%q, %q)",
				c.line, key, value, c.wantKey, c.wantValue)
		}
	}
}

// 已存在的环境变量优先于 .env 里的值：
// 这样 CI 或 shell 里临时 export 的配置能盖过文件里的默认值。
func TestLoadDotEnvDoesNotOverrideExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := "TEST_ENV_EXISTING=from-file\nTEST_ENV_ONLY_FILE=from-file\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TEST_ENV_EXISTING", "from-shell")
	os.Unsetenv("TEST_ENV_ONLY_FILE") // t.Setenv 只能设值，清不掉，这里手动来

	if err := LoadDotEnv(path); err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("TEST_ENV_EXISTING"); got != "from-shell" {
		t.Errorf("已存在的变量被 .env 覆盖了: got %q, 期望 from-shell", got)
	}
	if got := os.Getenv("TEST_ENV_ONLY_FILE"); got != "from-file" {
		t.Errorf(".env 里的新变量没生效: got %q, 期望 from-file", got)
	}
}

// 文件不存在不是错误：容器里环境变量由 compose 注入，没有 .env 是正常的
func TestLoadDotEnvMissingFile(t *testing.T) {
	if err := LoadDotEnv(filepath.Join(t.TempDir(), "不存在")); err != nil {
		t.Fatalf("文件不存在时应返回 nil, 实际 %v", err)
	}
}

// 缺变量时必须 panic 并报出变量名。
//
// 这是本包存在的理由：os.ExpandEnv 会把缺失的 ${VAR} 换成空串，
// 密码为空的服务照常启动，直到第一个请求查库才报 Access denied。
func TestMustLoadPanicsOnMissingEnv(t *testing.T) {
	const absent = "TEST_ENV_DEFINITELY_ABSENT"
	original, had := os.LookupEnv(absent)
	os.Unsetenv(absent)
	defer func() {
		if had {
			os.Setenv(absent, original)
		}
	}()

	path := filepath.Join(t.TempDir(), "test.yaml")
	content := "Name: test\nMySQL:\n  DataSource: root:${" + absent + "}@tcp(127.0.0.1:3306)/x\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("缺环境变量时应 panic")
		}
		if msg := fmt.Sprint(r); !strings.Contains(msg, absent) {
			t.Fatalf("panic 信息里应含变量名 %s, 实际: %s", absent, msg)
		}
	}()

	var c struct{ Name string }
	MustLoad(path, &c)
}

// 正常路径：${VAR} 被替换成环境变量的值，且替换只发生一次
// （密码里带 $ 不会被二次展开）
func TestMustLoadInterpolates(t *testing.T) {
	t.Setenv("TEST_ENV_PWD", "p$ss-word")

	path := filepath.Join(t.TempDir(), "test.yaml")
	content := "Name: test\nMySQL:\n  DataSource: root:${TEST_ENV_PWD}@tcp(127.0.0.1:3306)/x\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	var c struct {
		Name  string
		MySQL struct{ DataSource string }
	}
	MustLoad(path, &c)

	const want = "root:p$ss-word@tcp(127.0.0.1:3306)/x"
	if c.MySQL.DataSource != want {
		t.Errorf("DataSource = %q, 期望 %q", c.MySQL.DataSource, want)
	}
}
