package csbmysql_test

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/terraform-provider-csbmysql/csbmysql"
)

var _ = Describe("Provider", func() {
	It("Can be used for initializing a tf project", func() {

		applyHCL(fmt.Sprintf(`
		provider "csbmysql" {
		  host            = "%s"
		  port            = %d
		  username        = "%s"
		  password        = "%s"
		  database        = "%s"

		}
		resource "csbmysql_binding_user" "binding_user" {
		  username = "%s"
		  password = "%s"
		}
		`, dbHost, port, adminUser, adminPass, database, name, "binding-password"),

			func(state *terraform.State) error {
				By("CHECKING RESOURCE CREATE")

				db, err := sql.Open("mysql", adminUserURI)
				defer func(db *sql.DB) {
					_ = db.Close()
				}(db)

				Expect(err).NotTo(HaveOccurred())
				getUserStatement, err := db.Prepare("SELECT user, host from mysql.user where User=?")
				Expect(err).NotTo(HaveOccurred())
				rows, err := getUserStatement.Query(name)
				Expect(err).NotTo(HaveOccurred())
				Expect(rows.Next()).To(BeTrue())
				var rowUser, rowHost string
				Expect(rows.Scan(&rowUser, &rowHost)).NotTo(HaveOccurred())
				Expect(rowUser).To(Equal(name))
				Expect(rowHost).To(Equal(bindingHost))
				Expect(rows.Next()).To(BeFalse())
				return nil
			},
			func(state *terraform.State) error {
				By("CHECKING RESOURCE DELETE")
				db, err := sql.Open("mysql", adminUserURI)
				Expect(err).NotTo(HaveOccurred())

				By("checking that the binding user is deleted")
				checkUserStatement, err := db.Prepare("SELECT user FROM mysql.user WHERE user = ?")
				Expect(err).NotTo(HaveOccurred())
				rows, err := checkUserStatement.Query(name)
				Expect(err).NotTo(HaveOccurred())
				Expect(rows.Next()).To(BeFalse())

				return nil
			})
	})

})

func applyHCL(hcl string, checkOnCreate, checkOnDestroy resource.TestCheckFunc) {

	resource.Test(GinkgoT(), resource.TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"csbmysql": func() (*schema.Provider, error) { return csbmysql.Provider(), nil },
		},
		CheckDestroy: checkOnDestroy,
		Steps: []resource.TestStep{{
			ResourceName: "csbmysql_binding_user.example",
			Config:       hcl,
			Check:        checkOnCreate,
		}},
	})
}
