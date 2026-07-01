package helpers

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func createValidation() *Validation {
	return &Validation{
		Data:   make(url.Values),
		Errors: make(map[string]string),
	}
}

func createRequestWithForm(formData map[string]string) *http.Request {
	form := url.Values{}
	for key, value := range formData {
		form.Set(key, value)
	}

	req := httptest.NewRequest("POST", "/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Form = form
	return req
}

func TestCreateNewValidator(t *testing.T) {
	helpers := Helpers{}
	data := url.Values{}
	validator := helpers.NewValidator(data)
	if validator.Data == nil {
		t.Error("validator expected to have Data field and it did not")
	}

	if validator.Errors == nil {
		t.Error("validator expected to have Errors field and it did not")
	}
}

func TestValidation_ToString(t *testing.T) {
	v := createValidation()
	v.Errors["name"] = "Name is required"
	v.Errors["email"] = "Invalid email"

	result := v.ToString()
	if !strings.Contains(result, "Name is required") {
		t.Error("Expected toString to contain 'Name is required'")
	}
	if !strings.Contains(result, "Invalid email") {
		t.Error("Expected toString to contain 'Invalid email'")
	}
}

func TestValidation_AddError(t *testing.T) {
	v := createValidation()

	// Test adding new error
	v.AddError("firstName", "The :attribute field is required")
	expected := "The first name field is required"
	if v.Errors["firstName"] != expected {
		t.Errorf("Expected '%s', got '%s'", expected, v.Errors["firstName"])
	}

	// Test that duplicate keys don't overwrite
	v.AddError("firstName", "Different message")
	if v.Errors["firstName"] != expected {
		t.Error("AddError should not overwrite existing errors")
	}
}

func TestValidation_Has(t *testing.T) {
	v := createValidation()
	req := createRequestWithForm(map[string]string{
		"username": "testuser",
		"empty":    "",
	})

	// Test field with value
	if !v.Has("username", req) {
		t.Error("Has should return true for field with value")
	}

	// Test empty field
	if v.Has("empty", req) {
		t.Error("Has should return false for empty field")
	}

	// Test non-existent field
	if v.Has("nonexistent", req) {
		t.Error("Has should return false for non-existent field")
	}
}

func TestValidation_Required(t *testing.T) {
	v := createValidation()
	req := createRequestWithForm(map[string]string{
		"name":   "John",
		"empty":  "",
		"spaces": "   ",
	})

	v.Required(req, "name", "email", "empty", "spaces")

	// Should not have error for valid field
	if _, exists := v.Errors["name"]; exists {
		t.Error("Should not have error for field with value")
	}

	// Should have errors for missing fields
	if _, exists := v.Errors["email"]; !exists {
		t.Error("Should have error for missing field")
	}
	if _, exists := v.Errors["empty"]; !exists {
		t.Error("Should have error for empty field")
	}
	if _, exists := v.Errors["spaces"]; !exists {
		t.Error("Should have error for whitespace-only field")
	}
}

func TestValidation_RequiredJSON(t *testing.T) {
	type TestStruct struct {
		Name  string
		Email string
	}

	v := createValidation()
	testData := &TestStruct{Name: "John", Email: "john@example.com"}

	// Test with valid struct pointer
	v.RequiredJSON(testData, "Name", "Email", "NonExistent")

	if _, exists := v.Errors["Name"]; exists {
		t.Error("Should not have error for existing field Name")
	}
	if _, exists := v.Errors["Email"]; exists {
		t.Error("Should not have error for existing field Email")
	}
	if _, exists := v.Errors["NonExistent"]; !exists {
		t.Error("Should have error for non-existent field")
	}

	// Test with non-pointer
	v2 := createValidation()
	v2.RequiredJSON(TestStruct{}, "Name")
	if _, exists := v2.Errors["Name"]; !exists {
		t.Error("Should have error when not passing pointer")
	}
}

func TestValidation_HasJSON(t *testing.T) {
	type TestStruct struct {
		Name  string
		Email string
		Age   int
	}

	v := createValidation()
	testData := &TestStruct{Name: "John", Email: "", Age: 25}

	v.HasJSON(testData, "Name", "Email", "Age")

	// Name has value - should pass
	if _, exists := v.Errors["Name"]; exists {
		t.Error("Should not have error for field with value")
	}

	// Email is empty - should fail
	if _, exists := v.Errors["Email"]; !exists {
		t.Error("Should have error for empty string field")
	}

	// Age has value - should pass
	if _, exists := v.Errors["Age"]; exists {
		t.Error("Should not have error for non-zero int field")
	}
}

func TestValidation_Check(t *testing.T) {
	v := createValidation()

	// Test condition that passes
	v.Check(true, "field1", "Should not appear")
	if _, exists := v.Errors["field1"]; exists {
		t.Error("Should not add error when condition is true")
	}

	// Test condition that fails
	v.Check(false, "field2", "This should appear")
	if v.Errors["field2"] != "This should appear" {
		t.Error("Should add error when condition is false")
	}
}

func TestValidation_IsEmail(t *testing.T) {
	v := createValidation()

	// Test valid emails
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"user+tag@example.org",
	}

	for i, email := range validEmails {
		field := "email" + string(rune(i))
		v.IsEmail(field, email)
		if _, exists := v.Errors[field]; exists {
			t.Errorf("Should not have error for valid email: %s", email)
		}
	}

	// Test invalid emails
	invalidEmails := []string{
		"invalid.email",
		"@example.com",
		"user@",
		"",
	}

	for i, email := range invalidEmails {
		field := "invalid" + string(rune(i))
		v.IsEmail(field, email)
		if _, exists := v.Errors[field]; !exists {
			t.Errorf("Should have error for invalid email: %s", email)
		}
	}
}

func TestValidation_IsInt(t *testing.T) {
	v := createValidation()

	// Test valid integers
	validInts := []string{"123", "-45", "0"}
	for i, value := range validInts {
		field := "int" + string(rune(i))
		v.IsInt(field, value)
		if _, exists := v.Errors[field]; exists {
			t.Errorf("Should not have error for valid int: %s", value)
		}
	}

	// Test invalid integers
	invalidInts := []string{"12.5", "abc", "123abc", ""}
	for i, value := range invalidInts {
		field := "invalid" + string(rune(i))
		v.IsInt(field, value)
		if _, exists := v.Errors[field]; !exists {
			t.Errorf("Should have error for invalid int: %s", value)
		}
	}
}

func TestValidation_IsFloat(t *testing.T) {
	v := createValidation()

	// Test valid floats
	validFloats := []string{"12.99", "0.5", "123", "-45.67"}
	for i, value := range validFloats {
		field := "float" + string(rune(i))
		v.IsFloat(field, value)
		if _, exists := v.Errors[field]; exists {
			t.Errorf("Should not have error for valid float: %s", value)
		}
	}

	// Test invalid floats
	invalidFloats := []string{"abc", "12.34.56", "12,34", ""}
	for i, value := range invalidFloats {
		field := "invalid" + string(rune(i))
		v.IsFloat(field, value)
		if _, exists := v.Errors[field]; !exists {
			t.Errorf("Should have error for invalid float: %s", value)
		}
	}
}

func TestValidation_IsDateISO(t *testing.T) {
	v := createValidation()

	// Test valid ISO dates
	validDates := []string{"2023-12-25", "1990-01-01", "2000-02-29"}
	for i, date := range validDates {
		field := "date" + string(rune(i))
		v.IsDateISO(field, date)
		if _, exists := v.Errors[field]; exists {
			t.Errorf("Should not have error for valid date: %s", date)
		}
	}

	// Test invalid dates
	invalidDates := []string{"12/25/2023", "2023-13-01", "invalid", ""}
	for i, date := range invalidDates {
		field := "invalid" + string(rune(i))
		v.IsDateISO(field, date)
		if _, exists := v.Errors[field]; !exists {
			t.Errorf("Should have error for invalid date: %s", date)
		}
	}
}

func TestValidation_NoSpaces(t *testing.T) {
	v := createValidation()

	// Test values without spaces
	noSpaceValues := []string{"username", "test123", "user_name", "test-value"}
	for i, value := range noSpaceValues {
		field := "nospace" + string(rune(i))
		v.NoSpaces(field, value)
		if _, exists := v.Errors[field]; exists {
			t.Errorf("Should not have error for value without spaces: %s", value)
		}
	}

	// Test values with spaces
	spaceValues := []string{"user name", "test 123", " leading", "trailing "}
	for i, value := range spaceValues {
		field := "space" + string(rune(i))
		v.NoSpaces(field, value)
		if _, exists := v.Errors[field]; !exists {
			t.Errorf("Should have error for value with spaces: %s", value)
		}
	}
}

func TestValidation_NotEmpty(t *testing.T) {
	v := createValidation()

	// Test non-empty values
	v.NotEmpty("name", "John")
	if _, exists := v.Errors["name"]; exists {
		t.Error("Should not have error for non-empty value")
	}

	// Test empty values
	v.NotEmpty("empty", "")
	if _, exists := v.Errors["empty"]; !exists {
		t.Error("Should have error for empty value")
	}

	// Test whitespace-only value
	v.NotEmpty("spaces", "   ")
	if _, exists := v.Errors["spaces"]; !exists {
		t.Error("Should have error for whitespace-only value")
	}

	// Test with custom message
	v.NotEmpty("custom", "", "Custom error message")
	if v.Errors["custom"] != "Custom error message" {
		t.Error("Should use custom error message")
	}
}

func TestValidation_Password(t *testing.T) {
	v := createValidation()

	// Test valid password (default length)
	v.Password("password", "ValidPass123")
	if _, exists := v.Errors["password"]; exists {
		t.Error("Should not have error for valid password")
	}

	// Test password too short
	v2 := createValidation()
	v2.Password("short", "Short1")
	if _, exists := v2.Errors["short"]; !exists {
		t.Error("Should have error for password too short")
	}

	// Test password without uppercase
	v3 := createValidation()
	v3.Password("noupper", "validpassword123")
	if _, exists := v3.Errors["noupper"]; !exists {
		t.Error("Should have error for password without uppercase")
	}

	// Test password without lowercase
	v4 := createValidation()
	v4.Password("nolower", "VALIDPASSWORD123")
	if _, exists := v4.Errors["nolower"]; !exists {
		t.Error("Should have error for password without lowercase")
	}

	// Test with custom minimum length
	v5 := createValidation()
	v5.Password("custom", "Short1", 6)
	if _, exists := v5.Errors["custom"]; exists {
		t.Error("Should not have error for password meeting custom length requirement")
	}
}

func TestValidation_Valid(t *testing.T) {
	// Test validation with no errors
	v1 := createValidation()
	if !v1.Valid() {
		t.Error("Should be valid when no errors exist")
	}

	// Test validation with errors
	v2 := createValidation()
	v2.AddError("field", "Some error")
	if v2.Valid() {
		t.Error("Should not be valid when errors exist")
	}
}

// Note: PasswordUncompromised is not tested here as it makes external HTTP requests
// In a real application, you would mock the HTTP client or use dependency injection
// to make this method testable without making actual network calls.

func TestFormatFieldName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"firstName", "first name"},
		{"emailAddress", "email address"},
		{"ID", "id"},
		{"simpleField", "simple field"},
	}

	for _, tc := range testCases {
		result := formatFieldName(tc.input)
		if result != tc.expected {
			t.Errorf("formatFieldName(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

// ----------------------------------------------------------------------------
// IsEmailScriptSafe — golden-set regression vs PHP egulias/email-validator
// SpoofCheckValidation with only ICU SINGLE_SCRIPT enabled.
// ----------------------------------------------------------------------------

func TestIsEmailScriptSafe(t *testing.T) {
	cases := []struct {
		name  string
		email string
		want  bool // true = pass, false = fail (mixed scripts detected)
	}{
		// --- Pass cases --------------------------------------------------
		{"ascii_basic", "foo@example.com", true},
		{"ascii_subdomain", "alice123@subdomain.example.org", true},
		{"ascii_plus_tag", "email+tag@example.co", true},
		{"ascii_dotted_local", "name.surname@example.com", true},
		{"ascii_apostrophe", "O'Brien@example.com", true},
		{"pure_han_japanese", "山田@example.jp", true},
		{"han_hiragana_japanese", "山田太郎@日本.jp", true},
		{"hangul_korean", "김철수@한국.kr", true},
		{"han_latin_domain_japanese_context", "田中@日本.com", true},
		{"pure_cyrillic", "пример@пример.рф", true},

		// --- Fail cases: Cyrillic homograph spoofing in local part -------
		// 'а' (U+0430), 'о' (U+043E), 'і' (U+0456) are Cyrillic look-alikes.
		{"cyrillic_a_in_paypal", "pаypal@example.com", false},
		{"cyrillic_o_in_google", "gооgle@example.com", false},
		{"cyrillic_i_in_microsoft", "mіcrosoft@example.com", false},
		{"cyrillic_a_in_amazon", "аmazon@example.com", false},

		// --- Fail cases: Cyrillic homograph spoofing in domain -----------
		{"cyrillic_in_domain_google", "test@gооgle.com", false},
		{"cyrillic_mixed_in_domain", "hello@pаypа1.com", false},
		{"cyrillic_in_domain_only", "user@аmazon.com", false},

		// --- Fail cases: other mixed-script combinations -----------------
		{"greek_mixed_with_latin", "hεllo@example.com", false}, // Greek epsilon
		{"hebrew_mixed_with_latin", "shאlom@example.com", false},
		{"arabic_local_latin_domain", "مرحبا@example.com", false},
		{"thai_mixed_with_latin", "hเllo@example.com", false},
		{"devanagari_mixed_with_latin", "nनmste@example.com", false},

		// --- Fail cases: non-whitelisted multi-script combos -------------
		{"hangul_plus_han_no_latin_ok", "한中@example.com", true}, // Korean+Han is whitelisted-adjacent: Han+Hangul (no Latin) — still subset of {Latin, Han, Hangul}, passes.
		{"hangul_plus_hiragana_fail", "한ひ@example.com", false}, // Korean + Japanese hiragana — no whitelist
		{"cyrillic_plus_greek_fail", "аα@example.com", false},  // pure spoof: Cyrillic + Greek

		// --- Edge cases --------------------------------------------------
		{"empty_string", "", true}, // ASCII-only fast path; caller chains IsEmailRFC for emptiness check
		{"digits_and_punct", "1234567890+.@example.com", true},
		{"unicode_with_combining_marks_latin", "café@example.com", true}, // Inherited combining acute, still single Latin script
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := createValidation()
			got := v.IsEmailScriptSafe("email", tc.email)
			if got != tc.want {
				t.Fatalf("IsEmailScriptSafe(%q) = %v, want %v (errors=%v)",
					tc.email, got, tc.want, v.Errors)
			}
			if tc.want && len(v.Errors) != 0 {
				t.Fatalf("expected no error on pass, got %v", v.Errors)
			}
			if !tc.want {
				if _, ok := v.Errors["email"]; !ok {
					t.Fatalf("expected error on fail, got none")
				}
			}
		})
	}
}

// ----------------------------------------------------------------------------
// IsEmailMX — matches PHP egulias/email-validator DNSCheckValidation (the
// email:mx / email:dns rule).
//
// The malformed-input cases below never touch the network. The DNS cases are
// gated behind a short-timeout context and skip themselves when there's no
// resolver available, so the suite stays green offline and in CI sandboxes.
// ----------------------------------------------------------------------------

func TestIsEmailMX_MalformedInput(t *testing.T) {
	// These fail on parsing alone and must never hit DNS, so a cancelled
	// context is safe — if any of them tried to resolve, it would error out
	// the same way, but the point is they short-circuit before the lookup.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cases := []struct {
		name  string
		email string
	}{
		{"no_at_sign", "not-an-email"},
		{"trailing_at", "user@"},
		{"empty_string", ""},
		{"domain_is_bare_dot", "user@."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := createValidation()
			if got := v.IsEmailMX(ctx, "email", tc.email); got {
				t.Fatalf("IsEmailMX(%q) = true, want false", tc.email)
			}
			if _, ok := v.Errors["email"]; !ok {
				t.Fatalf("expected error on fail, got none")
			}
		})
	}
}

// hasResolver reports whether outbound DNS actually works in this environment.
// Used to skip the live-lookup cases rather than fail them offline.
func hasResolver(t *testing.T) bool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var r net.Resolver
	if _, err := r.LookupHost(ctx, "google.com"); err != nil {
		return false
	}
	return true
}

func TestIsEmailMX_LiveLookup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live DNS lookups in -short mode")
	}
	if !hasResolver(t) {
		t.Skip("no DNS resolver available; skipping live MX checks")
	}

	cases := []struct {
		name  string
		email string
		want  bool // true = domain can receive mail
	}{
		// gmail.com has real MX records.
		{"has_mx", "someone@gmail.com", true},
		// github.io serves web content with A records but publishes no MX,
		// so it's accepted via the RFC 5321 §5.1 address-record fallback.
		{"a_record_fallback", "someone@github.io", true},
		// example.com publishes an RFC 7505 null MX ("0 ."), explicitly opting
		// out of mail, so it must be rejected even though it has A records.
		{"null_mx_rejected", "someone@example.com", false},
		// A domain that resolves nowhere at all.
		{"nonexistent_domain", "someone@this-domain-does-not-exist-xyz-9f8a.invalid", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			v := createValidation()
			got := v.IsEmailMX(ctx, "email", tc.email)
			if got != tc.want {
				t.Fatalf("IsEmailMX(%q) = %v, want %v (errors=%v)",
					tc.email, got, tc.want, v.Errors)
			}
			if tc.want && len(v.Errors) != 0 {
				t.Fatalf("expected no error on pass, got %v", v.Errors)
			}
			if !tc.want {
				if _, ok := v.Errors["email"]; !ok {
					t.Fatalf("expected error on fail, got none")
				}
			}
		})
	}
}
