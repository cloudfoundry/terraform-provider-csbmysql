//go:build tools
// +build tools

package tools

import (
	_ "github.com/onsi/ginkgo/v2/ginkgo/generators"
	_ "github.com/onsi/ginkgo/v2/ginkgo/internal"
	_ "github.com/onsi/ginkgo/v2/ginkgo/labels"
)

// This file imports packages that are used when running go generate, or used
// during the development process but not otherwise depended on by built code.
