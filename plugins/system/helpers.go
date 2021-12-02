package system

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func cpuSample() (uint64, uint64, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}

	var idle, total uint64

	r := bufio.NewReader(f)

	for {
		line, _, err := r.ReadLine()
		if err != nil {
			return 0, 0, err
		}

		fields := strings.Fields(string(line))
		if fields[0] != "cpu" {
			continue
		}

		for i := 1; i < len(fields); i++ {
			value, err := strconv.ParseUint(fields[i], 10, 64)
			if err != nil {
				return 0, 0, err
			}

			total += value

			if i == 4 {
				idle = value
			}
		}

		return idle, total, nil
	}
}
