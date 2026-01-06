package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ReviewVocabulary allows user to interactively select words to include in the deck
func ReviewVocabulary(items []VocabularyItem) []VocabularyItem {
	if len(items) == 0 {
		return items
	}

	reader := bufio.NewReader(os.Stdin)
	var selected []VocabularyItem

	fmt.Println("\n=== Review vocabulary ===")
	fmt.Println("For each word: [Y]es to add, [N]o to skip, [Q]uit review (add remaining)")
	fmt.Println()

	for i, item := range items {
		fmt.Printf("─────────────────────────────────────\n")
		fmt.Printf("[%d/%d]\n\n", i+1, len(items))
		fmt.Printf("  Word:       %s\n", item.Word)
		fmt.Printf("  IPA:        /%s/\n", item.IPA)
		fmt.Printf("  Definition: %s\n", item.Definition)
		fmt.Printf("  Example:    %s\n", item.ExampleEN)
		fmt.Printf("              %s\n", item.ExampleRU)
		fmt.Println()

		for {
			fmt.Print("Add to deck? [Y/n/q]: ")
			input, err := reader.ReadString('\n')
			if err != nil {
				// On error, default to yes
				selected = append(selected, item)
				break
			}

			input = strings.TrimSpace(strings.ToLower(input))

			switch input {
			case "", "y", "yes":
				selected = append(selected, item)
				fmt.Println("✓ Added")
				goto next
			case "n", "no":
				fmt.Println("✗ Skipped")
				goto next
			case "q", "quit":
				// Add current and all remaining
				selected = append(selected, items[i:]...)
				fmt.Printf("\nAdded remaining %d words\n", len(items)-i)
				return selected
			default:
				fmt.Println("Invalid input. Use Y, N, or Q")
			}
		}
	next:
	}

	fmt.Printf("\n=== Selected %d of %d words ===\n\n", len(selected), len(items))
	return selected
}
