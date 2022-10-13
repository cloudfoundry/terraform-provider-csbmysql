package csbmysql_test

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const (
	adminUser = "root"
	adminPass = "change-me"
	dbHost    = "127.0.0.1"
	port      = 3306
	database  = "nuclear-flux"
)

var (
	adminUserURI = fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql", adminUser, adminPass, dbHost, port)
)

func TestTerraformProviderCSBMySQL(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Terraform Provider CSB MySQL Suite")
}

var _ = BeforeSuite(func() {
	By("Building provider")
	mustRun("go", "build", "..")

	By("Starting MySQL server")
	mysqlVersion, ok := os.LookupEnv("TEST_MYSQL_VERSION_IMAGE_TAG")
	if !ok {
		mysqlVersion = "8"
	}

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
		fmt.Sprintf("mysql:%s", mysqlVersion),
	)
	Eventually(ensureMysqlIsUp).WithTimeout(2 * time.Minute).WithPolling(time.Second).Should(Succeed())

	By("Populating initial data")
	db, err := sql.Open("mysql", adminUserURI)
	Expect(err).NotTo(HaveOccurred())
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	executeSql(db, fmt.Sprintf("create database `%s`", database))
	executeSql(db, fmt.Sprintf("create table `%s`"+`.previous_table (
    pk int primary key not null auto_increment,
    value varchar(255) not null
)`, database))
	executeSql(db, fmt.Sprintf("insert into `%s`.previous_table(pk, value) values (1, 'value')", database))
})

var _ = AfterSuite(func() {
	mustRun("docker", "kill", "mysql")
	mustRun("docker", "rm", "mysql")
})

func mustRun(command ...string) {
	start, err := gexec.Start(exec.Command(
		command[0], command[1:]...,
	), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(start).WithTimeout(time.Minute).WithPolling(time.Second).Should(gexec.Exit(0))
}

func ensureMysqlIsUp(g Gomega) error {
	session, err := gexec.Start(exec.Command("docker", "ps", "-f", "name=mysql"), nil, nil)
	g.Eventually(session).WithTimeout(30 * time.Second).WithPolling(time.Second).Should(gexec.Exit(0))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(session).To(gbytes.Say("healthy"))
	return nil
}

func executeSql(db *sql.DB, statement string) {
	_, err := db.Exec(statement)
	Expect(err).NotTo(HaveOccurred())
}
