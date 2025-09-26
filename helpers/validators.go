package helpers

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/asaskevich/govalidator"
	"github.com/fatih/camelcase"
)

type Validation struct {
	Data   url.Values
	Errors map[string]string
}

// Create new validator for use with a form.
//
// Example:
// validator := helpers.NewValidator(r.Form)
func (h *Helpers) NewValidator(data url.Values) *Validation {
	return &Validation{
		Errors: make(map[string]string),
		Data:   data,
	}
}

// Test if any errors exist.
//
// Example:
//
//	if !validator.Valid() {
//		... Handle validation fail
//	}
func (v *Validation) Valid() bool {
	return len(v.Errors) == 0
}

// ToString converts the Validation.Errors map to a human-readable string format.
// Useful for displaying all validation errors as a single message.
//
// Example:
//
//		validator.Required(r, "name", "email")
//		if !validator.Valid() {
//		    errorString := validator.ToString()
//	 Returns: " Name is required Email is required"
//		}
func (v *Validation) ToString() string {
	b := new(bytes.Buffer)
	for _, value := range v.Errors {
		fmt.Fprintf(b, " %s", value)
	}
	return b.String()
}

// AddError adds an error message to the validation errors map if the key doesn't already exist.
// The message supports :attribute placeholder which gets replaced with a formatted field name.
//
// Example:
//
//		validator.AddError("email", "The :attribute field must be valid")
//	 Results in: "The email field must be valid"
func (v *Validation) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		fieldName := formatFieldName(key)
		v.Errors[key] = strings.ReplaceAll(message, ":attribute", fieldName)
	}
}

// Has checks if a field exists in the HTTP request form data and has a non-empty value.
// Returns true if the field exists and has content, false otherwise.
//
// Example:
//
//		if validator.Has("username", r) {
//	 Field exists and has a value
//		}
func (v *Validation) Has(field string, r *http.Request) bool {
	isInRequest := r.Form.Get(field)
	return isInRequest != ""
}

// Required validates that specified fields exist and are not empty in the HTTP request form.
// Adds error messages for any missing or empty required fields.
//
// Example:
//
// validator.Required(r, "name", "email", "password")
// Checks that all three fields have values
func (v *Validation) Required(r *http.Request, fields ...string) {
	for _, field := range fields {
		value := strings.TrimSpace(r.Form.Get(field))
		if value == "" {
			v.AddError(field, fmt.Sprintf("The %s field is required.", ":attribute"))
		}
	}
}

// formatFieldName converts camelCase or PascalCase field names to lowercase with spaces.
// Used internally to create human-readable field names for error messages.
func formatFieldName(field string) string {
	s := camelcase.Split(field)
	var n string
	if len(s) > 0 {
		for i := range s {
			n = n + " " + (s[i])
		}
	}
	return strings.TrimSpace(strings.ToLower(n))
}

// RequiredJSON validates that specified fields exist in a JSON struct using reflection.
// Checks if the provided interface contains the required fields as struct properties.
//
// Example:
//
//	type User struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//	user := &User{}
//	validator.RequiredJSON(user, "Name", "Email")
func (v *Validation) RequiredJSON(json interface{}, fields ...string) {
	reflectedType := reflect.TypeOf(json)
	reflectedKind := reflectedType.Kind()

	// If we do not have reflection pointer, add all required fields as errors
	if reflectedKind != reflect.Ptr {
		for _, field := range fields {
			v.AddError(field, "The :attribute field is required")
		}
		return
	}

	// Using reflection, search for the required fields
	vp := reflect.ValueOf(json)
	vs := reflect.Indirect(vp)
	for _, field := range fields {
		var ok bool
		for i := 0; i < vs.NumField(); i++ {
			name := vs.Type().Field(i).Name
			if strings.EqualFold(field, name) {
				ok = true
				break
			}
		}
		if !ok {
			v.AddError(field, "The :attribute field is required")
		}
	}
}

// HasJSON validates that specified fields exist in a JSON struct and contain non-empty values.
// Uses reflection to check both field existence and value content.
//
// Example:
//
//	type User struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//	user := &User{Name: "John", Email: ""}
//	validator.HasJSON(user, "Name", "Email")
//
// Name passes, Email fails validation
func (v *Validation) HasJSON(json interface{}, fields ...string) {
	reflectedType := reflect.TypeOf(json)
	reflectedKind := reflectedType.Kind()

	// If we do not have reflection pointer, add all required fields as errors
	if reflectedKind != reflect.Ptr {
		for _, field := range fields {
			v.AddError(field, "The :attribute field is required")
		}
		return
	}

	// Using reflection, search for the required fields and check their values
	vp := reflect.ValueOf(json)
	vs := reflect.Indirect(vp)
	for _, field := range fields {
		var ok bool
		for i := 0; i < vs.NumField(); i++ {
			name := vs.Type().Field(i).Name
			if strings.EqualFold(field, name) {
				value := vs.Field(i).Interface()
				if value != "" && value != nil {
					ok = true
				}
				break
			}
		}
		if !ok {
			v.AddError(field, "The :attribute field is required")
		}
	}
}

// Check adds an error message if the given condition is false.
// Useful for custom validation logic with conditional error reporting.
//
// Example:
//
//	age, _ := strconv.Atoi(r.Form.Get("age"))
//	validator.Check(age >= 18, "age", "Must be 18 or older")
func (v *Validation) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// IsEmail validates that a field contains a properly formatted email address.
// Uses the govalidator library for RFC-compliant email validation.
//
// Example:
//
//	email := r.Form.Get("email")
//	validator.IsEmail("email", email)
//
// Validates format like "user@example.com"
func (v *Validation) IsEmail(field, value string) {
	if !govalidator.IsEmail(value) {
		v.AddError(field, "Invalid email address")
	}
}

// IsEmailInPublicDomain validates that an email address exists and is in a public domain.
// Uses govalidator's existence check which may perform DNS lookups.
//
// Example:
//
//	email := r.Form.Get("email")
//	validator.IsEmailInPublicDomain("email", email)
//
// Checks if email domain is reachable and public
func (v *Validation) IsEmailInPublicDomain(field, value string) {
	if !govalidator.IsExistingEmail(value) {
		v.AddError(field, "Invalid email address")
	}
}

// IsInt validates that a field contains a valid integer value.
// Uses strconv.Atoi for parsing validation.
//
// Example:
//
//	quantity := r.Form.Get("quantity")
//	validator.IsInt("quantity", quantity)
//
// Accepts "123", "-45", but rejects "12.5", "abc"
func (v *Validation) IsInt(field, value string) {
	_, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(field, "This field must be a integer")
	}
}

// IsFloat validates that a field contains a valid floating-point number.
// Uses strconv.ParseFloat with 64-bit precision for validation.
//
// Example:
//
//	price := r.Form.Get("price")
//	validator.IsFloat("price", price)
//
// Accepts "12.99", "0.5", "123", but rejects "abc", "12.34.56"
func (v *Validation) IsFloat(field, value string) {
	_, err := strconv.ParseFloat(value, 64)
	if err != nil {
		v.AddError(field, "This field must be a floating point number")
	}
}

// IsDateISO validates that a field contains a valid date in ISO format (YYYY-MM-DD).
// Uses Go's time.Parse with the standard date layout for validation.
//
// Example:
//
//		birthDate := r.Form.Get("birth_date")
//		validator.IsDateISO("birth_date", birthDate)
//	 Accepts "2023-12-25", "1990-01-01", but rejects "12/25/2023", "invalid"
func (v *Validation) IsDateISO(field, value string) {
	_, err := time.Parse("2006-01-02", value)
	if err != nil {
		v.AddError(field, "This field must be a date in the form of YYYY-MM-DD")
	}
}

// StringLength validates that a field's value is within the specified length range.
//
// Example:
//
//	username := r.Form.Get("username")
//	validator.StringLength("username", username, 3, 50) // between 3-50 chars
//	validator.StringLength("username", username, 0, 50) // max 50 chars
func (v *Validation) StringLength(field, value string, minLength, maxLength int) {
	length := len(value)

	if length < minLength {
		v.AddError(field, fmt.Sprintf("The %s field must be at least %d characters", field, minLength))
	} else if length > maxLength {
		v.AddError(field, fmt.Sprintf("The %s field must not exceed %d characters", field, maxLength))
	}
}

// NoSpaces validates that a field contains no whitespace characters.
// Useful for usernames, slugs, or other fields that shouldn't contain spaces.
//
// Example:
//
//	username := r.Form.Get("username")
//	validator.NoSpaces("username", username)
//
// Accepts "john_doe", "user123", but rejects "john doe", "user name"
func (v *Validation) NoSpaces(field, value string) {
	if govalidator.HasWhitespace(value) {
		v.AddError(field, "Spaces are not allowed")
	}
}

// NotEmpty validates that a field is not empty after trimming whitespace.
// Accepts optional custom error message, otherwise uses default message.
//
// Example:
//
//	name := r.Form.Get("name")
//	validator.NotEmpty("name", name)
//
// Or with custom message:
//
//	validator.NotEmpty("name", name, "Name cannot be blank")
func (v *Validation) NotEmpty(field, value string, message ...string) {
	if strings.TrimSpace(value) == "" {
		vs := fmt.Sprintf("The %s field must contain a value.", field)
		if len(message) > 0 {
			vs = message[0]
		}
		v.AddError(field, vs)
	}
}

// Password validates that a password meets security requirements including minimum length,
// mixed case letters. Default minimum length is 12 characters.
//
// Example:
//
//	password := r.Form.Get("password")
//	validator.Password("password", password)        // Uses default 12 char minimum
//	validator.Password("password", password, 8)     // Uses custom 8 char minimum
//
// Requires uppercase, lowercase, and minimum length
func (v *Validation) Password(field string, value string, length ...int) {
	minLength := 12
	if len(length) > 0 {
		minLength = length[0]
	}

	if len(value) < minLength {
		message := fmt.Sprintf("The field does not meet the minimum length of %d characters", minLength)
		v.AddError(field, message)
	}

	// Check for mixed case
	hasUpper := false
	hasLower := false
	for _, char := range value {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsLower(char) {
			hasLower = true
		}
	}

	if !hasUpper {
		v.AddError(field, "The field must contain a uppercase character")
	}
	if !hasLower {
		v.AddError(field, "The field must contain a lowercase character")
	}
}

// PasswordUncompromised checks if a password appears in known data breaches using the
// HaveIBeenPwned API with k-anonymity (only sends first 5 chars of SHA1 hash).
// Optional threshold parameter sets minimum breach count to trigger error (default: 1).
//
// Example:
//
//	password := r.Form.Get("password")
//	validator.PasswordUncompromised("password", password)     // Any breach count fails
//	validator.PasswordUncompromised("password", password, 5)  // Only fails if seen 5+ times
//
// Checks against HaveIBeenPwned database securely
func (v *Validation) PasswordUncompromised(field string, value string, threshold ...int) {
	thresholdVerifier := 1
	if len(threshold) > 0 {
		thresholdVerifier = threshold[0]
	}

	// Create SHA1 hash of password
	hasher := sha1.New()
	hasher.Write([]byte(value))
	hash := strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))
	hashPrefix := hash[0:5]
	hashSuffix := hash[5:]

	// Query HaveIBeenPwned API with k-anonymity
	uri := fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", hashPrefix)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Printf("Error creating request for password validation: %s\n", err)
		return
	}
	req.Header.Set("Add-Padding", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making password validation request: %s\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Password validation API returned status %d. Unable to verify password.\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading password validation response: %s\n", err)
		return
	}

	// Parse response and check for our password hash
	hashSuffixes := strings.Split(string(body), "\n")
	for _, suffix := range hashSuffixes {
		parts := strings.Split(suffix, ":")
		if len(parts) != 2 {
			continue
		}

		pwnedHash := strings.TrimSpace(parts[0])
		pwnedCount, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			continue
		}

		if pwnedHash == hashSuffix && pwnedCount >= thresholdVerifier {
			v.AddError(field, "The password provided was discovered in a recent data leak; please select another password.")
			break
		}
	}
}
