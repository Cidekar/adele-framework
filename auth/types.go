package auth

import (
	"log"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/cidekar/adele-framework/database"
)

type Auth struct {
	AppName  string
	DB       *database.Database
	ErrorLog *log.Logger
	Session  *scs.SessionManager
}

type User struct {
	ID        int       `db:"id,omitempty" json:"id"`
	FirstName string    `db:"first_name"   json:"firstName"`
	LastName  string    `db:"last_name"    json:"lastName"`
	Email     string    `db:"email"        json:"email"`
	Active    int       `db:"user_active"  json:"active"`
	Password  string    `db:"password"     json:"-"`
	CreatedAt time.Time `db:"created_at"   json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at"   json:"updatedAt"`
}

type RememberToken struct {
	ID            int       `db:"id,omitempty"   json:"id"`
	UserID        int       `db:"user_id"        json:"userId"`
	RememberToken string    `db:"remember_token" json:"-"`
	CreatedAt     time.Time `db:"created_at"     json:"createdAt"`
	UpdatedAt     time.Time `db:"updated_at"     json:"updatedAt"`
}
