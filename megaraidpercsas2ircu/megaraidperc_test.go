package megaraidpercsas2ircu

import (
	"bytes"
	"hardwareAnalyzer/hardwarecontrollerscommon"
	"hardwareAnalyzer/utils"
	"strconv"
	"testing"
)

// Test GetMegaraidPercDriveSerialNumber
func TestGetMegaraidPercDriveSerialNumber(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	testSerialNumber := "TESTSERIALNUMBER"

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestGetMegaraidPercDriveSerialNumber function")
		var outputStdout, outputStderr bytes.Buffer
		outputStdout.WriteString("SN = " + testSerialNumber)
		return &outputStdout, &outputStderr, nil
	}

	manufacturer := "XX"
	controllerId := "XX"
	eidslot := "XX:YY"
	serialNumber, err := GetMegaraidPercDriveSerialNumber(manufacturer, controllerId, eidslot)

	if err != nil {
		t.Fatalf(`TestGetMegaraidPercDriveSerialNumber, getMegaraidPercDriveSerialNumber returned error: %s`, err)
	}

	if serialNumber != testSerialNumber {
		t.Fatalf(`TestGetMegaraidPercDriveSerialNumber serialNumber: %s should match testSerialNumber: %s`, serialNumber, testSerialNumber)
	}
}

// Test GetMegaraidPercDriveSerialNumber outputStderr
func TestGetMegaraidPercDriveSerialNumberOutputStderr(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestGetMegaraidPercDriveSerialNumberOutputStderr function")
		var outputStdout, outputStderr bytes.Buffer
		outputStderr.WriteString("RANDOM ERROR")
		return &outputStdout, &outputStderr, nil
	}

	manufacturer := "XX"
	controllerId := "XX"
	eidslot := "XX:YY"
	serialNumber, err := GetMegaraidPercDriveSerialNumber(manufacturer, controllerId, eidslot)

	if err == nil {
		t.Fatalf(`TestGetMegaraidPercDriveSerialNumberOutputStderr returned nil error`)
	}

	if serialNumber != "Unknown" {
		t.Fatalf(`TestGetMegaraidPercDriveSerialNumberOutputStderr serialNumber: %s should match Unknown`, serialNumber)
	}
}

// Test CheckMegaraidPerc
func TestCheckMegaraidPerc(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestCheckMegaraidPerc function")
		var outputStdout, outputStderr bytes.Buffer
		outputStdout.WriteString(`
			Status = Success
			Description = None
		`)
		return &outputStdout, &outputStderr, nil
	}

	megaRaidCheck, err := CheckMegaraidPerc("mega")
	if err != nil {
		t.Fatalf(`TestCheckMegaraidPerc returned error: %s`, err)
	}

	if !megaRaidCheck {
		t.Fatalf(`TestCheckMegaraidPerc megaRaidCheck: %v should be TRUE`, megaRaidCheck)
	}
}

// Test CheckMegaraidPerc outputStderr
func TestCheckMegaraidPercOutputStderr(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestCheckMegaraidPercOutputStderr function")
		var outputStdout, outputStderr bytes.Buffer
		outputStderr.WriteString("RANDOM ERROR")
		return &outputStdout, &outputStderr, nil
	}

	megaRaidCheck, err := CheckMegaraidPerc("mega")

	if err == nil {
		t.Fatalf(`TestCheckMegaraidPercOutputStderr returned nil error`)
	}

	if megaRaidCheck {
		t.Fatalf(`TestCheckMegaraidPercOutputStderr megaRaidCheck: %v should be FALSE`, megaRaidCheck)
	}
}

// Test CheckMegaraidPerc No controller
func TestCheckMegaraidPercNoController(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestCheckMegaraidPercNoController function")
		var outputStdout, outputStderr bytes.Buffer
		outputStdout.WriteString(`
			Status = RANDOM TEXT
			Description = No Controller found
		`)
		return &outputStdout, &outputStderr, nil
	}

	megaRaidCheck, err := CheckMegaraidPerc("mega")
	if err != nil {
		t.Fatalf(`TestCheckMegaraidPercNoController returned error: %s`, err)
	}

	if megaRaidCheck {
		t.Fatalf(`TestCheckMegaraidPercNoController megaRaidCheck: %v should be FALSE`, megaRaidCheck)
	}
}

// Test CheckMegaraidPerc invalid command output
func TestCheckMegaraidPercInvalidCommandOutput(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestCheckMegaraidPercInvalidCommandOutput function")
		var outputStdout, outputStderr bytes.Buffer
		outputStdout.WriteString(`
			XXXX = RANDOM TEXT
		`)
		return &outputStdout, &outputStderr, nil
	}

	megaRaidCheck, err := CheckMegaraidPerc("mega")
	if err != nil {
		t.Fatalf(`TestCheckMegaraidPercInvalidCommandOutput returned error: %s`, err)
	}

	if megaRaidCheck {
		t.Fatalf(`TestCheckMegaraidPercInvalidCommandOutput megaRaidCheck: %v should be FALSE`, megaRaidCheck)
	}
}

// Test ProcessHWMegaraidPercRaid
func TestProcessHWMegaraidPercRaid(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	getMegaraidPercDriveSerialNumberOri := GetMegaraidPercDriveSerialNumber
	getJbodOsDeviceOri := hardwarecontrollerscommon.GetJbodOsDevice
	getRaidOSDeviceOri := hardwarecontrollerscommon.GetRaidOSDevice // unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
		GetMegaraidPercDriveSerialNumber = getMegaraidPercDriveSerialNumberOri
		hardwarecontrollerscommon.GetJbodOsDevice = getJbodOsDeviceOri
		hardwarecontrollerscommon.GetRaidOSDevice = getRaidOSDeviceOri
	}()

	diskCounter := 1

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestProcessHWMegaraidPercRaid function")
		var outputStdout, outputStderr bytes.Buffer
		// storcli /call show all
		// No CacheCade disks in mocked output because I dont have any of it in my servers
		outputStdout.WriteString(`
			Controller = 0
			Model = LSI MegaRAID AlfaExploit Model
			Controller Status = Optimal
			TOPOLOGY :
			------------------------------------------------------------------------------
			DG Arr Row EID:Slot DID Type   State BT       Size PDC  PI SED DS3  FSpace TR
			------------------------------------------------------------------------------
			0 -   -   -        -   RAID10 Optl  N    1.454 TB enbl N  N   dflt N      N
			0 0   -   -        -   RAID1  Optl  N  744.687 GB enbl N  N   dflt N      N
			0 0   0   252:0    5   DRIVE  Onln  N  744.687 GB enbl N  N   dflt -      N
			0 0   1   252:1    7   DRIVE  Onln  N  893.750 GB enbl N  N   dflt -      N
			0 1   -   -        -   RAID1  Optl  N  744.687 GB enbl N  N   dflt N      N
			0 1   0   252:2    6   DRIVE  Onln  N  744.687 GB enbl N  N   dflt -      N
			0 1   1   252:3    4   DRIVE  Onln  N  744.687 GB enbl N  N   dflt -      N
			------------------------------------------------------------------------------

			VD LIST :
			--------------------------------------------------------------
			DG/VD TYPE   State Access Consist Cache Cac sCC     Size Name
			--------------------------------------------------------------
			0/0   RAID10 Optl  RW     Yes     RWTD  -   ON  1.454 TB
			--------------------------------------------------------------

			PD LIST :
			---------------------------------------------------------------------------------
			EID:Slt DID State DG       Size Intf Med SED PI SeSz Model               Sp Type
			---------------------------------------------------------------------------------
			252:0     5 Onln   0 744.687 GB SATA SSD N   N  512B INTEL SSDSC2BB800H4 U  -
			252:1     7 Onln   0 893.750 GB SATA SSD N   N  512B INTEL SSDSC2KB960G8 U  -
			252:2     6 Onln   0 744.687 GB SATA SSD N   N  512B INTEL SSDSC2BB800H4 U  -
			252:3     4 Onln   0 744.687 GB SATA SSD N   N  512B INTEL SSDSC2BB800H4 U  -
			252:4    21 UGood  - 223.062 GB SATA SSD N   N  512B INTEL SSDSC2KB240G8 U  -
			---------------------------------------------------------------------------------
		`)
		return &outputStdout, &outputStderr, nil
	}

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	GetMegaraidPercDriveSerialNumber = func(manufacturer, controllerId, eidSlot string) (string, error) {
		//fmt.Println("-- Executing mocked TestProcessHWMegaraidPercRaid function")
		testSerialNumber := "TESTSERIALNUMBER" + strconv.Itoa(diskCounter)
		diskCounter++
		return testSerialNumber, nil
	}

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	hardwarecontrollerscommon.GetJbodOsDevice = func(manufacturer, controllerId, eidslot string) (string, error) {
		//fmt.Println("-- Executing mocked TestProcessHWMegaraidPercRaid function")
		testOsDevice := "BogusDisk-OSUnknown"
		return testOsDevice, nil
	}

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	hardwarecontrollerscommon.GetRaidOSDevice = func(manufacturer, controllerId, dg string) (string, error) {
		testOsDevice := "TESTOSRAIDDEVICE"
		return testOsDevice, nil
	}

	newControllers, newRaids, newNoRaidDisks, err := ProcessHWMegaraidPercRaid("mega")
	if err != nil {
		t.Fatalf(`TestProcessHWMegaraidPercRaid returned error: %s`, err)
	}

	//fmt.Println("---------------- newControllers --------------------")
	//spew.Dump(newControllers)
	for _, newController := range newControllers {
		newControllerIdWanted := "mega-0"
		if newController.Id != newControllerIdWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newController.Id: %v should be: %v`, newController.Id, newControllerIdWanted)
		}

		newControllerManufacturerWanted := "mega"
		if newController.Manufacturer != newControllerManufacturerWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newController.Manufacturer: %v should be: %v`, newController.Manufacturer, newControllerManufacturerWanted)
		}

		newControllerModelWanted := "LSI MegaRAID AlfaExploit Model"
		if newController.Model != newControllerModelWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newController.Model: %v should be: %v`, newController.Model, newControllerModelWanted)
		}

		newControllerStatusWanted := "Optimal"
		if newController.Status != newControllerStatusWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newController.Status: %v should be: %v`, newController.Status, newControllerStatusWanted)
		}
	}

	//fmt.Println("---------------- newRaids --------------------")
	//spew.Dump(newRaids)
	for i, newRaid := range newRaids {
		newRaidControllerIdWanted := "mega-0"
		if newRaid.ControllerId != newRaidControllerIdWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newRaid.ControllerId: %v should be: %v`, newRaid.ControllerId, newRaidControllerIdWanted)
		}

		var newRaidControllerRaidLevelWanted int
		if i == 0 {
			newRaidControllerRaidLevelWanted = 0
		} else if i == 1 || i == 2 {
			newRaidControllerRaidLevelWanted = 1
		}
		if newRaid.RaidLevel != newRaidControllerRaidLevelWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newRaid.RaidLevel: %v should be: %v`, newRaid.RaidLevel, newRaidControllerRaidLevelWanted)
		}

		newRaidControllerDgWanted := "0"
		if newRaid.Dg != newRaidControllerDgWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newRaid.Dg: %v should be: %v`, newRaid.Dg, newRaidControllerDgWanted)
		}

		newRaidControllerRaidTypeWanted := ""
		if i == 0 {
			newRaidControllerRaidTypeWanted = "RAID10"
		} else if i == 1 || i == 2 {
			newRaidControllerRaidTypeWanted = "RAID1"
		}

		if newRaid.RaidType != newRaidControllerRaidTypeWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newRaid.RaidType: %v should be: %v`, newRaid.RaidType, newRaidControllerRaidTypeWanted)
		}

		newRaidControllerStateWanted := "Optl"
		if newRaid.State != newRaidControllerStateWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newRaid.State: %v should be: %v`, newRaid.State, newRaidControllerStateWanted)
		}

		newRaidControllerSizeWanted := ""
		if i == 0 {
			newRaidControllerSizeWanted = "1.454 TB"
		} else if i == 1 || i == 2 {
			newRaidControllerSizeWanted = "744.687 GB"
		}

		if newRaid.Size != newRaidControllerSizeWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newRaid.Size: %v should be: %v`, newRaid.Size, newRaidControllerSizeWanted)
		}

		// Extra check for sub-raid disks RAID10(RAID1/RAID1)
		if i != 0 {
			for j, disk := range newRaid.Disks {
				diskControllerIDWanted := "mega-0"
				if disk.ControllerId != diskControllerIDWanted {
					t.Fatalf(`TestProcessHWMegaraidPercRaid disk.ControllerID: %v should be: %v`, disk.ControllerId, diskControllerIDWanted)
				}

				diskDgWanted := "0"
				if disk.Dg != diskDgWanted {
					t.Fatalf(`TestProcessHWMegaraidPercRaid disk.Dg: %v should be: %v`, disk.Dg, diskDgWanted)
				}

				diskEidSlotWanted := ""
				if i == 1 && j == 0 {
					diskEidSlotWanted = "252:0"
				} else if i == 1 && j == 1 {
					diskEidSlotWanted = "252:1"
				} else if i == 2 && j == 0 {
					diskEidSlotWanted = "252:2"
				} else if i == 2 && j == 1 {
					diskEidSlotWanted = "252:3"
				}

				if disk.EidSlot != diskEidSlotWanted {
					t.Fatalf(`TestProcessHWMegaraidPercRaid disk.EidSlot: %v should be: %v`, disk.EidSlot, diskEidSlotWanted)
				}

				diskStateWanted := "Onln"
				if disk.State != diskStateWanted {
					t.Fatalf(`TestProcessHWMegaraidPercRaid disk.State: %v should be: %v`, disk.State, diskStateWanted)
				}

				diskSizeWanted := ""
				if i == 1 && j == 0 {
					diskSizeWanted = "744.687 GB"
				} else if i == 1 && j == 1 {
					diskSizeWanted = "893.750 GB"
				} else if i == 2 && j == 0 {
					diskSizeWanted = "744.687 GB"
				} else if i == 2 && j == 1 {
					diskSizeWanted = "744.687 GB"
				}
				if disk.Size != diskSizeWanted {
					t.Fatalf(`TestProcessHWMegaraidPercRaid disk.Size: %v should be: %v`, disk.Size, diskSizeWanted)
				}

				diskIntfWanted := "SATA"
				if disk.Intf != diskIntfWanted {
					t.Fatalf(`TestProcessHWMegaraidPercRaid disk.Intf: %v should be: %v`, disk.Intf, diskIntfWanted)
				}

				diskMediumWanted := "SSD"
				if disk.Medium != diskMediumWanted {
					t.Fatalf(`TestProcessHWMegaraidPercRaid disk.Medium: %v should be: %v`, disk.Medium, diskMediumWanted)
				}

				diskModelWanted := ""
				if i == 1 && j == 0 {
					diskModelWanted = "INTEL SSDSC2BB800H4"
				} else if i == 1 && j == 1 {
					diskModelWanted = "INTEL SSDSC2KB960G8"
				} else if i == 2 && j == 0 {
					diskModelWanted = "INTEL SSDSC2BB800H4"
				} else if i == 2 && j == 1 {
					diskModelWanted = "INTEL SSDSC2BB800H4"
				}
				if disk.Model != diskModelWanted {
					t.Fatalf(`TestProcessHWMegaraidPercRaid disk.Model: %v should be: %v`, disk.Model, diskModelWanted)
				}

				diskSerialNumberWanted := ""
				if i == 1 && j == 0 {
					diskSerialNumberWanted = "TESTSERIALNUMBER" + strconv.Itoa(j+1)
				} else if i == 1 && j == 1 {
					diskSerialNumberWanted = "TESTSERIALNUMBER" + strconv.Itoa(j+1)
				} else if i == 2 && j == 0 {
					diskSerialNumberWanted = "TESTSERIALNUMBER" + strconv.Itoa(j+3)
				} else if i == 2 && j == 1 {
					diskSerialNumberWanted = "TESTSERIALNUMBER" + strconv.Itoa(j+3)
				}
				if disk.SerialNumber != diskSerialNumberWanted {
					t.Fatalf(`TestProcessHWMegaraidPercRaid disk.SerialNumber: %v should be: %v`, disk.SerialNumber, diskSerialNumberWanted)
				}

				diskOsDeviceWanted := ""
				if disk.OsDevice != diskOsDeviceWanted {
					t.Fatalf(`TestProcessHWMegaraidPercRaid disk.OsDevice: %v should be: nil`, disk.OsDevice)
				}
			}
		}

		newRaidControllerOsDeviceWanted := "TESTOSRAIDDEVICE"
		if newRaid.OsDevice != newRaidControllerOsDeviceWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newRaid.OsDevice: %v should be: %v`, newRaid.OsDevice, newRaidControllerOsDeviceWanted)
		}
	}

	//fmt.Println("---------------- newNoRaidDisks --------------------")
	//spew.Dump(newNoRaidDisks)
	for _, newNoRaidDisk := range newNoRaidDisks {
		newNoRaidDiskControllerId := "mega-0"
		if newNoRaidDisk.ControllerId != newNoRaidDiskControllerId {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newNoRaidDisk.ControllerId: %v should be: %v`, newNoRaidDisk.ControllerId, newNoRaidDiskControllerId)
		}

		newNoRaidDiskEidSlotWanted := "252:4"
		if newNoRaidDisk.EidSlot != newNoRaidDiskEidSlotWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newNoRaidDisk.EidSlot: %v should be: %v`, newNoRaidDisk.EidSlot, newNoRaidDiskEidSlotWanted)
		}

		newNoRaidDiskStateWanted := "BogusDisk"
		if newNoRaidDisk.State != newNoRaidDiskStateWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newNoRaidDisk.State: %v should be: %v`, newNoRaidDisk.State, newNoRaidDiskStateWanted)
		}

		newNoRaidDiskSizeWanted := "223.062 GB"
		if newNoRaidDisk.Size != newNoRaidDiskSizeWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newNoRaidDisk.Size: %v should be: %v`, newNoRaidDisk.Size, newNoRaidDiskSizeWanted)
		}

		newNoRaidDiskIntfWanted := "SATA"
		if newNoRaidDisk.Intf != newNoRaidDiskIntfWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newNoRaidDisk.Intf: %v should be: %v`, newNoRaidDisk.Intf, newNoRaidDiskIntfWanted)
		}

		newNoRaidDiskMediumWanted := "SSD"
		if newNoRaidDisk.Medium != newNoRaidDiskMediumWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newNoRaidDisk.Medium: %v should be: %v`, newNoRaidDisk.Medium, newNoRaidDiskMediumWanted)
		}

		newNoRaidDiskModelWanted := "INTEL SSDSC2KB240G8"
		if newNoRaidDisk.Model != newNoRaidDiskModelWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newNoRaidDisk.Model: %v should be: %v`, newNoRaidDisk.Model, newNoRaidDiskModelWanted)
		}

		newNoRaidDiskSerialNumberWanted := "TESTSERIALNUMBER5"
		if newNoRaidDisk.SerialNumber != newNoRaidDiskSerialNumberWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newNoRaidDisk.SerialNumber: %v should be: %v`, newNoRaidDisk.SerialNumber, newNoRaidDiskSerialNumberWanted)
		}

		newNoRaidDiskOsDeviceWanted := "JBOD-BogusDisk-OSUnknown"
		if newNoRaidDisk.OsDevice != newNoRaidDiskOsDeviceWanted {
			t.Fatalf(`TestProcessHWMegaraidPercRaid newNoRaidDisk.OsDevice: %v should be: %v`, newNoRaidDisk.OsDevice, newNoRaidDiskOsDeviceWanted)
		}
	}
}

// Test ProcessHWMegaraidPercRaid outputStderr
func TestProcessHWMegaraidPercRaidOutputStderr(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestProcessHWMegaraidPercRaidOutputStderr function")
		var outputStdout, outputStderr bytes.Buffer
		// storcli /call show all
		// No CacheCade disks in mocked output because I dont have any of it in my servers
		outputStderr.WriteString("RANDOM ERROR")
		return &outputStdout, &outputStderr, nil
	}

	_, _, _, err := ProcessHWMegaraidPercRaid("mega")
	if err == nil {
		t.Fatalf(`TestProcessHWMegaraidPercRaidOutputStderr returned nil error, it should be != nil`)
	}
}
