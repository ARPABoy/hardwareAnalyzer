package adaptec

import (
	"bytes"
	"hardwareAnalyzer/hardwarecontrollerscommon"
	"hardwareAnalyzer/utils"
	"math"
	"strconv"
	"testing"

	human "github.com/dustin/go-humanize"
)

// Test CheckAdaptecRaid
func TestCheckAdaptecRaid(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function: TestCheckAdaptecRaid")
		var outputStdout, outputStderr bytes.Buffer
		outputStdout.WriteString("Controllers found: 1")
		return &outputStdout, &outputStderr, nil
	}

	adaptecRaidCheck, err := CheckAadaptecRaid()
	if err != nil {
		t.Fatalf(`TestCheckAdaptecRaid: error: %s`, err)
	}
	if !adaptecRaidCheck {
		t.Fatalf(`TestCheckAdaptecRaid: %v, want TRUE`, adaptecRaidCheck)
	}
}

// Test CheckAdaptecRaidControllerNotFound
func TestCheckAdaptecRaidControllerNotFound(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function: TestCheckAdaptecRaidControllerNotFound")
		var outputStdout, outputStderr bytes.Buffer
		outputStdout.WriteString("Controllers found: 0")
		return &outputStdout, &outputStderr, nil
	}

	adaptecRaidCheck, err := CheckAadaptecRaid()
	if err != nil {
		t.Fatalf(`TestCheckAadaptecRaidControllerNotFound: error: %s`, err)
	}
	if adaptecRaidCheck {
		t.Fatalf(`TestCheckAadaptecRaidControllerNotFound: %v, want FALSE`, adaptecRaidCheck)
	}
}

// Test ProcessHWAdaptecRaid
func TestProcessHWAdaptecRaid(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	getRaidOSDeviceOri := hardwarecontrollerscommon.GetRaidOSDevice
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
		hardwarecontrollerscommon.GetRaidOSDevice = getRaidOSDeviceOri
	}()

	// Mocked functions, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function: TestProcessHWAdaptecRaid")
		var outputStdout, outputStderr bytes.Buffer
		if command == "LIST" {
			outputStdout.WriteString("Controllers found: 1\n")
		} else {
			mockedData := `
				Controller Status                        : AlfaExploitStatus
				Controller Model                         : Adaptec AlfaExploitModel
				Logical Device number 0
				Logical Device name                      : raidAlfaExploit
				RAID level                               : 10
				Unique Identifier                        : C66227AE
				Status of Logical Device                 : AlfaExploitStatusLogicalDevice
				Size                                     : 666 TB
				Segment 0                                : Present (1111MB, AlfaExploitIntf1, HDD1, Enclosure:1, Slot:1)      AlfaExploitDisk-1
				Segment 1                                : Present (2222MB, AlfaExploitIntf2, HDD2, Enclosure:2, Slot:2)      AlfaExploitDisk-2
				Segment 2                                : Present (3333MB, AlfaExploitIntf3, HDD3, Enclosure:3, Slot:3)      AlfaExploitDisk-3
				Segment 3                                : Present (4444MB, AlfaExploitIntf4, HDD4, Enclosure:4, Slot:4)      AlfaExploitDisk-4
				Physical Device information
				State                              		 : AlfaExploitStateDisk1
				Transfer Speed                   	     : SATA 111.0 Gb/s
				Reported Location                	     : Enclosure 1, Slot 1( Connector Unknown )
				Model        	                         : AlfaExploitDiskModel-1
				Total Size                        	     : 1111 MB
				SSD                                		 : No
				Temperature                        		 : Not Supported

				State                              		 : AlfaExploitStateDisk2
				Transfer Speed                   	     : SATA 222.0 Gb/s
				Reported Location                	     : Enclosure 2, Slot 2( Connector Unknown )
				Model        	                         : AlfaExploitDiskModel-2
				Total Size                        	     : 2222 MB
				SSD                                		 : No
				Temperature                        		 : Not Supported

				State                              		 : AlfaExploitStateDisk3
				Transfer Speed                   	     : SATA 333.0 Gb/s
				Reported Location                	     : Enclosure 3, Slot 3( Connector Unknown )
				Model        	                         : AlfaExploitDiskModel-3
				Total Size                        	     : 3333 MB
				SSD                                		 : No
				Temperature                        		 : Not Supported

				State                              		 : AlfaExploitStateDisk4
				Transfer Speed                   	     : SATA 444.0 Gb/s
				Reported Location                	     : Enclosure 4, Slot 4( Connector Unknown )
				Model        	                         : AlfaExploitDiskModel-4
				Total Size                        	     : 4444 MB
				SSD                                		 : No
				Temperature                        		 : Not Supported
			`
			outputStdout.WriteString(mockedData)
		}
		return &outputStdout, &outputStderr, nil
	}

	// Mocked functions, this way we can run unit tests in servers without hardware raid controller installed.
	hardwarecontrollerscommon.GetRaidOSDevice = func(manufacturer, controllerId, dg string) (string, error) {
		//fmt.Println("-- Executing mocked GetRaidOSDevice function: TestProcessHWAdaptecRaid")
		return "XYZ", nil
	}

	var newControllers []utils.ControllerStruct
	var newRaids []utils.RaidStruct
	newControllers, newRaids, _, err := ProcessHWAdaptecRaid("adaptec")
	if err != nil {
		t.Fatalf(`TestProcessHWAdaptecRaid: error: %s`, err)
	}

	//fmt.Println("****** Controllers *******")
	controllerId := ""
	controllerManufacturer := ""
	controllerModel := ""
	controllerStatus := ""
	for _, controller := range newControllers {
		//fmt.Println(controller)
		controllerId = controller.Id
		controllerManufacturer = controller.Manufacturer
		controllerModel = controller.Model

		//fmt.Println("Controller ID: ", controller.Id)
		//fmt.Println("Controller Manufacturer: ", controller.Manufacturer)
		//fmt.Printf("Controller Model: |%s|\n", controller.Model)
		//fmt.Println("Controller Status: ", controller.Status)

		controllerStatus = controller.Status
	}

	controllerIdWanted := "adaptec-0"
	if controllerId != controllerIdWanted {
		t.Fatalf(`TestProcessHWAdaptecRaid: controllerId: %s muts match %v`, controllerId, controllerIdWanted)
	}

	controllerManufacturerWanted := "adaptec"
	if controllerManufacturer != controllerManufacturerWanted {
		t.Fatalf(`TestProcessHWAdaptecRaid: controllerManufacturer: %s muts match %v`, controllerManufacturer, controllerManufacturerWanted)
	}

	controllerModelWanted := "Adaptec AlfaExploitModel"
	if controllerModel != controllerModelWanted {
		t.Fatalf(`TestProcessHWAdaptecRaid: controllerModel: %s muts match %v`, controllerModel, controllerModelWanted)
	}

	controllerStatusWanted := "AlfaExploitStatus"
	if controllerStatus != controllerStatusWanted {
		t.Fatalf(`TestProcessHWAdaptecRaid: controllerStatus: %s muts match %v`, controllerStatus, controllerStatusWanted)
	}

	//fmt.Println("****** Raids ******")
	raidControllerId := ""
	raidRaidLevel := 0
	raidDg := ""
	raidRaidType := ""
	raidState := ""
	raidSize := ""
	raidOsDevice := ""

	for _, raid := range newRaids {
		//fmt.Println(newRaid)
		raidControllerId = raid.ControllerId
		raidRaidLevel = raid.RaidLevel
		raidDg = raid.Dg
		raidRaidType = raid.RaidType
		raidState = raid.State
		raidSize = raid.Size

		//fmt.Println(" Raid ControllerId: ", raid.ControllerId)
		//fmt.Println(" Raid RaidLevel: ", raid.RaidLevel)
		//fmt.Println(" Raid Dg: ", raid.Dg)
		//fmt.Println(" Raid RaidType: ", raid.RaidType)
		//fmt.Println(" Raid State: ", raid.State)
		//fmt.Println(" Raid Size: ", raid.Size)
		//fmt.Println(" Raid OsDevice: ", raid.OsDevice)

		raidOsDevice = raid.OsDevice

		raidControllerIdWanted := "adaptec-0"
		if raidControllerId != raidControllerIdWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidControllerId: %s muts match %v`, raidControllerId, raidControllerIdWanted)
		}

		raidRaidLevelWanted := 0
		if raidRaidLevel != raidRaidLevelWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidRaidLevel: %v muts match %v`, raidRaidLevel, raidRaidLevelWanted)
		}

		raidDgWanted := "0"
		if raidDg != raidDgWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidDg: %s muts match %v`, raidDg, raidDgWanted)
		}

		raidRaidTypeWanted := "RAID10"
		if raidRaidType != raidRaidTypeWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidRaidType: %s muts match %v`, raidRaidType, raidRaidTypeWanted)
		}

		raidStateWanted := "AlfaExploitStatusLogicalDevice"
		if raidState != raidStateWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidState: %s muts match %v`, raidState, raidStateWanted)
		}

		raidSizeWanted := "732 TB"
		if raidSize != raidSizeWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidSize: %s muts match %v`, raidSize, raidSizeWanted)
		}

		raidOsDeviceWanted := "XYZ"
		if raidOsDevice != raidOsDeviceWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidOsDevice: %s muts match %v`, raidOsDevice, raidOsDeviceWanted)
		}

		diskControllerId := ""
		diskDg := ""
		diskEidSlot := ""
		diskState := ""
		diskSize := ""
		diskIntf := ""
		diskMedium := ""
		diskModel := ""
		diskSerialNumber := ""
		diskOsDevice := ""
		for i, disk := range raid.Disks {
			i++
			//fmt.Println("")
			diskControllerId = disk.ControllerId
			diskDg = disk.Dg
			diskEidSlot = disk.EidSlot
			diskState = disk.State
			diskSize = disk.Size
			diskIntf = disk.Intf
			diskMedium = disk.Medium
			diskModel = disk.Model
			diskSerialNumber = disk.SerialNumber

			//fmt.Println("  Disk ControllerId: ", disk.ControllerId)
			//fmt.Println("  Disk Dg: ", disk.Dg)
			//fmt.Println("  Disk EidSlot: ", disk.EidSlot)
			//fmt.Println("  Disk State: ", disk.State)
			//fmt.Println("  Disk Size: ", disk.Size)
			//fmt.Println("  Disk Intf: ", disk.Intf)
			//fmt.Println("  Disk Medium: ", disk.Medium)
			//fmt.Println("  Disk Model: ", disk.Model)
			//fmt.Println("  Disk SerialNumber: ", disk.SerialNumber)
			//fmt.Println("  Disk OsDevice: ", disk.OsDevice)

			diskOsDevice = disk.OsDevice

			diskControllerIdWanted := "adaptec-0"
			if diskControllerId != diskControllerIdWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskControllerId: %s muts match %v`, diskControllerId, diskControllerIdWanted)
			}

			diskDgWanted := "0"
			if diskDg != diskDgWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskDg: %v muts match %v`, diskDg, diskDgWanted)
			}

			diskEidSlotWanted := strconv.Itoa(i) + ":" + strconv.Itoa(i)
			if diskEidSlot != diskEidSlotWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskEidSlot: %s muts match %v`, diskEidSlot, diskEidSlotWanted)
			}

			diskStateWanted := "AlfaExploitStateDisk" + strconv.Itoa(i)
			if diskState != diskStateWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskState: %s muts match %v`, diskState, diskStateWanted)
			}

			// Disk size must be compared in human form
			// Testing size disk is in MB: multiplierFactor -> 2"
			multiplierFactor := 2
			testSizeValue := strconv.Itoa(i) + strconv.Itoa(i) + strconv.Itoa(i) + strconv.Itoa(i)
			testSizeValueFloat, _ := strconv.ParseFloat(testSizeValue, 64)
			testSizeFloat := testSizeValueFloat * math.Pow(float64(1024), float64(multiplierFactor))
			testSizeWanted := human.Bytes(uint64(testSizeFloat))
			if diskSize != testSizeWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskSize: %s muts match %v`, diskSize, testSizeWanted)
			}

			diskIntfWanted := "AlfaExploitIntf" + strconv.Itoa(i)
			if diskIntf != diskIntfWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskIntf: %s muts match %v`, diskIntf, diskIntfWanted)
			}

			diskMediumWanted := "HDD" + strconv.Itoa(i)
			if diskMedium != diskMediumWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskMedium: %s muts match %v`, diskMedium, diskMediumWanted)
			}

			diskModelWanted := "AlfaExploitDiskModel-" + strconv.Itoa(i)
			if diskModel != diskModelWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskModel: %s muts match %v`, diskModel, diskModelWanted)
			}

			diskSerialNumberWanted := "AlfaExploitDisk-" + strconv.Itoa(i)
			if diskSerialNumber != diskSerialNumberWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskSerialNumber: %s muts match %v`, diskSerialNumber, diskSerialNumberWanted)
			}

			if diskOsDevice != "" {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskOsDevice: %s muts match nil`, diskOsDevice)
			}
		}
	}
}
