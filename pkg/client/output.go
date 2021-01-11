package client

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/acme-dns/acme-dns-client/pkg/integration"
)

func successMarker() string {
	return fmt.Sprintf("[%s*%s]", ANSI_GREEN, ANSI_CLEAR)
}

func infoMarker() string {
	return fmt.Sprintf("[%si%s]", ANSI_BLUE, ANSI_CLEAR)
}

func warningMarker() string {
	return fmt.Sprintf("[%sW%s]", ANSI_YELLOW, ANSI_CLEAR)
}

func errorMarker() string {
	return fmt.Sprintf("[%s!%s]", ANSI_RED, ANSI_CLEAR)
}

func debugMarker() string {
	return fmt.Sprintf("[%sD%s]", ANSI_YELLOW, ANSI_CLEAR)
}

func PrintError(input string, offset int) {
	padding := strings.Repeat(" ", offset)
	fmt.Printf("%s%s %s\n", padding, errorMarker(), input)
}

func PrintInfo(input string, offset int) {
	padding := strings.Repeat(" ", offset)
	fmt.Printf("%s%s %s\n", padding, infoMarker(), input)
}

func PrintWarning(input string, offset int) {
	padding := strings.Repeat(" ", offset)
	fmt.Printf("%s%s %s\n", padding, warningMarker(), input)
}

func PrintSuccess(input string, offset int) {
	padding := strings.Repeat(" ", offset)
	fmt.Printf("%s%s %s\n", padding, successMarker(), input)
}

func PrintDebug(input string, offset int) {
	padding := strings.Repeat(" ", offset)
	fmt.Printf("%s%s %s\n", padding, debugMarker(), input)
}

func YesNoPrompt(question string, defVal bool) bool {
	reader := bufio.NewReader(os.Stdin)
	if defVal {
		fmt.Printf("%s [Y/n]: ", question)
	} else {
		fmt.Printf("%s [y/N]: ", question)
	}
	inp, _ := reader.ReadString('\n')
	inp = strings.TrimSpace(inp)
	if strings.ToLower(inp) == "y" {
		return true
	} else if strings.ToLower(inp) == "n" {
		return false
	}
	return defVal
}

func (c *ConfigurationState) PrintACMEAccountInfo(accs []integration.ACMEAccount) {
	if len(accs) > 0 {
		fmt.Printf("\n - ACME accounts found on the system:\n")
		for _, v := range accs {
			PrintInfo(fmt.Sprintf("URI: \t%s", v.URI), 2)
			PrintInfo(fmt.Sprintf("Contact: \t%s", v.Contact), 2)
			PrintInfo(fmt.Sprintf("File: \t%s", v.FilePath), 2)
			PrintInfo(fmt.Sprintf("Client: \t%s", v.Client), 2)
			fmt.Printf(" --------\n")
		}
	}
}

func printPauseCounter(seconds int) {
	for i := 0; i < seconds; i++ {
		fmt.Fprintf(os.Stderr, "%sWaiting for %d seconds... Press Ctrl + C to abort and exit.", TERMINAL_CLEAR_LINE, seconds - i)
		time.Sleep(1 * time.Second)
	}
	fmt.Fprintf(os.Stderr, "%s", TERMINAL_CLEAR_LINE)
}