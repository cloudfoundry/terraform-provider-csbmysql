package csbmysql_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"text/template"
	"time"

	"github.com/cloudfoundry/terraform-provider-csbmysql/csbmysql"
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
	adminUserURI = fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql?tls=skip-verify", adminUser, adminPass, dbHost, port)
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

	createFixtureVolume()

	mustRun(
		"docker",
		"run",
		"--name=mysql",
		"-d",
		"--publish=3306:3306",
		"-e",
		fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", adminPass),
		"--mount",
		"source=mysql_config,destination=/etc/mysql/conf.d",
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
	mustRun("docker", "rm", "-f", "mysql")
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

func parse(m interface{}, resourceTmpl string) (string, error) {
	var definitionBytes bytes.Buffer

	t := template.Must(template.New("resource").Parse(resourceTmpl))
	if err := t.Execute(&definitionBytes, m); err != nil {
		return "", err
	}

	return definitionBytes.String(), nil
}

type definition struct {
	ProviderName,
	ResourceName,
	DBHost,
	AdminUser,
	AdminPass,
	Database,
	Username,
	Password string
	Port        int
	SSLRootCert []byte
}

type setDefinitionFunc func(*definition)

func testGetResourceDefinition(optFns ...setDefinitionFunc) string {
	caCertPath := path.Join(getCurrentDirectory(), "testfixtures", "ssl_mysql", "certs", "ca.crt")
	rootCertificate, err := os.ReadFile(caCertPath)
	Expect(err).NotTo(HaveOccurred())
	c := definition{
		ProviderName: providerName,
		ResourceName: csbmysql.ResourceNameKey,
		DBHost:       dbHost,
		AdminUser:    adminUser,
		AdminPass:    adminPass,
		Database:     database,
		Port:         port,
		SSLRootCert:  rootCertificate,
	}

	for _, fn := range optFns {
		fn(&c)
	}

	hcl, err := parse(&c, csbMySQLResource)
	Expect(err).NotTo(HaveOccurred())
	return hcl
}

func resourceDefinitionWithUsername(username string) setDefinitionFunc {
	return func(config *definition) {
		config.Username = username
	}
}

func resourceDefinitionWithPassword(password string) setDefinitionFunc {
	return func(config *definition) {
		config.Password = password
	}
}

func createFixtureVolume() {
	mustRun("docker", "volume", "create", "mysql_config")
	for _, folder := range []string{"certs", "keys"} {
		dockerVolumeRun("rm", "-rf", fmt.Sprintf("/mnt/%s", folder))
		dockerVolumeRun("cp", "-r", fmt.Sprintf("/fixture/ssl_mysql/%s", folder), "/mnt")
	}
	dockerVolumeRun("rm", "/mnt/my.cnf")
	dockerVolumeRun("cp", "/fixture/my.cnf", "/mnt")
	dockerVolumeRun("chown", "mysql", "/mnt/keys/server.key")
	dockerVolumeRun("chmod", "0600", "/mnt/keys/server.key")
}

func dockerVolumeRun(cmd ...string) {
	fmt.Fprintln(GinkgoWriter, "Running docker command", cmd)
	fixturePath := path.Join(getCurrentDirectory(), "testfixtures")
	volumeMount := fmt.Sprintf("%s:/fixture", fixturePath)
	dockerVolumeCommand := []string{"docker", "run", "-v", volumeMount, "--mount", "source=mysql_config,destination=/mnt", "mysql"}
	dockerVolumeCommand = append(dockerVolumeCommand, cmd...)
	mustRun(dockerVolumeCommand...)
}

func getCurrentDirectory() string {
	_, file, _, _ := runtime.Caller(1)
	return filepath.Dir(file)
}
