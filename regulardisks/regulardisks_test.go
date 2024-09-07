package regulardisks

import (
	"fmt"
	"hardwareAnalyzer/utils"
	"testing"
)

// Test ProcessRegularDisks
func TestProcessRegularDisks(t *testing.T) {
	// Copy original functions content
	getSystemDisksOri := GetSystemDisks
	getDiskPartitionSizeOri := utils.GetDiskPartitionSize
	getDiskDataOri := utils.GetDiskData
	// unmock functions content
	defer func() {
		GetSystemDisks = getSystemDisksOri
		utils.GetDiskPartitionSize = getDiskPartitionSizeOri
		utils.GetDiskData = getDiskDataOri
	}()

	// Mocked function
	GetSystemDisks = func() ([]string, error) {
		//fmt.Println("-- Executing mocked GetSystemDisks")
		diskArray := []string{"sdf", "sdg", "sdh", "sdi"}
		return diskArray, nil
	}

	// Mocked function
	utils.GetDiskPartitionSize = func(diskDrive string) (string, error) {
		if diskDrive == "sdf" {
			return "100GB", nil
		} else if diskDrive == "sdg" {
			return "200GB", nil
		} else if diskDrive == "sdh" {
			return "300GB", nil
		} else if diskDrive == "sdi" {
			return "400GB", nil
		} else {
			return "Unknown", fmt.Errorf("Error: TestProcessRegularDisks - GetDiskPartitionSize: Unknown diskDrive received: %v", diskDrive)
		}
	}

	// Mocked function
	utils.GetDiskData = func(diskDrive string) (string, string, string, string, error) {
		diskIntf := "ata"
		diskMedium := "ata"
		diskSerialNumber := ""
		diskModel := ""
		if diskDrive == "sdf" {
			diskSerialNumber = "FFFFFFFF"
			diskModel = "SDFMODEL"
		} else if diskDrive == "sdg" {
			diskSerialNumber = "GGGGGGGG"
			diskModel = "SDGMODEL"
		} else if diskDrive == "sdh" {
			diskSerialNumber = "HHHHHHHH"
			diskModel = "SDHMODEL"
		} else if diskDrive == "sdi" {
			diskSerialNumber = "IIIIIIII"
			diskModel = "SDIMODEL"
		}
		return diskSerialNumber, diskModel, diskIntf, diskMedium, nil
	}

	var raids []utils.RaidStruct
	var noRaidDisks []utils.NoRaidDiskStruct
	newControllers, newRaids, err := ProcessRegularDisks(raids, noRaidDisks)

	if err != nil {
		t.Fatalf(`TestProcessRegularDisks: error: %v`, err)
	}

	//fmt.Println("----------------- newControllers ------------------")
	//spew.Dump(newControllers)

	for _, newController := range newControllers {
		newControllerIdWanted := "motherBoard-0"
		if newController.Id != newControllerIdWanted {
			t.Fatalf(`TestProcessRegularDisks: newController.Id: %v muts match %v`, newController.Id, newControllerIdWanted)
		}

		newControllerManufacturerWanted := "motherboard"
		if newController.Manufacturer != newControllerManufacturerWanted {
			t.Fatalf(`TestProcessRegularDisks: newController.Manufacturer: %v muts match %v`, newController.Manufacturer, newControllerManufacturerWanted)
		}

		newControllerModelWanted := "MOTHERBOARD"
		if newController.Model != newControllerModelWanted {
			t.Fatalf(`TestProcessRegularDisks: newController.Model: %v muts match %v`, newController.Model, newControllerModelWanted)
		}

		newControllerStatusWanted := "Good"
		if newController.Status != newControllerStatusWanted {
			t.Fatalf(`TestProcessRegularDisks: newController.Status: %v muts match %v`, newController.Status, newControllerStatusWanted)
		}
	}

	//fmt.Println("----------------- newRaids ------------------")
	//spew.Dump(newRaids)

	for _, newRaid := range newRaids {
		newRaidControllerIdWanted := "motherBoard-0"
		if newRaid.ControllerId != newRaidControllerIdWanted {
			t.Fatalf(`TestProcessRegularDisks: newController.ControllerId: %v muts match %v`, newRaid.ControllerId, newRaidControllerIdWanted)
		}

		newRaidRaidLevelWanted := 0
		if newRaid.RaidLevel != newRaidRaidLevelWanted {
			t.Fatalf(`TestProcessRegularDisks: newController.RaidLevel: %v muts match %v`, newRaid.RaidLevel, newRaidRaidLevelWanted)
		}

		newRaidDgWanted := ""
		if newRaid.Dg != newRaidDgWanted {
			t.Fatalf(`TestProcessRegularDisks: newController.Dg: %v muts be nil`, newRaid.Dg)
		}

		newRaidRaidTypeWanted := ""
		if newRaid.RaidType != newRaidRaidTypeWanted {
			t.Fatalf(`TestProcessRegularDisks: newController.RaidType: %v muts be nil`, newRaid.RaidType)
		}

		newRaidStateWanted := "Good"
		if newRaid.State != newRaidStateWanted {
			t.Fatalf(`TestProcessRegularDisks: newController.State: %v muts match %v`, newRaid.State, newRaidStateWanted)
		}

		newRaidSizeWanted := ""
		if newRaid.Size != newRaidSizeWanted {
			t.Fatalf(`TestProcessRegularDisks: newController.Size: %v muts be nil`, newRaid.Size)
		}

		newRaidDisksLenWanted := 4
		if len(newRaid.Disks) != newRaidDisksLenWanted {
			t.Fatalf(`TestProcessRegularDisks: len(newRaid.Disks): %v muts be %v`, len(newRaid.Disks), newRaidDisksLenWanted)
		}

		for i, disk := range newRaid.Disks {
			diskControllerIdWanted := "motherBoard-0"
			if disk.ControllerId != diskControllerIdWanted {
				t.Fatalf(`TestProcessRegularDisks: disk.ControllerId: %v muts match %v`, disk.ControllerId, diskControllerIdWanted)
			}

			diskDgWanted := ""
			if disk.Dg != diskDgWanted {
				t.Fatalf(`TestProcessRegularDisks: disk.Dg: %v muts be nil`, disk.Dg)
			}

			diskEidSlotWanted := ""
			if disk.EidSlot != diskEidSlotWanted {
				t.Fatalf(`TestProcessRegularDisks: disk.EidSlot: %v muts be nil`, disk.EidSlot)
			}

			diskStateWanted := "ONLINE"
			if disk.State != diskStateWanted {
				t.Fatalf(`TestProcessRegularDisks: disk.State: %v muts match %v`, disk.State, diskStateWanted)
			}

			diskSizeWanted := ""
			if i == 0 {
				diskSizeWanted = "100GB"
			} else if i == 1 {
				diskSizeWanted = "200GB"
			} else if i == 2 {
				diskSizeWanted = "300GB"
			} else {
				diskSizeWanted = "400GB"
			}
			if disk.Size != diskSizeWanted {
				t.Fatalf(`TestProcessRegularDisks: disk.Size: %v muts match %v`, disk.Size, diskSizeWanted)
			}

			diskIntfWanted := "ata"
			if disk.Intf != diskIntfWanted {
				t.Fatalf(`TestProcessRegularDisks: disk.Intf: %v muts match %v`, disk.Intf, diskIntfWanted)
			}

			diskMediumWanted := "ata"
			if disk.Medium != diskMediumWanted {
				t.Fatalf(`TestProcessRegularDisks: disk.Medium: %v muts match %v`, disk.Medium, diskMediumWanted)
			}

			diskModelWanted := ""
			if i == 0 {
				diskModelWanted = "SDFMODEL"
			} else if i == 1 {
				diskModelWanted = "SDGMODEL"
			} else if i == 2 {
				diskModelWanted = "SDHMODEL"
			} else {
				diskModelWanted = "SDIMODEL"
			}
			if disk.Model != diskModelWanted {
				t.Fatalf(`TestProcessRegularDisks: disk.Model: %v muts match %v`, disk.Model, diskModelWanted)
			}

			diskSerialNumberWanted := ""
			if i == 0 {
				diskSerialNumberWanted = "FFFFFFFF"
			} else if i == 1 {
				diskSerialNumberWanted = "GGGGGGGG"
			} else if i == 2 {
				diskSerialNumberWanted = "HHHHHHHH"
			} else {
				diskSerialNumberWanted = "IIIIIIII"
			}
			if disk.SerialNumber != diskSerialNumberWanted {
				t.Fatalf(`TestProcessRegularDisks: disk.SerialNumber: %v muts match %v`, disk.SerialNumber, diskSerialNumberWanted)
			}

			diskOsDeviceWanted := ""
			if i == 0 {
				diskOsDeviceWanted = "sdf"
			} else if i == 1 {
				diskOsDeviceWanted = "sdg"
			} else if i == 2 {
				diskOsDeviceWanted = "sdh"
			} else {
				diskOsDeviceWanted = "sdi"
			}
			if disk.OsDevice != diskOsDeviceWanted {
				t.Fatalf(`TestProcessRegularDisks: disk.OsDevice: %v muts match %v`, disk.OsDevice, diskOsDeviceWanted)
			}
		}

		newOsDeviceWanted := ""
		if newRaid.OsDevice != newOsDeviceWanted {
			t.Fatalf(`TestProcessRegularDisks: newController.OsDevice: %v muts be nil`, newRaid.OsDevice)
		}
	}
}

// Test ProcessRegularDisks, match with raid disk
func TestProcessRegularDisksMatchWithRaidDisk(t *testing.T) {
	// Copy original functions content
	getSystemDisksOri := GetSystemDisks
	getDiskPartitionSizeOri := utils.GetDiskPartitionSize
	getDiskDataOri := utils.GetDiskData
	// unmock functions content
	defer func() {
		GetSystemDisks = getSystemDisksOri
		utils.GetDiskPartitionSize = getDiskPartitionSizeOri
		utils.GetDiskData = getDiskDataOri
	}()

	// Mocked function
	GetSystemDisks = func() ([]string, error) {
		//fmt.Println("-- Executing mocked GetSystemDisks")
		diskArray := []string{"sdf", "sdg", "sdh", "sdi"}
		return diskArray, nil
	}

	// Mocked function
	utils.GetDiskPartitionSize = func(diskDrive string) (string, error) {
		if diskDrive == "sdf" {
			return "100GB", nil
		} else if diskDrive == "sdg" {
			return "200GB", nil
		} else if diskDrive == "sdh" {
			return "300GB", nil
		} else if diskDrive == "sdi" {
			return "400GB", nil
		} else {
			return "Unknown", fmt.Errorf("Error: TestProcessRegularDisksMatchWithRaidDisk - GetDiskPartitionSize: Unknown diskDrive received: %v", diskDrive)
		}
	}

	// Mocked function
	utils.GetDiskData = func(diskDrive string) (string, string, string, string, error) {
		diskIntf := "ata"
		diskMedium := "ata"
		diskSerialNumber := ""
		diskModel := ""
		if diskDrive == "sdf" {
			diskSerialNumber = "FFFFFFFF"
			diskModel = "SDFMODEL"
		} else if diskDrive == "sdg" {
			diskSerialNumber = "GGGGGGGG"
			diskModel = "SDGMODEL"
		} else if diskDrive == "sdh" {
			diskSerialNumber = "HHHHHHHH"
			diskModel = "SDHMODEL"
		} else if diskDrive == "sdi" {
			diskSerialNumber = "IIIIIIII"
			diskModel = "SDIMODEL"
		}
		return diskSerialNumber, diskModel, diskIntf, diskMedium, nil
	}

	// Raid has one of the motherboard disk attached: sdi
	var raids []utils.RaidStruct
	raid := utils.RaidStruct{
		ControllerId: "motherBoard-0",
		State:        "Good",
		OsDevice:     "sdi",
	}
	raids = append(raids, raid)
	//fmt.Println("-------- raids1 ---------")
	//spew.Dump(raids)

	var noRaidDisks []utils.NoRaidDiskStruct
	newControllers, newRaids, err := ProcessRegularDisks(raids, noRaidDisks)

	if err != nil {
		t.Fatalf(`TestProcessRegularDisks: error: %v`, err)
	}

	//fmt.Println("----------------- newControllers ------------------")
	//spew.Dump(newControllers)

	for _, newController := range newControllers {
		newControllerIdWanted := "motherBoard-0"
		if newController.Id != newControllerIdWanted {
			t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: newController.Id: %v muts match %v`, newController.Id, newControllerIdWanted)
		}

		newControllerManufacturerWanted := "motherboard"
		if newController.Manufacturer != newControllerManufacturerWanted {
			t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: newController.Manufacturer: %v muts match %v`, newController.Manufacturer, newControllerManufacturerWanted)
		}

		newControllerModelWanted := "MOTHERBOARD"
		if newController.Model != newControllerModelWanted {
			t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: newController.Model: %v muts match %v`, newController.Model, newControllerModelWanted)
		}

		newControllerStatusWanted := "Good"
		if newController.Status != newControllerStatusWanted {
			t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: newController.Status: %v muts match %v`, newController.Status, newControllerStatusWanted)
		}
	}

	//fmt.Println("----------------- newRaids ------------------")
	//spew.Dump(newRaids)

	for _, newRaid := range newRaids {
		newRaidControllerIdWanted := "motherBoard-0"
		if newRaid.ControllerId != newRaidControllerIdWanted {
			t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: newController.ControllerId: %v muts match %v`, newRaid.ControllerId, newRaidControllerIdWanted)
		}

		newRaidRaidLevelWanted := 0
		if newRaid.RaidLevel != newRaidRaidLevelWanted {
			t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: newController.RaidLevel: %v muts match %v`, newRaid.RaidLevel, newRaidRaidLevelWanted)
		}

		newRaidDgWanted := ""
		if newRaid.Dg != newRaidDgWanted {
			t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: newController.Dg: %v muts be nil`, newRaid.Dg)
		}

		newRaidRaidTypeWanted := ""
		if newRaid.RaidType != newRaidRaidTypeWanted {
			t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: newController.RaidType: %v muts be nil`, newRaid.RaidType)
		}

		newRaidStateWanted := "Good"
		if newRaid.State != newRaidStateWanted {
			t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: newController.State: %v muts match %v`, newRaid.State, newRaidStateWanted)
		}

		newRaidSizeWanted := ""
		if newRaid.Size != newRaidSizeWanted {
			t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: newController.Size: %v muts be nil`, newRaid.Size)
		}

		newRaidDisksLenWanted := 3
		if len(newRaid.Disks) != newRaidDisksLenWanted {
			t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: len(newRaid.Disks): %v muts be %v`, len(newRaid.Disks), newRaidDisksLenWanted)
		}

		for i, disk := range newRaid.Disks {
			diskControllerIdWanted := "motherBoard-0"
			if disk.ControllerId != diskControllerIdWanted {
				t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: disk.ControllerId: %v muts match %v`, disk.ControllerId, diskControllerIdWanted)
			}

			diskDgWanted := ""
			if disk.Dg != diskDgWanted {
				t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: disk.Dg: %v muts be nil`, disk.Dg)
			}

			diskEidSlotWanted := ""
			if disk.EidSlot != diskEidSlotWanted {
				t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: disk.EidSlot: %v muts be nil`, disk.EidSlot)
			}

			diskStateWanted := "ONLINE"
			if disk.State != diskStateWanted {
				t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: disk.State: %v muts match %v`, disk.State, diskStateWanted)
			}

			diskSizeWanted := ""
			if i == 0 {
				diskSizeWanted = "100GB"
			} else if i == 1 {
				diskSizeWanted = "200GB"
			} else {
				diskSizeWanted = "300GB"
			}
			if disk.Size != diskSizeWanted {
				t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: disk.Size: %v muts match %v`, disk.Size, diskSizeWanted)
			}

			diskIntfWanted := "ata"
			if disk.Intf != diskIntfWanted {
				t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: disk.Intf: %v muts match %v`, disk.Intf, diskIntfWanted)
			}

			diskMediumWanted := "ata"
			if disk.Medium != diskMediumWanted {
				t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: disk.Medium: %v muts match %v`, disk.Medium, diskMediumWanted)
			}

			diskModelWanted := ""
			if i == 0 {
				diskModelWanted = "SDFMODEL"
			} else if i == 1 {
				diskModelWanted = "SDGMODEL"
			} else {
				diskModelWanted = "SDHMODEL"
			}
			if disk.Model != diskModelWanted {
				t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: disk.Model: %v muts match %v`, disk.Model, diskModelWanted)
			}

			diskSerialNumberWanted := ""
			if i == 0 {
				diskSerialNumberWanted = "FFFFFFFF"
			} else if i == 1 {
				diskSerialNumberWanted = "GGGGGGGG"
			} else {
				diskSerialNumberWanted = "HHHHHHHH"
			}
			if disk.SerialNumber != diskSerialNumberWanted {
				t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: disk.SerialNumber: %v muts match %v`, disk.SerialNumber, diskSerialNumberWanted)
			}

			diskOsDeviceWanted := ""
			if i == 0 {
				diskOsDeviceWanted = "sdf"
			} else if i == 1 {
				diskOsDeviceWanted = "sdg"
			} else {
				diskOsDeviceWanted = "sdh"
			}
			if disk.OsDevice != diskOsDeviceWanted {
				t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: disk.OsDevice: %v muts match %v`, disk.OsDevice, diskOsDeviceWanted)
			}
		}

		newOsDeviceWanted := ""
		if newRaid.OsDevice != newOsDeviceWanted {
			t.Fatalf(`TestProcessRegularDisksMatchWithRaidDisk: newController.OsDevice: %v muts be nil`, newRaid.OsDevice)
		}
	}
}

// Test ProcessRegularDisks, no system disks
func TestProcessRegularDisksNoSystemDisks(t *testing.T) {
	// Copy original functions content
	getSystemDisksOri := GetSystemDisks
	// unmock functions content
	defer func() {
		GetSystemDisks = getSystemDisksOri
	}()

	// Mocked function
	GetSystemDisks = func() ([]string, error) {
		//fmt.Println("-- Executing mocked GetSystemDisks")
		var diskArray []string
		return diskArray, fmt.Errorf("Error: GetSystemDisks, diskArray empty len(diskArray): %v.", len(diskArray))
	}

	var raids []utils.RaidStruct
	var noRaidDisks []utils.NoRaidDiskStruct
	_, _, err := ProcessRegularDisks(raids, noRaidDisks)
	if err == nil {
		t.Fatalf(`TestProcessRegularDisksNoSystemDisks: err should be != nil`)
	}
}

// Test GetSystemDisks
func TestGetSystemDisks(t *testing.T) {
	diskArray, err := GetSystemDisks()
	if err != nil {
		t.Fatalf(`TestGetSystemDisks: err: %v`, err)
	}

	if len(diskArray) > 100 {
		t.Fatalf(`TestGetSystemDisks: Abnormal disks detected on system, disks found: %v`, len(diskArray))
	}
}
