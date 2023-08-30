package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
)

type InventorySection struct {
	Name    string
	Entries []string
}

func main() {
	// List inventory files in the inventory directory
	inventoryFiles, err := listInventoryFiles("inventory")
	if err != nil {
		fmt.Println("Error listing inventory files:", err)
		return
	}

	// Create a prompt for selecting an inventory file
	filePrompt := promptui.Select{
		Label: "Select an inventory file",
		Items: inventoryFiles,
	}

	_, selectedFile, err := filePrompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	// Read the selected inventory file
	file, err := os.Open(filepath.Join("inventory", selectedFile))
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var sections []InventorySection

	// Regular expression to match sections
	sectionPattern := regexp.MustCompile(`^\[(.*?)\]`)

	var currentSection InventorySection

	// Scan through the lines and extract section names and entries
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		match := sectionPattern.FindStringSubmatch(line)
		if len(match) > 1 {
			if currentSection.Name != "" {
				sections = append(sections, currentSection)
			}
			currentSection = InventorySection{Name: match[1]}
		} else if currentSection.Name != "" {
			currentSection.Entries = append(currentSection.Entries, line)
		}
	}
	if currentSection.Name != "" {
		sections = append(sections, currentSection)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Create a prompt for selecting sections
	prompt := promptui.Select{
		Label: "Select a section",
		Items: getSectionNames(sections),
		Size:  10,
	}

	index, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	selectedSection := sections[index]

	fmt.Println("Selected section:", selectedSection.Name)
	fmt.Println(strings.Join(selectedSection.Entries, "\n"))

	// Prompt for yes/no
	promptYesNo := promptui.Prompt{
		Label:     "Do you want to run a task for this group? (yes/no)",
		IsConfirm: true,
	}

	result, err := promptYesNo.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	if result == "yes" || result == "y" {
		fmt.Println("Running task...")

		// Execute the "ls -al" command
		sectionName := selectedSection.Name
		ansibleCommand := fmt.Sprintf("ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i %s create_%s_instance.yaml -v", selectedFile, sectionName)
		cmd := exec.Command("bash", "-c", ansibleCommand)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("Command failed: %v\n", err)
		}
		// You can add your task execution code here
	} else {
		fmt.Println("Task not run.")
	}
}

func listInventoryFiles(directoryPath string) ([]string, error) {
	var files []string

	// List files in the "inventory" directory
	dirEntries, err := os.ReadDir("inventory")
	if err != nil {
		return files, err
	}
	// Filter for .elk.instance files
	for _, entry := range dirEntries {
		files = append(files, entry.Name())
	}

	return files, nil
}

func getSectionNames(sections []InventorySection) []string {
	names := make([]string, len(sections))
	for i, section := range sections {
		names[i] = section.Name
	}
	return names
}
