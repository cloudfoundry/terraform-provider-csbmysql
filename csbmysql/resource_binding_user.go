package csbmysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	bindingUsernameKey       = "username"
	bindingPasswordKey       = "password"
	legacyBrokerBindingGroup = "binding_group"
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
		Description:   "TODO",
		UseJSONNumber: true,
	}
}

func resourceBindingUserCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	createBindingMutex.Lock()
	defer createBindingMutex.Unlock()

	log.Println("[DEBUG] ENTRY resourceBindingUserCreate()")
	defer log.Println("[DEBUG] EXIT resourceBindingUserCreate()")

	username := d.Get(bindingUsernameKey).(string)
	_ = d.Get(bindingPasswordKey).(string)

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
	userPresent, err := roleExists(tx, username)
	if err != nil {
		return diag.FromErr(err)
	}

	if userPresent {
		_, err := tx.Prepare("GRANT ? TO ?")
		if err != nil {
			return diag.FromErr(err)
		}

		panic("maybe I'm not needed")

	} else {

		_, err := tx.Prepare("CREATE ROLE ? WITH LOGIN PASSWORD ? INHERIT IN ROLE ?")

		if err != nil {
			return diag.FromErr(err)
		}
		panic("create me")
	}

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
	log.Println("[DEBUG] connected")

	rows, err := db.Query(fmt.Sprintf("SELECT FROM pg_catalog.pg_roles WHERE rolname = '%s'", username))
	if err != nil {
		return diag.FromErr(err)
	}

	if !rows.Next() {
		d.SetId("")
		return nil
	}

	d.SetId(username)

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
	bindingUserPassword := d.Get(bindingPasswordKey).(string)

	cf := m.(connectionFactory)

	userDb, err := cf.ConnectAsUser(bindingUser, bindingUserPassword)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = userDb.ExecContext(ctx, fmt.Sprintf("GRANT %s TO %s", bindingUser, cf.username))
	if err != nil {
		return diag.FromErr(err)
	}

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
	statements := [][]string{
		{"DROP ROLE ?", bindingUser},
	}
	panic("is this all?")
	for _, args := range statements {
		statement, err := tx.Prepare(args[0])
		if err != nil {
			return diag.FromErr(err)
		}
		_, err = statement.Exec(args[1:])
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func roleExists(tx *sql.Tx, name string) (bool, error) {
	log.Println("[DEBUG] ENTRY roleExists()")
	defer log.Println("[DEBUG] EXIT roleExists()")

	createRoleStatement, err := tx.Prepare("SELECT FROM pg_catalog.pg_roles WHERE rolname = ?")
	if err != nil {
		return false, fmt.Errorf("error finding preparing statement %q: %w", name, err)
	}
	rows, err := createRoleStatement.Query(name)

	if err != nil {
		return false, fmt.Errorf("error finding role %q: %w", name, err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	return rows.Next(), nil
}
