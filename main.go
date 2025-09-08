package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	crawledMobile  []string // China Mobile prefixes
	crawledUnicom  []string // China Unicom prefixes
	crawledTelecom []string // China Telecom prefixes
)

type Config struct {
	MiddleCodes []string `json:"middleCodes"`
}

func main() {
	initDefaultSegments()
	fmt.Printf("Loaded built-in operator prefixes:\n")
	fmt.Printf("China Mobile: %d | China Unicom: %d | China Telecom: %d\n",
		len(crawledMobile), len(crawledUnicom), len(crawledTelecom))

	scanner := bufio.NewScanner(os.Stdin)
	// 使用循环代替递归调用main，避免栈溢出和资源泄漏
	for {
		middleCodes, err := selectMiddleCodes(scanner)
		if err != nil {
			fmt.Printf("Failed to get middle codes: %v\n", err)
			continue
		}

		err = generatePhoneNumbers(middleCodes)
		if err != nil {
			fmt.Printf("Phone number generation failed: %v\n", err)
		} else {
			fmt.Println("\n✅ Phone numbers have been successfully exported to phonedict.txt")
		}

		// 询问是否退出
		if !askToContinue(scanner) {
			fmt.Println("Exiting program...")
			return
		}
		fmt.Println("-------------------------- Restart --------------------------")
	}
}

// 选择中间码输入方式，提取为独立函数
func selectMiddleCodes(scanner *bufio.Scanner) ([]string, error) {
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
			config, err := loadConfig()
			if err != nil {
				fmt.Printf("Config file processing failed: %v\n", err)
				continue
			}
			middleCodes = config.MiddleCodes
			if len(middleCodes) == 0 {
				fmt.Println("Warning: middleCodes in config.json is empty, using default middle code [0537]")
				middleCodes = []string{"0537"}
			}
			fmt.Printf("Successfully read %d 4-digit middle codes from config file: %v\n", len(middleCodes), middleCodes)
			return middleCodes, nil
		case "2":
			middleCodes, err := inputMiddleCodes(scanner)
			if err != nil {
				fmt.Printf("Input error: %v\n", err)
				continue
			}
			fmt.Printf("Manual input successful, total %d 4-digit middle codes: %v\n", len(middleCodes), middleCodes)
			return middleCodes, nil
		default:
			fmt.Println("Invalid option, please enter 1 or 2")
		}
	}
}

// 询问是否继续运行，提取为独立函数
func askToContinue(scanner *bufio.Scanner) bool {
	for {
		fmt.Print("\nExit program? (y/n, 'n' to reselect middle code input method): ")
		scanner.Scan()
		quitChoice := strings.TrimSpace(scanner.Text())
		switch quitChoice {
		case "y", "Y":
			return false
		case "n", "N":
			return true
		default:
			fmt.Println("Invalid input, please enter y or n")
		}
	}
}

func loadConfig() (Config, error) {
	var config Config
	configPath := "config.json"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("%s not found, creating automatically...\n", configPath)
		defaultConfig := Config{
			MiddleCodes: []string{"0537", "0100", "0210", "0755"},
		}
		jsonData, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return config, fmt.Errorf("failed to generate default config: %v", err)
		}
		if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
			return config, fmt.Errorf("failed to create %s: %v", configPath, err)
		}
		fmt.Printf("✅ %s created successfully, sample middle codes: %v\n", configPath, defaultConfig.MiddleCodes)
		fmt.Println("Note: You can edit this file directly to modify the middleCodes list (must be 4-digit numbers)")
		return defaultConfig, nil
	} else if err != nil {
		return config, fmt.Errorf("failed to check %s status: %v", configPath, err)
	}

	file, err := os.Open(configPath)
	if err != nil {
		return config, fmt.Errorf("failed to open %s: %v", configPath, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return config, fmt.Errorf("failed to parse %s format (check commas and quotes): %v", configPath, err)
	}

	validRegex := regexp.MustCompile(`^\d{4}$`)
	validCodes := []string{}
	for _, code := range config.MiddleCodes {
		if validRegex.MatchString(code) {
			validCodes = append(validCodes, code)
		} else {
			fmt.Printf("Warning: Invalid middle code %s in %s (must be 4-digit number), skipped\n", code, configPath)
		}
	}
	config.MiddleCodes = validCodes

	return config, nil
}

func inputMiddleCodes(scanner *bufio.Scanner) ([]string, error) {
	fmt.Print("Enter multiple 4-digit middle codes (separate with commas, e.g., 0537,0100,0210): ")
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return nil, fmt.Errorf("input cannot be empty")
	}

	rawCodes := strings.Split(input, ",")
	var validCodes []string
	validRegex := regexp.MustCompile(`^\d{4}$`)
	seen := make(map[string]bool)

	for _, code := range rawCodes {
		code = strings.TrimSpace(code)
		if validRegex.MatchString(code) && !seen[code] {
			validCodes = append(validCodes, code)
			seen[code] = true
		}
	}

	if len(validCodes) == 0 {
		return nil, fmt.Errorf("no valid middle codes detected (must enter 4-digit numbers separated by commas)")
	}

	return validCodes, nil
}

func initDefaultSegments() {
	crawledMobile = []string{"134", "135", "136", "137", "138", "139", "147", "150", "151", "152", "157", "158", "159", "178", "182", "183", "184", "187", "188", "198"}
	crawledUnicom = []string{"130", "131", "132", "145", "155", "156", "166", "175", "176", "185", "186"}
	crawledTelecom = []string{"133", "149", "153", "173", "177", "180", "181", "189", "199"}
}

func generatePhoneNumbers(middleCodes []string) error {
	file, err := os.Create("phonedict.txt")
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close() // 确保文件在函数退出时关闭
	writer := bufio.NewWriter(file)
	defer writer.Flush() // 确保缓冲区数据写入文件

	allSegments := append(append(crawledMobile, crawledUnicom...), crawledTelecom...)
	totalSegments := len(allSegments)
	totalMiddle := len(middleCodes)
	totalNumbers := totalSegments * totalMiddle * 10000

	fmt.Printf("\n📱 Starting phone number generation:\n")
	fmt.Printf("Total prefixes: %d | Total middle codes: %d | Suffix range per combination: 0000-9999\n",
		totalSegments, totalMiddle)
	fmt.Printf("Estimated total numbers to generate: %d\n", totalNumbers)

	generatedCount := 0
	for _, seg := range allSegments {
		for _, middle := range middleCodes {
			for suffix := 0; suffix < 10000; suffix++ {
				suffixStr := fmt.Sprintf("%04d", suffix)
				phone := seg + middle + suffixStr
				if _, err := writer.WriteString(phone + "\n"); err != nil {
					return fmt.Errorf("failed to write to file: %v", err)
				}
				generatedCount++
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
