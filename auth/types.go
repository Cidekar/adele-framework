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
	ID        int       `db:"id,omitempty"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Active    int       `db:"user_active"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type RememberToken struct {
	ID            int       `db:"id,omitempty"`
	UserID        int       `db:"user_id"`
	RememberToken string    `db:"remember_token"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}
