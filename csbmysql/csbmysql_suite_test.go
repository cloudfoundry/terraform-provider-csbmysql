package csbmysql_test

import (
	"fmt"
	"github.com/onsi/gomega/gbytes"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	name        = "binding-user"
	adminUser   = "root"
	adminPass   = "change-me"
	dbHost      = "127.0.0.1"
	bindingHost = "%"
	port        = 3306
	database    = "mysql"
)

var (
	adminUserURI = fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql", adminUser, adminPass, dbHost, port)
)

func TestTerraformProviderCSBMySQL(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Terraform Provider CSB MySQL Suite")
}

var _ = BeforeSuite(func() {
	// Build Provider
	mustRun("go",
		"build",
		"..",
	)

	// StartMysql
	mustRun(
		"docker",
		"run",
		"--name=mysql",
		"-d",
		"--publish=3306:3306",
		"-e",
		fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", adminPass),
		"--health-cmd",
		fmt.Sprintf("mysqladmin -h %s -P %d -u %s -p%s ping", dbHost, port, adminUser, adminPass),
		"mysql:8",
	)
	Eventually(ensureMysqlIsUp).WithTimeout(2 * time.Minute).WithPolling(time.Second).Should(Succeed())
})

var _ = AfterSuite(func() {
	mustRun("docker",
		"kill",
		"mysql",
	)
	mustRun("docker",
		"rm",
		"mysql",
	)
})

func mustRun(command ...string) {
	start, err := gexec.Start(exec.Command(
		command[0], command[1:]...,
	), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(start).WithTimeout(30 * time.Second).WithPolling(time.Second).Should(gexec.Exit(0))
}

func ensureMysqlIsUp(g Gomega) error {
	session, err := gexec.Start(exec.Command("docker", "ps", "-f", "name=mysql"), nil, nil)
	g.Eventually(session).WithTimeout(30 * time.Second).WithPolling(time.Second).Should(gexec.Exit(0))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(session).To(gbytes.Say("healthy"))
	return nil
}
