package softraid

import (
	"bufio"
	"bytes"
	"fmt"
	"hardwareAnalyzer/utils"
	"os"
	"testing"
)

// Test CheckSoftRaid
func TestCheckSoftRaid(t *testing.T) {
	// Copy original functions content
	getSoftraidsOri := GetSoftraids
	// unmock functions content
	defer func() {
		GetSoftraids = getSoftraidsOri
	}()

	// Mocked function
	GetSoftraids = func() (*bufio.Scanner, *os.File, error) {
		//fmt.Println("-- Executing mocked GetSoftraids function")
		// Dont open any file, just fill a buffer to pass it to bufio.NewScanner
		var buffer bytes.Buffer
		buffer.WriteString(`
			Personalities : [raid0] [raid1] [raid6] [raid5] [raid4] [linear] [multipath] [raid10]
			md3 : active raid1 nvme0n1p3[0] nvme1n1p3[1]
				51198912 blocks [2/2] [UU]

			md2 : active raid1 nvme0n1p2[0] nvme1n1p2[1]
				523200 blocks [2/2] [UU]

			md5 : active raid0 nvme0n1p5[0] nvme1n1p5[1]
				2238671872 blocks 512k chunks

			unused devices: <none>
		`)

		scanner := bufio.NewScanner(&buffer)
		return scanner, nil, nil
	}

	softRaidCheck, err := CheckSoftRaid()
	if err != nil {
		t.Fatalf(`TestCheckSoftRaid: error: %s`, err)
	}
	if !softRaidCheck {
		t.Fatalf(`TestCheckSoftRaid: %v, want TRUE`, softRaidCheck)
	}
}

// Test CheckSoftRaid no kernel support
func TestCheckSoftRaidNoKernelSupport(t *testing.T) {
	// Copy original functions content
	getSoftraidsOri := GetSoftraids
	// unmock functions content
	defer func() {
		GetSoftraids = getSoftraidsOri
	}()

	// Mocked function
	GetSoftraids = func() (*bufio.Scanner, *os.File, error) {
		//fmt.Println("-- Executing mocked GetSoftraids function")
		return nil, nil, fmt.Errorf("RANDOM ERROR")
	}

	softRaidCheck, err := CheckSoftRaid()
	if err != nil {
		t.Fatalf(`TestCheckSoftRaidNoKernelSupport: error: %s`, err)
	}
	if softRaidCheck {
		t.Fatalf(`TestCheckSoftRaidNoKernelSupport: %v, want FALSE`, softRaidCheck)
	}
}

// Test CheckSoftRaid no softraids
func TestCheckSoftRaidNoSoftraids(t *testing.T) {
	// Copy original functions content
	getSoftraidsOri := GetSoftraids
	// unmock functions content
	defer func() {
		GetSoftraids = getSoftraidsOri
	}()

	// Mocked function
	GetSoftraids = func() (*bufio.Scanner, *os.File, error) {
		//fmt.Println("-- Executing mocked GetSoftraids function")
		// Dont open any file, just fill a buffer to pass it to bufio.NewScanner
		var buffer bytes.Buffer
		buffer.WriteString(``)
		scanner := bufio.NewScanner(&buffer)
		return scanner, nil, nil
	}

	softRaidCheck, err := CheckSoftRaid()
	if err != nil {
		t.Fatalf(`TestCheckSoftRaidNoSoftraids: error: %s`, err)
	}
	if softRaidCheck {
		t.Fatalf(`TestCheckSoftRaidNoSoftraids: %v, want FALSE`, softRaidCheck)
	}
}

// Test ProcessSoftRaid
func TestProcessSoftRaid(t *testing.T) {
	// Copy original functions content
	getSoftraidsOri := GetSoftraids
	getDiskData := utils.GetDiskData
	getDiskPartitionSize := utils.GetDiskPartitionSize // unmock functions content
	defer func() {
		GetSoftraids = getSoftraidsOri
		utils.GetDiskData = getDiskData
		utils.GetDiskPartitionSize = getDiskPartitionSize
	}()

	// Mocked function
	GetSoftraids = func() (*bufio.Scanner, *os.File, error) {
		//fmt.Println("-- Executing mocked GetSoftraids function")
		// Dont open any file, just fill a buffer to pass it to bufio.NewScanner
		var buffer bytes.Buffer
		buffer.WriteString(`
			Personalities : [raid0] [raid1] [raid6] [raid5] [raid4] [linear] [multipath] [raid10]
			md3 : active raid1 nvme0n1p3[0] nvme1n1p3[1]
				51198912 blocks [2/2] [UU]

			md2 : active raid1 nvme0n1p2[0] nvme1n1p2[1]
				523200 blocks [2/2] [UU]

			md5 : active raid0 nvme0n1p5[0] nvme1n1p5[1]
				2238671872 blocks 512k chunks

			unused devices: <none>
		`)

		scanner := bufio.NewScanner(&buffer)
		return scanner, nil, nil
	}

	// Mocked function
	utils.GetDiskData = func(diskDrive string) (string, string, string, string, error) {
		diskIntf := "SATA"
		diskMedium := "SSD"
		diskSerialNumber := ""
		diskModel := ""
		if diskDrive == "nvme0n1p3" {
			diskSerialNumber = "AAAAAAAA"
			diskModel = "AAAMODEL"
		} else if diskDrive == "nvme1n1p3" {
			diskSerialNumber = "BBBBBBBB"
			diskModel = "BBBMODEL"
		} else if diskDrive == "nvme0n1p2" {
			diskSerialNumber = "CCCCCCCC"
			diskModel = "CCCMODEL"
		} else if diskDrive == "nvme1n1p2" {
			diskSerialNumber = "DDDDDDDD"
			diskModel = "DDDMODEL"
		} else if diskDrive == "nvme0n1p5" {
			diskSerialNumber = "EEEEEEEE"
			diskModel = "EEEMODEL"
		} else if diskDrive == "nvme1n1p5" {
			diskSerialNumber = "FFFFFFFF"
			diskModel = "FFFMODEL"
		}
		return diskSerialNumber, diskModel, diskIntf, diskMedium, nil
	}

	// Mocked function
	utils.GetDiskPartitionSize = func(diskDrive string) (string, error) {
		if diskDrive == "md3" {
			return "100GB", nil
		} else if diskDrive == "md2" {
			return "200GB", nil
		} else if diskDrive == "md5" {
			return "600GB", nil
		} else if diskDrive == "nvme0n1p3" {
			return "100GB", nil
		} else if diskDrive == "nvme1n1p3" {
			return "100GB", nil
		} else if diskDrive == "nvme0n1p2" {
			return "200GB", nil
		} else if diskDrive == "nvme1n1p2" {
			return "200GB", nil
		} else if diskDrive == "nvme0n1p5" {
			return "300GB", nil
		} else if diskDrive == "nvme1n1p5" {
			return "300GB", nil
		} else {
			return "Unknown", fmt.Errorf("Error: TestProcessSoftRaid - GetDiskPartitionSize: Unknown diskDrive received: %v", diskDrive)
		}
	}

	newControllers, newRaids, err := ProcessSoftRaid("softraid")
	if err != nil {
		t.Fatalf(`TestProcessSoftRaid: error: %s`, err)
	}

	//fmt.Println("----------- newControllers ------------")
	//spew.Dump(newControllers)
	for _, newController := range newControllers {
		newControllerIdWanted := "softraid-0"
		if newController.Id != newControllerIdWanted {
			t.Fatalf(`TestProcessSoftRaid: newController.Id: %v must match %v`, newController.Id, newControllerIdWanted)
		}

		newControllerManufacturerWanted := "mdadm"
		if newController.Manufacturer != newControllerManufacturerWanted {
			t.Fatalf(`TestProcessSoftRaid: newController.Manufacturer: %v must match %v`, newController.Manufacturer, newControllerManufacturerWanted)
		}

		newControllerModelWanted := "MDADM"
		if newController.Model != newControllerModelWanted {
			t.Fatalf(`TestProcessSoftRaid: newController.Model: %v must match %v`, newController.Model, newControllerModelWanted)
		}

		newControllerStatusWanted := "Good"
		if newController.Status != newControllerStatusWanted {
			t.Fatalf(`TestProcessSoftRaid: newController.Status: %v must match %v`, newController.Status, newControllerStatusWanted)
		}
	}

	//fmt.Println("----------- newRaids ------------")
	//spew.Dump(newRaids)
	for i, newRaid := range newRaids {
		newRaidControllerIdWanted := "softraid-0"
		if newRaid.ControllerId != newRaidControllerIdWanted {
			t.Fatalf(`TestProcessSoftRaid: newRaid.ControllerId: %v must match %v`, newRaid.ControllerId, newRaidControllerIdWanted)
		}

		newRaidRaidLevelWanted := 0
		if newRaid.RaidLevel != newRaidRaidLevelWanted {
			t.Fatalf(`TestProcessSoftRaid: newRaid.RaidLevel: %v must match %v`, newRaid.RaidLevel, newRaidRaidLevelWanted)
		}

		newRaidDgWanted := ""
		if i == 0 {
			newRaidDgWanted = "md3"
		} else if i == 1 {
			newRaidDgWanted = "md2"
		} else {
			newRaidDgWanted = "md5"
		}
		if newRaid.Dg != newRaidDgWanted {
			t.Fatalf(`TestProcessSoftRaid: newRaid.Dg: %v must match %v`, newRaid.Dg, newRaidDgWanted)
		}

		newRaidRaidTypeWanted := ""
		if i == 0 {
			newRaidRaidTypeWanted = "raid1"
		} else if i == 1 {
			newRaidRaidTypeWanted = "raid1"
		} else {
			newRaidRaidTypeWanted = "raid0"
		}
		if newRaid.RaidType != newRaidRaidTypeWanted {
			t.Fatalf(`TestProcessSoftRaid: newRaid.RaidType: %v must match %v`, newRaid.RaidType, newRaidRaidTypeWanted)
		}

		newRaidStateWanted := "Okay"
		if newRaid.State != newRaidStateWanted {
			t.Fatalf(`TestProcessSoftRaid: newRaid.State: %v must match %v`, newRaid.State, newRaidStateWanted)
		}

		newRaidSizeWanted := ""
		if i == 0 {
			newRaidSizeWanted = "100GB"
		} else if i == 1 {
			newRaidSizeWanted = "200GB"
		} else {
			newRaidSizeWanted = "600GB"
		}
		if newRaid.Size != newRaidSizeWanted {
			t.Fatalf(`TestProcessSoftRaid: newRaid.Size: %v must match %v`, newRaid.Size, newRaidSizeWanted)
		}

		for j, disk := range newRaid.Disks {
			diskDgWanted := ""
			diskOsDeviceWanted := ""
			diskSizeWanted := ""
			diskModelWanted := ""
			diskSerialNumberWanted := ""
			if i == 0 && j == 0 {
				diskDgWanted = "md3"
				diskOsDeviceWanted = "nvme0n1p3"
				diskSizeWanted = "100GB"
				diskModelWanted = "AAAMODEL"
				diskSerialNumberWanted = "AAAAAAAA"
			} else if i == 0 && j == 1 {
				diskDgWanted = "md3"
				diskOsDeviceWanted = "nvme1n1p3"
				diskSizeWanted = "100GB"
				diskModelWanted = "BBBMODEL"
				diskSerialNumberWanted = "BBBBBBBB"
			} else if i == 1 && j == 0 {
				diskDgWanted = "md2"
				diskOsDeviceWanted = "nvme0n1p2"
				diskSizeWanted = "200GB"
				diskModelWanted = "CCCMODEL"
				diskSerialNumberWanted = "CCCCCCCC"
			} else if i == 1 && j == 1 {
				diskDgWanted = "md2"
				diskOsDeviceWanted = "nvme1n1p2"
				diskSizeWanted = "200GB"
				diskModelWanted = "DDDMODEL"
				diskSerialNumberWanted = "DDDDDDDD"
			} else if i == 2 && j == 0 {
				diskDgWanted = "md5"
				diskOsDeviceWanted = "nvme0n1p5"
				diskSizeWanted = "300GB"
				diskModelWanted = "EEEMODEL"
				diskSerialNumberWanted = "EEEEEEEE"
			} else {
				diskDgWanted = "md5"
				diskOsDeviceWanted = "nvme1n1p5"
				diskSizeWanted = "300GB"
				diskModelWanted = "FFFMODEL"
				diskSerialNumberWanted = "FFFFFFFF"
			}

			diskControllerIdWanted := "softraid-0"
			if disk.ControllerId != diskControllerIdWanted {
				t.Fatalf(`TestProcessSoftRaid: disk.ControllerId: %v must match %v`, disk.ControllerId, diskControllerIdWanted)
			}

			if disk.Dg != diskDgWanted {
				t.Fatalf(`TestProcessSoftRaid: disk.Dg: %v must match %v`, disk.Dg, diskDgWanted)
			}

			diskEidSlotWanted := ""
			if disk.EidSlot != diskEidSlotWanted {
				t.Fatalf(`TestProcessSoftRaid: disk.EidSlot: %v should be nil`, disk.EidSlot)
			}

			diskStateWanted := "ONLINE"
			if disk.State != diskStateWanted {
				t.Fatalf(`TestProcessSoftRaid: disk.State: %v must match %v`, disk.State, diskStateWanted)
			}

			if disk.Size != diskSizeWanted {
				t.Fatalf(`TestProcessSoftRaid: %v disk.Size: %v must match %v`, disk.OsDevice, disk.Size, diskSizeWanted)
			}

			diskIntfWanted := "SATA"
			if disk.Intf != diskIntfWanted {
				t.Fatalf(`TestProcessSoftRaid: disk.Intf: %v must match %v`, disk.Intf, diskIntfWanted)
			}

			diskMediumWanted := "SSD"
			if disk.Medium != diskMediumWanted {
				t.Fatalf(`TestProcessSoftRaid: disk.Medium: %v must match %v`, disk.Medium, diskMediumWanted)
			}

			if disk.Model != diskModelWanted {
				t.Fatalf(`TestProcessSoftRaid: disk.Model: %v must match %v`, disk.Model, diskModelWanted)
			}

			if disk.SerialNumber != diskSerialNumberWanted {
				t.Fatalf(`TestProcessSoftRaid: disk.SerialNumber: %v must match %v`, disk.SerialNumber, diskSerialNumberWanted)
			}

			if disk.OsDevice != diskOsDeviceWanted {
				t.Fatalf(`TestProcessSoftRaid: disk.OsDevice: %v must match %v`, disk.OsDevice, diskOsDeviceWanted)
			}
		}

		newRaidOsDeviceWanted := ""
		if i == 0 {
			newRaidOsDeviceWanted = "md3"
		} else if i == 1 {
			newRaidOsDeviceWanted = "md2"
		} else {
			newRaidOsDeviceWanted = "md5"
		}
		if newRaid.OsDevice != newRaidOsDeviceWanted {
			t.Fatalf(`TestProcessSoftRaid: newRaid.OsDevice: %v must match %v`, newRaid.OsDevice, newRaidOsDeviceWanted)
		}
	}
}

// Test ProcessSoftRaid no kernel support
func TestProcessSoftRaidNoKernelSupport(t *testing.T) {
	// Copy original functions content
	getSoftraidsOri := GetSoftraids
	// unmock functions content
	defer func() {
		GetSoftraids = getSoftraidsOri
	}()

	// Mocked function
	GetSoftraids = func() (*bufio.Scanner, *os.File, error) {
		//fmt.Println("-- Executing mocked GetSoftraids function")
		return nil, nil, fmt.Errorf("RANDOM ERROR")
	}

	_, _, err := ProcessSoftRaid("softraid")
	if err == nil {
		t.Fatalf(`TestProcessSoftRaidNoKernelSupport: nil error detected`)
	}
}

// Test CheckSoftRaidDisks
func TestCheckSoftRaidDisks(t *testing.T) {

	var raids []utils.RaidStruct
	raid := utils.RaidStruct{
		ControllerId: "softraid-0",
		RaidLevel:    0,
		Dg:           "md1",
		State:        "Okay",
		RaidType:     "raid1",
		OsDevice:     "md1",
	}

	disksArray := []string{"sdb3", "sdc3"}
	serailNumberArray := []string{"S3YVNB0KB13582W", "S3YVNB0KB13580D"}
	for i, diskArray := range disksArray {
		serialNumber := serailNumberArray[i]
		disk := utils.DiskStruct{
			ControllerId: "softraid-0",
			Dg:           "md1",
			State:        "ONLINE",
			Size:         "2.0TB",
			Intf:         "SATA",
			Medium:       "SSD",
			Model:        "SamsungSSD860",
			SerialNumber: serialNumber,
			OsDevice:     diskArray,
		}
		raid.AddDisk(disk)
	}
	raids = append(raids, raid)

	var newRaids []utils.RaidStruct
	newRaid := utils.RaidStruct{
		ControllerId: "btrfs-0",
		RaidLevel:    0,
		Dg:           "39c52b77-13b6-43ac-8e16-d77ba4800e39",
		State:        "ONLINE",
		RaidType:     "single",
		OsDevice:     "md1",
	}

	newDisk := utils.DiskStruct{
		ControllerId: "btrfs-0",
		Dg:           "39c52b77-13b6-43ac-8e16-d77ba4800e39",
		State:        "ONLINE",
		Size:         "1.77TiB",
		Intf:         "Unknown",
		Medium:       "Unknown",
		Model:        "Unknown",
		SerialNumber: "Unknown",
		OsDevice:     "md1",
	}
	newRaid.AddDisk(newDisk)
	newRaids = append(newRaids, newRaid)

	err := CheckSoftRaidDisks(newRaids, raids)
	if err != nil {
		t.Fatalf(`TestCheckSoftRaidDisks: error: %s`, err)
	}

	for _, raid = range raids {
		raidOsDeviceWAnted := "md1 Btrfs"
		if raid.OsDevice != raidOsDeviceWAnted {
			t.Fatalf(`TestCheckSoftRaidDisks raid.OsDevice: %v should be: %v`, raid.OsDevice, raidOsDeviceWAnted)
		}
	}
}

// Test CheckSoftRaidDisks No Match
// Test CheckSoftRaidDisks
func TestCheckSoftRaidDisksNoMatch(t *testing.T) {

	var raids []utils.RaidStruct
	raid := utils.RaidStruct{
		ControllerId: "softraid-0",
		RaidLevel:    0,
		Dg:           "md1",
		State:        "Okay",
		RaidType:     "raid1",
		OsDevice:     "md1",
	}

	disksArray := []string{"sdb3", "sdc3"}
	serialNumberArray := []string{"S3YVNB0KB13582W", "S3YVNB0KB13580D"}
	for i, diskArray := range disksArray {
		serialNumber := serialNumberArray[i]
		disk := utils.DiskStruct{
			ControllerId: "softraid-0",
			Dg:           "md1",
			State:        "ONLINE",
			Size:         "2.0TB",
			Intf:         "SATA",
			Medium:       "SSD",
			Model:        "SamsungSSD860",
			SerialNumber: serialNumber,
			OsDevice:     diskArray,
		}
		raid.AddDisk(disk)
	}
	raids = append(raids, raid)

	// We change OsDevice for not matching Raid device
	var newRaids []utils.RaidStruct
	newRaid := utils.RaidStruct{
		ControllerId: "btrfs-0",
		RaidLevel:    0,
		Dg:           "39c52b77-13b6-43ac-8e16-d77ba4800e39",
		State:        "ONLINE",
		RaidType:     "single",
		OsDevice:     "sdf",
	}

	newDisk := utils.DiskStruct{
		ControllerId: "btrfs-0",
		Dg:           "39c52b77-13b6-43ac-8e16-d77ba4800e39",
		State:        "ONLINE",
		Size:         "1.77TiB",
		Intf:         "Unknown",
		Medium:       "Unknown",
		Model:        "Unknown",
		SerialNumber: "Unknown",
		OsDevice:     "sdf",
	}
	newRaid.AddDisk(newDisk)
	newRaids = append(newRaids, newRaid)

	err := CheckSoftRaidDisks(newRaids, raids)
	if err != nil {
		t.Fatalf(`TestCheckSoftRaidDisksNoMatch: error: %s`, err)
	}

	for _, raid = range raids {
		raidOsDeviceWAnted := "md1"
		if raid.OsDevice != raidOsDeviceWAnted {
			t.Fatalf(`TestCheckSoftRaidDisksNoMatch raid.OsDevice: %v should be: %v`, raid.OsDevice, raidOsDeviceWAnted)
		}
	}
}
