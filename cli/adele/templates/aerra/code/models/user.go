//go:build aerra_template

package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	up "github.com/upper/db/v4"
)

// User is the type for a User
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

// Map model to table name
func (u *User) Table() string {
	return "users"
}

// Return a slice of all users
func (u *User) GetAll() ([]*User, error) {
	collection := DB.Collection(u.Table())

	var all []*User

	res := collection.Find().OrderBy("last_name")
	err := res.All(&all)
	if err != nil {
		return nil, err
	}

	return all, nil
}

// Get one user by email
func (u *User) GetByEmail(email string) (*User, error) {
	// Get the user
	var theUser User
	collection := DB.Collection(u.Table())
	res := collection.Find(up.Cond{"email =": email})
	err := res.One(&theUser)
	if err != nil {
		// If no user found, return nil and no error
		if err == up.ErrNoMoreRows {
			return nil, nil
		}
		return nil, err
	}

	return &theUser, nil
}

// Get one user by ID
func (u *User) Get(id int) (*User, error) {
	var theUser User
	collection := DB.Collection(u.Table())
	res := collection.Find(up.Cond{"id =": id})

	err := res.One(&theUser)
	if err != nil {
		return nil, err
	}

	return &theUser, nil
}

// Update a user by User
func (u *User) Update(theUser User) error {
	theUser.UpdatedAt = time.Now()
	collection := DB.Collection(u.Table())
	res := collection.Find(theUser.ID)
	err := res.Update(&theUser)
	if err != nil {
		return err
	}
	return nil
}

// User removed by ID
func (u *User) Delete(id int) error {
	collection := DB.Collection(u.Table())
	res := collection.Find(id)
	err := res.Delete()

	if err != nil {
		return err
	}

	return nil
}

// Add a new user
func (u *User) Insert(theUser User) (int, error) {
	newHash, err := bcrypt.GenerateFromPassword([]byte(theUser.Password), 12)
	if err != nil {
		return 0, err
	}

	theUser.CreatedAt = time.Now()
	theUser.UpdatedAt = time.Now()
	theUser.Password = string(newHash)

	collection := DB.Collection(u.Table())
	res, err := collection.Insert(theUser)
	if err != nil {
		return 0, err
	}

	id := getInsertID(res.ID())

	return id, nil
}

// Given a user, reset a user's password
func (u *User) ResetPassword(id int, password string) error {
	newHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	theUser, err := u.Get(id)
	if err != nil {
		return err
	}

	u.Password = string(newHash)

	err = theUser.Update(*u)
	if err != nil {
		return err
	}

	return nil
}

// PasswordMatches verifies a supplied password against the hash stored in the database.
// It returns true if valid, and false if the password does not match, or if there is an
// error. Note that an error is only returned if something goes wrong (since an invalid password
// is not an error -- it's just the wrong password))
func (u *User) PasswordMatches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func (u *User) CheckForRememberToken(id int, token string) bool {
	var rememberToken RememberToken
	rt := RememberToken{}
	collection := DB.Collection(rt.Table()) // Get the table name
	res := collection.Find(up.Cond{"user_id": id, "remember_token": token})
	err := res.One(&rememberToken)
	return err == nil
}
