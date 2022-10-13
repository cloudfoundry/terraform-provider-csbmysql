package csbmysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	bindingUsernameKey = "username"
	bindingPasswordKey = "password"
	bindingUserHostAll = "%"
)

var (
	createBindingMutex sync.Mutex
	deleteBindingMutex sync.Mutex
)

func resourceBindingUser() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			bindingUsernameKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			bindingPasswordKey: {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
		CreateContext: resourceBindingUserCreate,
		ReadContext:   resourceBindingUserRead,
		UpdateContext: resourceBindingUserUpdate,
		DeleteContext: resourceBindingUserDelete,
		Description:   "A MySQL Server binding for the CSB brokerpak",
		UseJSONNumber: true,
	}
}

func resourceBindingUserCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	createBindingMutex.Lock()
	defer createBindingMutex.Unlock()

	log.Println("[DEBUG] ENTRY resourceBindingUserCreate()")
	defer log.Println("[DEBUG] EXIT resourceBindingUserCreate()")

	username := d.Get(bindingUsernameKey).(string)
	password := d.Get(bindingPasswordKey).(string)

	cf := m.(connectionFactory)

	db, err := cf.ConnectAsAdmin()

	if err != nil {
		return diag.FromErr(err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	log.Println("[DEBUG] connected")

	log.Println("[DEBUG] create binding user")
	userPresent, err := userExists(db, username, "%")
	if err != nil {
		return diag.FromErr(err)
	}

	if !userPresent {
		tlsRequired := "NONE"
		if cf.verifyTLS {
			tlsRequired = "SSL"
		}
		_, err := tx.Exec(fmt.Sprintf("CREATE USER %s@%s IDENTIFIED BY %s REQUIRE %s", quotedIdentifier(username),
			quotedIdentifier(bindingUserHostAll), quotedString(password), tlsRequired))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	grantStatement := fmt.Sprintf("GRANT ALL ON %s.* TO %s@%s", quotedIdentifier(cf.database),
		quotedIdentifier(username), quotedIdentifier(bindingUserHostAll))
	_, err = tx.Exec(grantStatement)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tx.Commit()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] setting ID %s\n", username)
	d.SetId(username)

	return nil
}

func resourceBindingUserRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	log.Println("[DEBUG] ENTRY resourceBindingUserRead()")
	defer log.Println("[DEBUG] EXIT resourceBindingUserRead()")

	username := d.Get(bindingUsernameKey).(string)

	cf := m.(connectionFactory)

	db, err := cf.ConnectAsAdmin()
	if err != nil {
		return diag.FromErr(err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	userPresent, err := userExists(db, username, bindingUserHostAll)
	if err != nil {
		return diag.FromErr(err)
	}
	if userPresent {
		d.SetId(username)
	}

	return nil
}

func resourceBindingUserUpdate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.FromErr(fmt.Errorf("update lifecycle not implemented"))
}

func resourceBindingUserDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	log.Println("[DEBUG] ENTRY resourceBindingUserDelete()")
	defer log.Println("[DEBUG] EXIT resourceBindingUserDelete()")

	deleteBindingMutex.Lock()
	defer deleteBindingMutex.Unlock()

	bindingUser := d.Get(bindingUsernameKey).(string)

	cf := m.(connectionFactory)

	db, err := cf.ConnectAsAdmin()
	if err != nil {
		return diag.FromErr(err)
	}

	defer func(connection *sql.DB) {
		_ = connection.Close()
	}(db)

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	defer func(transaction *sql.Tx) {
		_ = transaction.Rollback()
	}(tx)

	log.Println("[DEBUG] dropping binding user")
	_, err = tx.Exec(fmt.Sprintf("DROP USER '%s'@'%s'", bindingUser, bindingUserHostAll))
	if err != nil {
		return diag.FromErr(err)
	}

	err = tx.Commit()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func userExists(db *sql.DB, name, host string) (bool, error) {
	log.Println("[DEBUG] ENTRY roleExists()")
	defer log.Println("[DEBUG] EXIT roleExists()")

	checkUserStatement, err := db.Prepare("SELECT 1 FROM mysql.user WHERE user = ? and HOST = ?")
	if err != nil {
		return false, fmt.Errorf("error preparing statement %q: %w", name, err)
	}
	rows, err := checkUserStatement.Query(name, host)

	if err != nil {
		return false, fmt.Errorf("error finding user %q: %w", name, err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	return rows.Next(), nil
}
