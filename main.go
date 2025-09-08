package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Store operator-specific phone number prefixes (use built-in default prefixes directly)
var (
	crawledMobile  []string // China Mobile prefixes
	crawledUnicom  []string // China Unicom prefixes
	crawledTelecom []string // China Telecom prefixes
)

// Config defines the structure of config.json (stores multiple 4-digit middle codes)
type Config struct {
	MiddleCodes []string `json:"middleCodes"`
}

func main() {
	// Initialize built-in operator prefixes (no web crawling needed, load directly)
	initDefaultSegments()
	fmt.Printf("Loaded built-in operator prefixes:\n")
	fmt.Printf("China Mobile: %d | China Unicom: %d | China Telecom: %d\n",
		len(crawledMobile), len(crawledUnicom), len(crawledTelecom))

	// Select 4-digit middle code input method
	scanner := bufio.NewScanner(os.Stdin)
	var middleCodes []string
	fmt.Println("\nPlease select 4-digit middle code input method:")
	fmt.Println("1. Read from config.json (file will be auto-created if it doesn't exist)")
	fmt.Println("2. Manual input via command line (separate multiple codes with commas, e.g., 0537,0100,0210)")

	for {
		fmt.Print("Enter option (1/2): ")
		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())
		switch choice {
		case "1":
			// Read from config file (auto-create if not exists)
			config, err := loadConfig()
			if err != nil {
				fmt.Printf("Config file processing failed: %v\n", err)
				continue // Re-select input method
			}
			middleCodes = config.MiddleCodes
			if len(middleCodes) == 0 {
				fmt.Println("Warning: middleCodes in config.json is empty, using default middle code [0537]")
				middleCodes = []string{"0537"} // Fallback default value
			}
			fmt.Printf("Successfully read %d 4-digit middle codes from config file: %v\n", len(middleCodes), middleCodes)
			goto generateStep // Proceed to phone number generation step
		case "2":
			// Manual input via command line
			var err error
			middleCodes, err = inputMiddleCodes(scanner)
			if err != nil {
				fmt.Printf("Input error: %v\n", err)
				continue
			}
			fmt.Printf("Manual input successful, total %d 4-digit middle codes: %v\n", len(middleCodes), middleCodes)
			goto generateStep // Proceed to phone number generation step
		default:
			fmt.Println("Invalid option, please enter 1 or 2")
		}
	}

generateStep:
	// Generate phone numbers and export to file
	err := generatePhoneNumbers(middleCodes)
	if err != nil {
		fmt.Printf("Phone number generation failed: %v\n", err)
		return
	}
	fmt.Println("\n✅ Phone numbers have been successfully exported to phonedict.txt")

	// Choose to exit or regenerate
	for {
		fmt.Print("\nExit program? (y/n, 'n' to reselect middle code input method): ")
		scanner.Scan()
		quitChoice := strings.TrimSpace(scanner.Text())
		switch quitChoice {
		case "y", "Y":
			fmt.Println("Exiting program...")
			return
		case "n", "N":
			fmt.Println("-------------------------- Restart --------------------------")
			main() // Re-execute main process
			return
		default:
			fmt.Println("Invalid input, please enter y or n")
		}
	}
}

// loadConfig reads config.json; auto-creates it with sample content if the file doesn't exist
func loadConfig() (Config, error) {
	var config Config
	configPath := "config.json"

	// Check if the file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File doesn't exist, auto-create and write sample content
		fmt.Printf("%s not found, creating automatically...\n", configPath)
		// Sample config (contains common regional middle codes)
		defaultConfig := Config{
			MiddleCodes: []string{"0537", "0100", "0210", "0755"}, // Jining Beijing, Shanghai, Shenzhen
		}
		// Generate formatted JSON (with indentation for easy editing)
		jsonData, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return config, fmt.Errorf("failed to generate default config: %v", err)
		}
		// Write to file (permission: read/write for owner, read-only for others)
		if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
			return config, fmt.Errorf("failed to create %s: %v", configPath, err)
		}
		fmt.Printf("✅ %s created successfully, sample middle codes: %v\n", configPath, defaultConfig.MiddleCodes)
		fmt.Println("Note: You can edit this file directly to modify the middleCodes list (must be 4-digit numbers)")
		// Return default config
		return defaultConfig, nil
	} else if err != nil {
		// Other file status errors (e.g., permission issues)
		return config, fmt.Errorf("failed to check %s status: %v", configPath, err)
	}

	// File exists, read and parse it
	file, err := os.Open(configPath)
	if err != nil {
		return config, fmt.Errorf("failed to open %s: %v", configPath, err)
	}
	defer file.Close()

	// Parse JSON content
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return config, fmt.Errorf("failed to parse %s format (check commas and quotes): %v", configPath, err)
	}

	// Validate all middle codes are 4-digit numbers
	validRegex := regexp.MustCompile(`^\d{4}$`)
	validCodes := []string{}
	for _, code := range config.MiddleCodes {
		if validRegex.MatchString(code) {
			validCodes = append(validCodes, code)
		} else {
			fmt.Printf("Warning: Invalid middle code %s in %s (must be 4-digit number), skipped\n", code, configPath)
		}
	}
	// Update to valid middle codes list
	config.MiddleCodes = validCodes

	return config, nil
}

// inputMiddleCodes handles manual input of multiple 4-digit middle codes via command line (separated by commas)
func inputMiddleCodes(scanner *bufio.Scanner) ([]string, error) {
	fmt.Print("Enter multiple 4-digit middle codes (separate with commas, e.g., 0537,0100,0210): ")
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return nil, fmt.Errorf("input cannot be empty")
	}

	// Split input by commas
	rawCodes := strings.Split(input, ",")
	var validCodes []string
	validRegex := regexp.MustCompile(`^\d{4}$`)

	// Filter invalid codes and remove duplicates
	seen := make(map[string]bool)
	for _, code := range rawCodes {
		code = strings.TrimSpace(code) // Remove spaces (compatible with accidental spaces in input)
		if validRegex.MatchString(code) && !seen[code] {
			validCodes = append(validCodes, code)
			seen[code] = true
		}
	}

	// Check if there are valid codes
	if len(validCodes) == 0 {
		return nil, fmt.Errorf("no valid middle codes detected (must enter 4-digit numbers separated by commas)")
	}

	return validCodes, nil
}

// initDefaultSegments loads built-in operator phone number prefixes (no web crawling required)
func initDefaultSegments() {
	crawledMobile = []string{"134", "135", "136", "137", "138", "139", "147", "150", "151", "152", "157", "158", "159", "178", "182", "183", "184", "187", "188", "198"}
	crawledUnicom = []string{"130", "131", "132", "145", "155", "156", "166", "175", "176", "185", "186"}
	crawledTelecom = []string{"133", "149", "153", "173", "177", "180", "181", "189", "199"}
}

// generatePhoneNumbers generates phone numbers and exports them to a file
func generatePhoneNumbers(middleCodes []string) error {
	// Create output file
	file, err := os.Create("phonedict.txt")
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Merge all operator prefixes
	allSegments := append(append(crawledMobile, crawledUnicom...), crawledTelecom...)
	totalSegments := len(allSegments)
	totalMiddle := len(middleCodes)
	totalNumbers := totalSegments * totalMiddle * 10000 // Each prefix × each middle code × 10000 suffixes (0000-9999)

	fmt.Printf("\n📱 Starting phone number generation:\n")
	fmt.Printf("Total prefixes: %d | Total middle codes: %d | Suffix range per combination: 0000-9999\n",
		totalSegments, totalMiddle)
	fmt.Printf("Estimated total numbers to generate: %d\n", totalNumbers)

	generatedCount := 0
	// Generation logic: 3-digit prefix + 4-digit middle code + 4-digit suffix (11-digit phone number)
	for _, seg := range allSegments {
		for _, middle := range middleCodes {
			for suffix := 0; suffix < 10000; suffix++ {
				// Format suffix to 4 digits (pad with leading zeros if needed, e.g., 123 → 0123)
				suffixStr := fmt.Sprintf("%04d", suffix)
				phone := seg + middle + suffixStr
				// Write to file (one phone number per line)
				if _, err := writer.WriteString(phone + "\n"); err != nil {
					return fmt.Errorf("failed to write to file: %v", err)
				}
				generatedCount++
				// Flush buffer every 10000 numbers to avoid high memory usage
				if generatedCount%10000 == 0 {
					writer.Flush()
					fmt.Printf("Generated: %d / %d\n", generatedCount, totalNumbers)
				}
			}
		}
	}

	fmt.Printf("✅ Generation completed! Actual total numbers generated: %d\n", generatedCount)
	return nil
}
