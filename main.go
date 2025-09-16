package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type SearchRequest struct {
	Term          string   `json:"term"`
	Fields        []string `json:"fields"`
	Categories    []string `json:"categories,omitempty"`
	Wildcard      *bool    `json:"wildcard,omitempty"`
	CaseSensitive *bool    `json:"case_sensitive,omitempty"`
}

const apiURL = "https://breach.vip/api/search"
const maxResults = 10000
const passPrefix = "passwords_"
const emailPrefix = "emails_"

func main() {
	// CLI flags
	term := flag.String("term", "", "Search term (required)")
	fields := flag.String("fields", "domain", "Comma-separated fields (required)")
	categories := flag.String("categories", "", "Comma-separated categories (optional)") // What is the category? example shows "minecraft"
	wildcard := flag.Bool("wildcard", false, "Enable wildcard (optional)")
	caseSensitive := flag.Bool("case", false, "Case sensitive search (optional)")
	url := flag.String("url", apiURL, "API endpoint URL")
	out := flag.String("out", "output.json", "Output file path")
	flag.Parse()

	if *term == "" || *fields == "" {
		fmt.Println("Error: --term and --fields are required.")
		flag.Usage()
		os.Exit(1)
	}

	// Validate fields
	fieldArray := splitAndTrim(*fields)
	if err := verifyFields(fieldArray...); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Prepare request
	req := SearchRequest{
		Term:   *term,
		Fields: fieldArray,
	}
	if *categories != "" {
		req.Categories = splitAndTrim(*categories)
	}
	req.Wildcard = wildcard
	req.CaseSensitive = caseSensitive

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Println("Error marshaling request:", err)
		os.Exit(1)
	}
	fmt.Printf("Searching for %s on %s\n", *term, *url)
	// Send POST request
	resp, err := http.Post(*url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error sending request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Check Response Status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: received status code %d: %s\n", resp.StatusCode, resp.Status)
		os.Exit(1)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	// Save to file
	err = os.WriteFile(*out, respBody, 0644)
	if err != nil {
		fmt.Println("Error writing output file:", err)
		os.Exit(1)
	}
	fmt.Printf("Response saved to %s\n", *out)

	// Calculate results
	total, emails, passwords, err := parseResults(*out)
	if err != nil {
		fmt.Println("Error parsing results:", err)
		os.Exit(1)
	}
	// Save passwords and emails to files using prefixes
	if len(emails) > 0 {
		emailFile := emailPrefix + *out
		err = os.WriteFile(emailFile, []byte(strings.Join(emails, "\n")), 0644)
		if err != nil {
			fmt.Println("Error writing email file:", err)
			os.Exit(1)
		}
		fmt.Printf("Emails saved to %s\n", emailFile)
	}
	if len(passwords) > 0 {
		passFile := passPrefix + *out
		err = os.WriteFile(passFile, []byte(strings.Join(passwords, "\n")), 0644)
		if err != nil {
			fmt.Println("Error writing password file:", err)
			os.Exit(1)
		}
		fmt.Printf("Passwords saved to %s\n", passFile)
	}
	fmt.Printf("Emails: %d\n", len(emails))
	fmt.Printf("Passwords: %d\n", len(passwords))
	fmt.Printf("Total results: %d (Maximum: %d)\n", total, maxResults)
}

func splitAndTrim(s string) []string {
	var result []string
	for _, v := range bytes.Split([]byte(s), []byte{','}) {
		str := string(bytes.TrimSpace(v))
		if str != "" {
			result = append(result, strings.ToLower(str))
		}
	}
	return result
}

func verifyFields(fields ...string) error {
	// "email" "password" "domain" "username" "ip" "name" "uuid" "steamid" "phone" "discordid"
	validFields := map[string]struct{}{
		"email":     {},
		"password":  {},
		"domain":    {},
		"username":  {},
		"ip":        {},
		"name":      {},
		"uuid":      {},
		"steamid":   {},
		"phone":     {},
		"discordid": {},
	}

	for _, field := range fields {
		if _, ok := validFields[field]; !ok {
			return fmt.Errorf("invalid field: %s", field)
		}
	}
	return nil
}

// parseResults reads the resulting json file and returns results, email, and password counts
func parseResults(filename string) (total int, emails []string, passwords []string, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, nil, nil, err
	}
	// Try to parse as {"results": [...]}
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return 0, nil, nil, err
	}
	arr, ok := obj["results"].([]interface{})
	if !ok {
		return 0, nil, nil, fmt.Errorf("results field not found or not array")
	}
	total = len(arr)
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		// Email field
		if val, ok := m["email"]; ok {
			switch v := val.(type) {
			case string:
				emails = append(emails, v)
			case []interface{}:
				for _, e := range v {
					if es, ok := e.(string); ok {
						emails = append(emails, es)
					}
				}
			}
		}
		// Password field
		if val, ok := m["password"]; ok {
			switch v := val.(type) {
			case string:
				passwords = append(passwords, v)
			case []interface{}:
				for _, p := range v {
					if ps, ok := p.(string); ok {
						passwords = append(passwords, ps)
					}
				}
			}
		}
	}
	return total, emails, passwords, nil
}
