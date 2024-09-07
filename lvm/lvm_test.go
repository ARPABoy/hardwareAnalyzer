package lvm

import (
	"bytes"
	"fmt"
	"hardwareAnalyzer/utils"
	"testing"
)

// Test CheckLVMRaid
func TestCheckLVMRaid(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function: TestCheckLVMRaid")
		var outputStdout, outputStderr bytes.Buffer
		// lvs only returns output when some LVM is configured
		outputStdout.WriteString("RANDOM OUTPUT")
		return &outputStdout, &outputStderr, nil
	}

	lvmRaidCheck, err := CheckLVMRaid()
	if err != nil {
		t.Fatalf(`TestCheckLVMRaid: error: %s`, err)
	}
	if !lvmRaidCheck {
		t.Fatalf(`TestCheckLVMRaid: %v, want TRUE`, lvmRaidCheck)
	}
}

// Test CheckLVMRaid No raid
func TestCheckLVMRaidNoRaid(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function: TestCheckLVMRaid")
		var outputStdout, outputStderr bytes.Buffer
		// lvs only returns output when some LVM is configured
		outputStdout.WriteString("")
		return &outputStdout, &outputStderr, nil
	}

	lvmRaidCheck, err := CheckLVMRaid()
	if err != nil {
		t.Fatalf(`TestCheckLVMRaidNoRaid: error: %s`, err)
	}
	if lvmRaidCheck {
		t.Fatalf(`TestCheckLVMRaidNoRaid: %v, want FALSE`, lvmRaidCheck)
	}
}

// Test CheckLVMRaid Error
func TestCheckLVMRaidError(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function: TestCheckLVMRaid")
		var outputStdout, outputStderr bytes.Buffer
		// Write error to stderr:
		outputStderr.WriteString("RANDOM ERROR")
		return &outputStdout, &outputStderr, nil
	}

	lvmRaidCheck, err := CheckLVMRaid()
	if err == nil {
		t.Fatalf(`TestCheckLVMRaidError: error should be != nil`)
	}
	if lvmRaidCheck {
		t.Fatalf(`TestCheckLVMRaidError: %v, want FALSE`, lvmRaidCheck)
	}
}

// Test ProcessLVMRaid
func TestProcessLVMRaid(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function: TestProcessLVMRaid")
		var outputStdout, outputStderr bytes.Buffer

		switch command {
		case "vgs --noheadings --units b -o vg_name,vg_size,vg_missing_pv_count":
			outputStdout.WriteString("  test-vg 478482006016B           0")
		case "vgs --noheadings --units b -o lv_size,segtype,vg_name,lv_path":
			outputStdout.WriteString(`
					53687091200B linear test-vg /dev/test-vg/root-lv
  					214748364800B linear test-vg /dev/test-vg/lv-0
  					210046550016B linear test-vg /dev/test-vg/lv-1
				`)
		case "pvs --noheadings --units b -o pv_name,vg_name,pv_size":
			outputStdout.WriteString(`
					/dev/sda3  test-vg 478482006016B
				`)
		default:
			return &outputStdout, &outputStderr, fmt.Errorf("Unknown command: %v.", command)
		}

		return &outputStdout, &outputStderr, nil
	}

	newControllers, newVolumeGroups, newRaids, err := ProcessLVMRaid("lvm")
	if err != nil {
		t.Fatalf(`TestProcessLVMRaid Error: %v`, err)
	}

	//fmt.Println("------------ newControllers ------------")
	//spew.Dump(newControllers)
	for _, newController := range newControllers {
		newControllerIdWanted := "lvm-0"
		if newController.Id != newControllerIdWanted {
			t.Fatalf(`TestProcessLVMRaid: newController.Id: %v muts match %v`, newController.Id, newControllerIdWanted)
		}

		newControllerManufacturerWanted := "lvm"
		if newController.Manufacturer != newControllerManufacturerWanted {
			t.Fatalf(`TestProcessLVMRaid: newController.Manufacturer: %v muts match %v`, newController.Manufacturer, newControllerManufacturerWanted)
		}

		newControllerModelWanted := "LVM"
		if newController.Model != newControllerModelWanted {
			t.Fatalf(`TestProcessLVMRaid: newController.Model: %v muts match %v`, newController.Model, newControllerModelWanted)
		}

		newControllerStatusWanted := "Good"
		if newController.Status != newControllerStatusWanted {
			t.Fatalf(`TestProcessLVMRaid: newController.Status: %v muts match %v`, newController.Status, newControllerStatusWanted)
		}
	}

	//fmt.Println("------------ newVolumeGroups ------------")
	//spew.Dump(newVolumeGroups)
	for _, newVolumeGroup := range newVolumeGroups {
		newVolumeGroupControllerIdWanted := "lvm-0"
		if newVolumeGroup.ControllerId != newVolumeGroupControllerIdWanted {
			t.Fatalf(`TestProcessLVMRaid: newVolumeGroup.ControllerId: %v muts match %v`, newVolumeGroup.ControllerId, newVolumeGroupControllerIdWanted)
		}

		newVolumeGroupNameWanted := "test-vg"
		if newVolumeGroup.Name != newVolumeGroupNameWanted {
			t.Fatalf(`TestProcessLVMRaid: newVolumeGroup.Name: %v muts match %v`, newVolumeGroup.Name, newVolumeGroupNameWanted)
		}

		newVolumeGroupStateWanted := "ONLINE"
		if newVolumeGroup.State != newVolumeGroupStateWanted {
			t.Fatalf(`TestProcessLVMRaid: newVolumeGroup.State: %v muts match %v`, newVolumeGroup.State, newVolumeGroupStateWanted)
		}

		newVolumeGroupSizeWanted := "478 GB"
		if newVolumeGroup.Size != newVolumeGroupSizeWanted {
			t.Fatalf(`TestProcessLVMRaid: newVolumeGroup.Size: %v muts match %v`, newVolumeGroup.Size, newVolumeGroupSizeWanted)
		}

	}

	//fmt.Println("------------ newRaids ------------")
	//spew.Dump(newRaids)
	for i, newRaid := range newRaids {
		newRaidControllerIdWanted := "lvm-0"
		if newRaid.ControllerId != newRaidControllerIdWanted {
			t.Fatalf(`TestProcessLVMRaid: newRaid.ControllerId: %v muts match %v`, newRaid.ControllerId, newRaidControllerIdWanted)
		}

		newRaidRaidLevelWanted := 0
		if newRaid.RaidLevel != newRaidRaidLevelWanted {
			t.Fatalf(`TestProcessLVMRaid: newRaid.RaidLevel: %v muts match %v`, newRaid.RaidLevel, newRaidRaidLevelWanted)
		}

		newRaidDgWanted := "test-vg"
		if newRaid.Dg != newRaidDgWanted {
			t.Fatalf(`TestProcessLVMRaid: newRaid.Dg: %v muts match %v`, newRaid.Dg, newRaidDgWanted)
		}

		newRaidRaidTypeWanted := "linear"
		if newRaid.RaidType != newRaidRaidTypeWanted {
			t.Fatalf(`TestProcessLVMRaid: newRaid.RaidType: %v muts match %v`, newRaid.RaidType, newRaidRaidTypeWanted)
		}

		newRaidStateWanted := "ONLINE"
		if newRaid.State != newRaidStateWanted {
			t.Fatalf(`TestProcessLVMRaid: newRaid.State: %v muts match %v`, newRaid.State, newRaidStateWanted)
		}

		newRaidSizeWanted := ""
		if i == 0 {
			newRaidSizeWanted = "54 GB"
		} else if i == 1 {
			newRaidSizeWanted = "215 GB"
		} else if i == 2 {
			newRaidSizeWanted = "210 GB"
		}
		if newRaid.Size != newRaidSizeWanted {
			t.Fatalf(`TestProcessLVMRaid: newRaid.Size: %v muts match %v`, newRaid.Size, newRaidSizeWanted)
		}

		for _, disk := range newRaid.Disks {
			diskControllerIdWanted := "lvm-0"
			if disk.ControllerId != diskControllerIdWanted {
				t.Fatalf(`TestProcessLVMRaid: disk.ControllerId: %v muts match %v`, disk.ControllerId, diskControllerIdWanted)
			}

			diskDgWanted := "test-vg"
			if disk.Dg != diskDgWanted {
				t.Fatalf(`TestProcessLVMRaid: disk.Dg: %v muts match %v`, disk.Dg, diskDgWanted)
			}

			diskEidSlotWanted := ""
			if disk.EidSlot != diskEidSlotWanted {
				t.Fatalf(`TestProcessLVMRaid: disk.EidSlot: %v muts match nil`, disk.EidSlot)
			}

			diskStateWanted := "ONLINE"
			if disk.State != diskStateWanted {
				t.Fatalf(`TestProcessLVMRaid: disk.State: %v muts match %v`, disk.State, diskStateWanted)
			}

			diskSizeWanted := "478 GB"
			if disk.Size != diskSizeWanted {
				t.Fatalf(`TestProcessLVMRaid: disk.Size: %v muts match %v`, disk.Size, diskSizeWanted)
			}

			diskIntfWanted := "Unknown"
			if disk.Intf != diskIntfWanted {
				t.Fatalf(`TestProcessLVMRaid: disk.Intf: %v muts match %v`, disk.Intf, diskIntfWanted)
			}

			diskMediumWanted := "Unknown"
			if disk.Medium != diskMediumWanted {
				t.Fatalf(`TestProcessLVMRaid: disk.Medium: %v muts match %v`, disk.Medium, diskMediumWanted)
			}

			diskModelWanted := "Unknown"
			if disk.Model != diskModelWanted {
				t.Fatalf(`TestProcessLVMRaid: disk.Model: %v muts match %v`, disk.Model, diskModelWanted)
			}

			diskSerialNumberWanted := "3600605b000a042e022f56cdf086a1173"
			if disk.SerialNumber != diskSerialNumberWanted {
				t.Fatalf(`TestProcessLVMRaid: disk.SerialNumber: %v muts match %v`, disk.SerialNumber, diskSerialNumberWanted)
			}

			diskOsDeviceWanted := "sda3"
			if disk.OsDevice != diskOsDeviceWanted {
				t.Fatalf(`TestProcessLVMRaid: disk.OsDevice: %v muts match %v`, disk.OsDevice, diskOsDeviceWanted)
			}
		}

		newRaidOsDeviceWanted := ""
		if i == 0 {
			newRaidOsDeviceWanted = "test-vg/root-lv"
		} else if i == 1 {
			newRaidOsDeviceWanted = "test-vg/lv-0"
		} else if i == 2 {
			newRaidOsDeviceWanted = "test-vg/lv-1"
		}
		if newRaid.OsDevice != newRaidOsDeviceWanted {
			t.Fatalf(`TestProcessLVMRaid: newRaid.OsDevice: %v muts match %v`, newRaid.OsDevice, newRaidOsDeviceWanted)
		}
	}
}

// Test ProcessLVMRaid GetCommandOutput error
func TestProcessLVMRaidGetCommandOutputError(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function: TestProcessLVMRaid")
		var outputStdout, outputStderr bytes.Buffer
		outputStderr.WriteString("RANDOM ERROR")
		return &outputStdout, &outputStderr, nil
	}

	_, _, _, err := ProcessLVMRaid("lvm")
	if err == nil {
		t.Fatalf(`TestProcessLVMRaidGetCommandOutputError err should be != nil`)
	}
}
