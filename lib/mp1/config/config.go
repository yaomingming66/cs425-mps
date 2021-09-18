package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	ConfigItems []ConfigItem
}

type ConfigItem struct {
	NodeID   string
	NodeHost string
	NodePort string
}

func ConfigParser(path string) (config *Config, err error) {
	configItems, err := ConfigItemsParser(path)
	if err != nil {
		return nil, err
	}
	return &Config{
		ConfigItems: configItems,
	}, nil
}

func ConfigItemsParser(path string) (configItem []ConfigItem, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	configItem = make([]ConfigItem, 0)
	configItemCount := 0

	if scanner.Scan() {
		firstLine := scanner.Text()
		firstLine = strings.TrimSpace(firstLine)
		configItemCount, err = strconv.Atoi(firstLine)
		if err != nil {
			return nil, err
		}
	}

	i := 0
	for scanner.Scan() && i <= configItemCount {
		i += 1
		line := scanner.Text()
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) != 3 {
			log.Errorf("invalid input format, skip")
			continue
		}

		nodeIDStr := fields[0]
		nodeHostStr := fields[1]
		nodePortStr := fields[2]

		configItem = append(configItem, ConfigItem{
			NodeID:   nodeIDStr,
			NodeHost: nodeHostStr,
			NodePort: nodePortStr,
		})
	}

	err = scanner.Err()

	if err != nil {
		return nil, err
	}

	return configItem, nil
}
