package helpers

import (
	"testing"
)

func TestSanitize(t *testing.T) {
	h := &Helpers{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "XSS script tag",
			input:    `<script>alert('xss')</script>Hello`,
			expected: "Hello",
		},
		{
			name:     "XSS with event handler",
			input:    `<div onclick="malicious()">Click me</div>`,
			expected: "&lt;div&gt;Click me&lt;/div&gt;",
		},
		{
			name:     "JavaScript protocol",
			input:    `<a href="javascript:alert('xss')">Link</a>`,
			expected: "&lt;a href=&#34;&#39;xss&#39;&#34;&gt;Link&lt;/a&gt;",
		},
		{
			name:     "Path traversal",
			input:    `../../../etc/passwd`,
			expected: "etc/passwd",
		},
		{
			name:     "LDAP injection",
			input:    `admin)(uid=*`,
			expected: "adminuid=*",
		},
		{
			name:     "Control characters",
			input:    "Hello\x00World\x08",
			expected: "HelloWorld",
		},
		{
			name:     "Complex attack combination",
			input:    `<script>alert('xss')</script>../../../etc/passwd admin)(uid=*`,
			expected: "etc/passwd adminuid=*",
		},
		{
			name:     "Leading and trailing whitespace",
			input:    "  Hello World  ",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.Sanitize(tt.input)
			if result != tt.expected {
				t.Errorf("Sanitize() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCleanXSS(t *testing.T) {
	h := &Helpers{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Script tag removal",
			input:    `<script>alert('xss')</script>Hello`,
			expected: "Hello",
		},
		{
			name:     "Script tag with attributes",
			input:    `<script type="text/javascript">alert('xss')</script>Content`,
			expected: "Content",
		},
		{
			name:     "Multiple script tags",
			input:    `<script>alert(1)</script>Safe<script>alert(2)</script>`,
			expected: "Safe",
		},
		{
			name:     "Event handler onclick",
			input:    `<div onclick="malicious()">Content</div>`,
			expected: "&lt;div&gt;Content&lt;/div&gt;",
		},
		{
			name:     "Event handler onload",
			input:    `<body onload="malicious()">Content</body>`,
			expected: "&lt;body&gt;Content&lt;/body&gt;",
		},
		{
			name:     "Event handler onmouseover",
			input:    `<span onmouseover="evil()">Text</span>`,
			expected: "&lt;span&gt;Text&lt;/span&gt;",
		},
		{
			name:     "JavaScript protocol in href",
			input:    `<a href="javascript:alert('xss')">Link</a>`,
			expected: "&lt;a href=&#34;&#39;xss&#39;)&#34;&gt;Link&lt;/a&gt;",
		},
		{
			name:     "JavaScript protocol uppercase",
			input:    `<a href="JAVASCRIPT:alert('xss')">Link</a>`,
			expected: "&lt;a href=&#34;&#39;xss&#39;)&#34;&gt;Link&lt;/a&gt;",
		},
		{
			name:     "HTML escaping",
			input:    `<div>Normal content</div>`,
			expected: "&lt;div&gt;Normal content&lt;/div&gt;",
		},
		{
			name:     "Quotes and special characters",
			input:    `Hello "world" & <friends>`,
			expected: "Hello &#34;world&#34; &amp; &lt;friends&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.CleanXSS(tt.input)
			if result != tt.expected {
				t.Errorf("CleanXSS() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCleanInjection(t *testing.T) {
	h := &Helpers{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Null byte removal",
			input:    "Hello\x00World",
			expected: "HelloWorld",
		},
		{
			name:     "Multiple control characters",
			input:    "Hello\x00\x08\x1F\x7FWorld",
			expected: "HelloWorld",
		},
		{
			name:     "LDAP injection parentheses",
			input:    `admin)(uid=*`,
			expected: "adminuid=*",
		},
		{
			name:     "LDAP injection ampersand",
			input:    `user&(objectClass=*)`,
			expected: "userobjectClass=*",
		},
		{
			name:     "LDAP injection pipe",
			input:    `admin||(uid=*)`,
			expected: "adminuid=*",
		},
		{
			name:     "LDAP injection exclamation",
			input:    `!admin`,
			expected: "admin",
		},
		{
			name:     "Complex LDAP injection",
			input:    `)(cn=*))(&(password=*`,
			expected: "cn=*password=*",
		},
		{
			name:     "NoSQL-like injection",
			input:    `{"$ne": null}`,
			expected: `{"$ne": null}`,
		},
		{
			name:     "Control chars with normal text",
			input:    "Start\x01Middle\x02End",
			expected: "StartMiddleEnd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.CleanInjection(tt.input)
			if result != tt.expected {
				t.Errorf("CleanInjection() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCleanPathTraversal(t *testing.T) {
	h := &Helpers{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Normal filename",
			input:    "document.txt",
			expected: "document.txt",
		},
		{
			name:     "Normal path",
			input:    "documents/file.txt",
			expected: "documents/file.txt",
		},
		{
			name:     "Basic path traversal",
			input:    `../../../etc/passwd`,
			expected: "etc/passwd",
		},
		{
			name:     "Windows path traversal",
			input:    `..\..\..\windows\system32`,
			expected: "windows\\system32",
		},
		{
			name:     "Mixed path separators",
			input:    `../documents/../../secret.txt`,
			expected: "documents/secret.txt",
		},
		{
			name:     "Multiple traversal attempts",
			input:    `../../../../../../../../etc/passwd`,
			expected: "etc/passwd",
		},
		{
			name:     "Path traversal in middle",
			input:    `documents/../../../etc/passwd`,
			expected: "documents/etc/passwd",
		},
		{
			name:     "Windows and Unix mixed",
			input:    `..\../etc/passwd`,
			expected: "etc/passwd",
		},
		{
			name:     "Path with spaces",
			input:    `../documents/my file.txt`,
			expected: "documents/my file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.CleanPathTraversal(tt.input)
			if result != tt.expected {
				t.Errorf("CleanPathTraversal() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCleanAll(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "All attack vectors combined",
			input:    "<script>alert('xss')</script>../../../etc/passwd admin)(uid=* Hello\x00World",
			expected: "etc/passwd adminuid=* HelloWorld",
		},
		{
			name:     "XSS with path traversal",
			input:    `<script>alert('../../../etc/passwd')</script>`,
			expected: "",
		},
		{
			name:     "Complex real-world example",
			input:    `<img src="x" onerror="alert(1)">../config.php?user=admin)(uid=*`,
			expected: "&lt;img src=&#34;x&#34;&gt;config.php?user=adminuid=*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanAll(tt.input)
			if result != tt.expected {
				t.Errorf("CleanAll() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkSanitize(b *testing.B) {
	h := &Helpers{}
	input := `<script>alert('xss')</script>../../../etc/passwd admin)(uid=* Hello\x00World`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Sanitize(input)
	}
}

func BenchmarkCleanXSS(b *testing.B) {
	h := &Helpers{}
	input := `<script>alert('xss')</script><div onclick="evil()">Content</div>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.CleanXSS(input)
	}
}

func BenchmarkCleanInjection(b *testing.B) {
	h := &Helpers{}
	input := `admin)(uid=*\x00\x08\x1F`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.CleanInjection(input)
	}
}

func BenchmarkCleanPathTraversal(b *testing.B) {
	h := &Helpers{}
	input := `../../../etc/passwd`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.CleanPathTraversal(input)
	}
}
