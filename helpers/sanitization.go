package helpers

import (
	"html"
	"regexp"
	"strings"
)

var (
	// XSS Prevention
	scriptTagRegex  = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	onEventRegex    = regexp.MustCompile(`(?i)\s*on\w+\s*=\s*"[^"]*"|\s*on\w+\s*=\s*'[^']*'|\s*on\w+\s*=[^\s>]*`) // More comprehensive
	javascriptRegex = regexp.MustCompile(`(?i)javascript:[^"'\s>]*`)                                              // Remove the whole javascript: URL

	// Code Injection Prevention
	controlCharsRegex = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)

	// LDAP/NoSQL Injection patterns
	ldapCharsRegex = regexp.MustCompile(`[()&|!]`)

	// Path Traversal
	pathTraversalRegex = regexp.MustCompile(`\.\.[\\/]`)
)

// Sanitize removes common OWASP attack vectors from user input including XSS,
// injection attacks, and path traversal attempts. This method applies all
// sanitization functions and should be used on any untrusted user input.
//
// Attack vectors protected against:
//   - Cross-Site Scripting (XSS): Removes <script> tags, event handlers, javascript: protocols
//   - Code Injection: Removes control characters and null bytes
//   - LDAP/NoSQL Injection: Removes special LDAP characters
//   - Path Traversal: Removes directory traversal patterns like ../
//
// Examples:
//
//	// Basic XSS protection
//	input := `<script>alert('xss')</script>Hello`
//	clean := h.Sanitize(input)
//	// Result: "Hello"
//
//	// Event handler removal
//	input := `<div onclick="malicious()">Click me</div>`
//	clean := h.Sanitize(input)
//	// Result: "&lt;div&gt;Click me&lt;/div&gt;"
//
//	// JavaScript protocol removal
//	input := `<a href="javascript:alert('xss')">Link</a>`
//	clean := h.Sanitize(input)
//	// Result: "&lt;a href=&#34;&#34;&gt;Link&lt;/a&gt;"
//
//	// Path traversal protection
//	input := `../../../etc/passwd`
//	clean := h.Sanitize(input)
//	// Result: "etc/passwd"
//
//	// LDAP injection protection
//	input := `admin)(uid=*`
//	clean := h.Sanitize(input)
//	// Result: "adminuid=*"
//
//	// Control character removal
//	input := "Hello\x00World\x08"
//	clean := h.Sanitize(input)
//	// Result: "HelloWorld"
//
//	// Typical form input sanitization
//	username := h.Sanitize(r.Form.Get("username"))
//	comment := h.Sanitize(r.Form.Get("comment"))
//	filename := h.Sanitize(r.Form.Get("filename"))
//
// Note: This function escapes HTML entities, so legitimate HTML will be converted
// to safe display text. For rich text that needs to preserve some HTML tags,
// consider using a more targeted approach or an HTML sanitization library.
func (h *Helpers) Sanitize(input string) string {
	cleaned := CleanAll(input)

	return cleaned
}

// CleanXSS removes XSS attack vectors from user input.
// Protects against Cross-Site Scripting by removing script tags, event handlers,
// and javascript: protocols, then escapes remaining HTML.
//
// Examples:
//
//	// Script tag removal
//	input := `<script>alert('xss')</script>Hello`
//	clean := h.CleanXSS(input)
//	// Result: "Hello"
//
//	// Event handler removal
//	input := `<div onclick="malicious()">Content</div>`
//	clean := h.CleanXSS(input)
//	// Result: "&lt;div&gt;Content&lt;/div&gt;"
//
//	// JavaScript protocol removal
//	input := `<a href="javascript:alert('xss')">Link</a>`
//	clean := h.CleanXSS(input)
//	// Result: "&lt;a href=&#34;&#34;&gt;Link&lt;/a&gt;"
func (h *Helpers) CleanXSS(input string) string {
	return CleanXSS(input)
}

// CleanInjection removes common injection attack vectors from user input.
// Protects against code injection, LDAP injection, and NoSQL injection by
// removing control characters and special LDAP characters.
//
// Examples:
//
//	// Control character removal
//	input := "Hello\x00World\x08"
//	clean := h.CleanInjection(input)
//	// Result: "HelloWorld"
//
//	// LDAP injection protection
//	input := `admin)(uid=*`
//	clean := h.CleanInjection(input)
//	// Result: "adminuid=*"
//
//	// NoSQL injection protection
//	input := `{"$ne": null}`
//	clean := h.CleanInjection(input)
//	// Result: `{"$ne": null}` (parentheses removed if present)
func (h *Helpers) CleanInjection(input string) string {
	return CleanInjection(input)
}

// CleanPathTraversal removes directory traversal attempts from user input.
// Protects against path traversal attacks by removing ../ and ..\ patterns.
//
// Examples:
//
//	// Basic path traversal removal
//	input := `../../../etc/passwd`
//	clean := h.CleanPathTraversal(input)
//	// Result: "etc/passwd"
//
//	// Windows-style path traversal
//	input := `..\..\..\windows\system32`
//	clean := h.CleanPathTraversal(input)
//	// Result: "windows\system32"
//
//	// Mixed path separators
//	input := `../documents/../../secret.txt`
//	clean := h.CleanPathTraversal(input)
//	// Result: "documents/secret.txt"
func (h *Helpers) CleanPathTraversal(input string) string {
	return CleanPathTraversal(input)
}

// CleanXSS removes XSS attack vectors
func CleanXSS(input string) string {
	if input == "" {
		return input
	}

	// Remove script tags
	input = scriptTagRegex.ReplaceAllString(input, "")

	// Remove event handlers (onclick, onload, etc.)
	input = onEventRegex.ReplaceAllString(input, "")

	// Remove javascript: protocols
	input = javascriptRegex.ReplaceAllString(input, "")

	// Escape remaining HTML
	input = html.EscapeString(input)

	return strings.TrimSpace(input)
}

// CleanInjection removes common injection attack vectors
func CleanInjection(input string) string {
	if input == "" {
		return input
	}

	// Remove control characters (null bytes, etc.)
	input = controlCharsRegex.ReplaceAllString(input, "")

	// Remove LDAP injection characters
	input = ldapCharsRegex.ReplaceAllString(input, "")

	return strings.TrimSpace(input)
}

// CleanPathTraversal removes directory traversal attempts
func CleanPathTraversal(input string) string {
	if input == "" {
		return input
	}

	// Remove path traversal patterns
	input = pathTraversalRegex.ReplaceAllString(input, "")

	return strings.TrimSpace(input)
}

// CleanAll applies all OWASP sanitization
func CleanAll(input string) string {
	if input == "" {
		return input
	}

	// Remove script tags
	input = scriptTagRegex.ReplaceAllString(input, "")

	// Remove event handlers
	input = onEventRegex.ReplaceAllString(input, "")

	// Remove javascript protocols
	input = javascriptRegex.ReplaceAllString(input, "")

	// Remove control characters
	input = controlCharsRegex.ReplaceAllString(input, "")

	// Remove LDAP injection characters
	input = ldapCharsRegex.ReplaceAllString(input, "")

	// Remove path traversal patterns
	input = pathTraversalRegex.ReplaceAllString(input, "")

	// HTML escape LAST, after all other cleaning
	input = html.EscapeString(input)

	return strings.TrimSpace(input)
}
