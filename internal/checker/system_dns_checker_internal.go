package checker

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func currentSystemDNSes() ([]string, error) {
	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dnses := make([]string, 0, 3)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "nameserver" {
			continue
		}
		dnses = append(dnses, fields[1])
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(dnses) == 0 {
		return nil, fmt.Errorf("no system dns found")
	}
	return dnses, nil
}
