package csbmysql_test

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/cloudfoundry/terraform-provider-csbmysql/csbmysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	bindingHost      = "%"
	providerName     = "csbmysql"
	csbMySQLResource = `
provider "{{.ProviderName}}" {
  host            = "{{.DBHost}}"
  port            = {{.Port}}
  username        = "{{.AdminUser}}"
  password        = "{{.AdminPass}}"
  database        = "{{.Database}}"
  sslrootcert     = <<EOF
{{btoa .SSLRootCert}}
EOF
}

resource "{{.ResourceName}}" "binding_user" {
  username = "{{.Username}}"
  password = "{{.Password}}"
}
`
)

var (
	tfStateResourceName = fmt.Sprintf("%s.binding_user", csbmysql.ResourceNameKey)
)

var _ = Describe("Provider", func() {

	DescribeTable("User can be created", func(username, password string, requireSSL bool) {
		provider := initTestProvider()
		resource.Test(GinkgoT(), resource.TestCase{
			IsUnitTest:        true,
			ProviderFactories: getTestProviderFactories(provider),
			CheckDestroy:      checkUserIsDestroyed(username),
			Steps: []resource.TestStep{{
				Config: testGetResourceDefinition(
					resourceDefinitionWithUsername(username),
					resourceDefinitionWithPassword(password),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfStateResourceName, "username", username),
					resource.TestCheckResourceAttr(tfStateResourceName, "password", password),
					checkUserIsCreated(username, password),
				),
			}},
		})
	},
		Entry("without TLS", "some-user", "some-password", false))

})

func checkUserIsCreated(username, password string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		By("CHECKING RESOURCE CREATE")
		By("Confirming that the binding user exists")

		db, err := sql.Open("mysql", adminUserURI)
		Expect(err).NotTo(HaveOccurred())
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)

		getUserStatement, err := db.Prepare("SELECT user, host from mysql.user where User=?")
		Expect(err).NotTo(HaveOccurred())
		defer func(getUserStatement *sql.Stmt) { _ = getUserStatement.Close() }(getUserStatement)
		rows, err := getUserStatement.Query(username)

		Expect(err).NotTo(HaveOccurred())
		Expect(rows.Next()).To(BeTrue())

		var rowUser, rowHost string
		Expect(rows.Scan(&rowUser, &rowHost)).NotTo(HaveOccurred())
		Expect(rowUser).To(Equal(username))
		Expect(rowHost).To(Equal(bindingHost))
		Expect(rows.Next()).To(BeFalse())

		By("Connecting as the binding user")

		userURI := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=skip-verify", username, password, dbHost, port, database)
		dbUser, err := sql.Open("mysql", userURI)
		Expect(err).NotTo(HaveOccurred())

		defer func(dbUser *sql.DB) {
			_ = dbUser.Close()
		}(dbUser)

		By("Creating and populating new tables as the binding user")
		_, err = dbUser.Exec(`CREATE TABLE IF NOT EXISTS tasks (
						task_id INT AUTO_INCREMENT PRIMARY KEY,
						title VARCHAR(255) NOT NULL,
						start_date DATE,
						due_date DATE,
						status TINYINT NOT NULL,
						priority TINYINT NOT NULL,
						description TEXT,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
					)`,
		)
		Expect(err).NotTo(HaveOccurred())

		result, err := dbUser.Exec("insert into tasks(task_id, title, status, priority) values (1, 'task', 2, 3), (2, 'another task', 3, 4)")
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeNumerically("==", 2))

		By("Reading and modifying existing data as the binding user")
		rows, err = dbUser.Query(`select * from previous_table`)
		Expect(err).NotTo(HaveOccurred())

		Expect(rows.Next()).To(BeTrue())
		var (
			id    int
			value string
		)
		err = rows.Scan(&id, &value)
		Expect(err).NotTo(HaveOccurred())
		Expect(id).To(BeNumerically("==", 1))
		Expect(value).To(Equal("value"))
		Expect(rows.Next()).To(BeFalse())

		result, err = dbUser.Exec(`insert into previous_table(pk, value) values (2, 'cheese')`)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeNumerically("==", 1))

		result, err = dbUser.Exec(`update previous_table set value='new_value' where pk=1`)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeNumerically("==", 1))

		result, err = dbUser.Exec(`delete from previous_table where pk in (1, 2)`)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeNumerically("==", 2))

		_, err = dbUser.Exec("drop table previous_table")
		Expect(err).NotTo(HaveOccurred())

		return nil
	}
}

func checkUserIsDestroyed(username string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		var (
			taskId   int
			title    string
			status   int8
			priority int8
		)

		By("CHECKING RESOURCE DELETE")
		By("Confirming that the binding user is deleted")
		db, err := sql.Open("mysql", adminUserURI)
		Expect(err).NotTo(HaveOccurred())
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)

		rows, err := db.Query("SELECT user FROM mysql.user WHERE user = ?", username)
		Expect(err).NotTo(HaveOccurred())
		Expect(rows.Next()).To(BeFalse())

		By("Accessing the removed user's data")
		rows, err = db.Query(fmt.Sprintf("select task_id, title, status, priority from `%s`.tasks order by task_id", database))
		Expect(err).NotTo(HaveOccurred())
		Expect(rows.Next()).To(BeTrue())
		Expect(rows.Scan(&taskId, &title, &status, &priority)).NotTo(HaveOccurred())

		Expect(taskId).To(BeNumerically("==", 1))
		Expect(title).To(Equal("task"))
		Expect(status).To(BeNumerically("==", 2))
		Expect(priority).To(BeNumerically("==", 3))

		Expect(rows.Next()).To(BeTrue())
		Expect(rows.Scan(&taskId, &title, &status, &priority)).NotTo(HaveOccurred())
		Expect(taskId).To(BeNumerically("==", 2))

		return nil
	}
}

func getTestProviderFactories(provider *schema.Provider) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		providerName: func() (*schema.Provider, error) {
			if provider == nil {
				return provider, errors.New("provider cannot be nil")
			}

			return provider, nil
		},
	}
}

func initTestProvider() *schema.Provider {
	testAccProvider := &schema.Provider{
		Schema: csbmysql.ProviderSchema(),
		ResourcesMap: map[string]*schema.Resource{
			csbmysql.ResourceNameKey: csbmysql.ResourceBindingUser(),
		},
		ConfigureContextFunc: csbmysql.ProviderConfigureContext,
	}
	err := testAccProvider.InternalValidate()
	Expect(err).NotTo(HaveOccurred())

	return testAccProvider
}
