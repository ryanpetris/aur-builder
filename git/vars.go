package git

import (
	"os"
	"strings"
)

var insecureSkipTls = false

func init() {
	if val, ok := os.LookupEnv("GIT_SSL_NO_VERIFY"); ok {
		insecureSkipTls = strings.ToLower(val) == "true" || val == "1"
	}
}
