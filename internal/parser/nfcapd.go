package parser

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type FlowData struct {
	Timestamp time.Time
	SrcIP     string
	DstIP     string
	SrcPort   int
	DstPort   int
	Protocol  int
	Packets   int64
	Bytes     int64
}

func ParseNfcapdFile(filePath string) ([]FlowData, error) {
	cmd := exec.Command("nfdump", "-r", filePath, "-o", "fmt:%ts,%sa,%da,%sp,%dp,%pr,%pkt,%byt")

	output, err := cmd.Output()
	if err != nil {
		// Игнорируем ошибку, если nfdump ничего не выводит, но завершается с кодом 1
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 && len(output) == 0 {
				return []FlowData{}, nil
			}
		}
		return nil, fmt.Errorf("failed to execute nfdump for file %s: %v, output: %s", filePath, err, string(output))
	}

	var flows []FlowData
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Summary") || strings.HasPrefix(line, "Time") || strings.HasPrefix(line, "Total") || line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) != 8 {
			log.Printf("Skipping malformed line: %s", line)
			continue
		}

		ts, err := time.Parse("2006-01-02 15:04:05.000", strings.TrimSpace(parts[0]))
		if err != nil {
			log.Printf("Skipping line due to invalid timestamp: %s", line)
			continue
		}

		srcPort, _ := strconv.Atoi(strings.TrimSpace(parts[3]))
		dstPort, _ := strconv.Atoi(strings.TrimSpace(parts[4]))
		protocol, _ := getProtocolNumber(strings.TrimSpace(parts[5]))
		packets, _ := strconv.ParseInt(strings.TrimSpace(parts[6]), 10, 64)
		bytes, _ := strconv.ParseInt(strings.TrimSpace(parts[7]), 10, 64)

		flows = append(flows, FlowData{
			Timestamp: ts,
			SrcIP:     strings.TrimSpace(parts[1]),
			DstIP:     strings.TrimSpace(parts[2]),
			SrcPort:   srcPort,
			DstPort:   dstPort,
			Protocol:  protocol,
			Packets:   packets,
			Bytes:     bytes,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading nfdump output: %v", err)
	}

	return flows, nil
}

func getProtocolNumber(proto string) (int, error) {
	switch strings.ToUpper(proto) {
	case "TCP":
		return 6, nil
	case "UDP":
		return 17, nil
	case "ICMP":
		return 1, nil
	default:
		return strconv.Atoi(proto)
	}
}
