package gc

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func IsInteractiveStdin() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func PromptString(label string, current string, required bool) (string, error) {
	r := bufio.NewReader(os.Stdin)
	for {
		if current != "" {
			fmt.Fprintf(os.Stderr, "%s [%s]: ", label, current)
		} else {
			fmt.Fprintf(os.Stderr, "%s: ", label)
		}

		line, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}
		v := strings.TrimSpace(line)
		if v == "" {
			if current != "" {
				return current, nil
			}
			if required {
				continue
			}
			return "", nil
		}
		return v, nil
	}
}

func PromptSecret(label string, current string, required bool) (string, error) {
	for {
		if current != "" {
			fmt.Fprintf(os.Stderr, "%s [stored]: ", label)
		} else {
			fmt.Fprintf(os.Stderr, "%s: ", label)
		}

		if !IsInteractiveStdin() {
			return "", errors.New("stdin is not a terminal")
		}

		buf, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", err
		}
		secret := strings.TrimSpace(string(buf))
		if secret == "" {
			if current != "" {
				return current, nil
			}
			if required {
				continue
			}
			return "", nil
		}
		return secret, nil
	}
}
