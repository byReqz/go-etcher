package etcher

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// getBlockdeviceSize gets the size of a blockdevice from the kernel filesystem.
func getBlockdeviceSize(name string) (size int, err error) {
	sizefile, err := os.Open("/sys/block/" + name + "/size")
	if err != nil {
		return
	}
	sizeread, err := io.ReadAll(sizefile)
	if err != nil {
		return
	}
	_ = sizefile.Close()

	sizestring := strings.TrimSuffix(string(sizeread), "\n")
	size, err = strconv.Atoi(sizestring)
	if err != nil {
		return
	}
	size = size * 512 // multiply blockcount by blocksize

	return size, err
}

// GetBlockdevices gets the available blockdevices for reading from/writing to.
func GetBlockdevices() (devices []Blockdevice, err error) {
	block, err := os.ReadDir("/sys/block")
	if err != nil {
		return
	} else if len(block) == 0 {
		return devices, fmt.Errorf("no blockdevices found")
	}

	for _, device := range block {
		var dev Blockdevice

		dev.Path = "/dev/" + device.Name()

		if strings.Contains(device.Name(), "sd") {
			dev.Type = "sata"
		} else if strings.Contains(device.Name(), "hd") {
			dev.Type = "ide"
		} else if strings.Contains(device.Name(), "nvme") {
			dev.Type = "nvme"
		} else if strings.Contains(device.Name(), "dm") {
			dev.Type = "dmcrypt"
		} else if strings.Contains(device.Name(), "md") {
			dev.Type = "raid"
		}

		dev.Size, err = getBlockdeviceSize(device.Name())
		if err != nil {
			return
		}

		devices = append(devices, dev)
	}

	return devices, nil
}
