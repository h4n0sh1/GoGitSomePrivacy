package main

import (
	"fmt"
	"log"
	"strings"
)

// Simulates the new auto-splitting logic
func testNameSplitting(fullName string, exactMatch bool) {
	var firstName, lastName string

	fmt.Printf("\n=== Testing: %q ===\n", fullName)
	fmt.Printf("Exact match flag: %v\n\n", exactMatch)

	// Auto-split full name into first and last names for better detection
	// unless --exact flag is used
	if fullName != "" && !exactMatch && firstName == "" && lastName == "" {
		parts := strings.Fields(fullName)
		if len(parts) >= 2 {
			firstName = parts[0]
			lastName = parts[len(parts)-1]
			log.Printf("Auto-detecting: first name=%q, last name=%q (use --exact to disable)", firstName, lastName)
		} else {
			log.Printf("Warning: Full name %q has fewer than 2 parts, cannot auto-split", fullName)
		}
	}

	// Show what will be searched
	fmt.Println("\nSearching for:")
	if fullName != "" {
		fmt.Printf("  ✓ Full name: %q\n", fullName)
	}
	if firstName != "" {
		fmt.Printf("  ✓ First name: %q\n", firstName)
	}
	if lastName != "" {
		fmt.Printf("  ✓ Last name: %q\n", lastName)
	}

	// Show example matches
	fmt.Println("\nWould match commits containing:")
	if fullName != "" {
		fmt.Printf("  - \"Fixed bug reported by %s\"\n", fullName)
	}
	if firstName != "" {
		fmt.Printf("  - \"Thanks %s for the contribution\"\n", firstName)
	}
	if lastName != "" {
		fmt.Printf("  - \"As suggested by %s\"\n", lastName)
	}
}

func main() {
	fmt.Println("GoGitSomePrivacy - Name Splitting Test")
	fmt.Println("======================================")

	// Test Case 1: Normal usage (default behavior)
	testNameSplitting("John Doe", false)

	// Test Case 2: Exact match only
	testNameSplitting("John Doe", true)

	// Test Case 3: Complex name
	testNameSplitting("Mary Jane Watson", false)

	// Test Case 4: Single word (edge case)
	testNameSplitting("Prince", false)

	// Test Case 5: Hyphenated name
	testNameSplitting("Jean-Claude Van Damme", false)

	fmt.Println("\n======================================")
	fmt.Println("Test completed!")
}
