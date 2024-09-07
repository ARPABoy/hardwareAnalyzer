package btrfs

import (
	"bytes"
	"hardwareAnalyzer/utils"
	"math"
	"strconv"
	"testing"

	human "github.com/dustin/go-humanize"
)

// Test CheckBtrfsRaid
func TestCheckBtrfsRaid(t *testing.T) {
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
		outputStdout.WriteString("Label: none  uuid: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXX")
		return &outputStdout, &outputStderr, nil
	}

	btrfsRaidCheck, err := CheckBtrfsRaid()
	if err != nil {
		t.Fatalf(`TestCheckBtrfsRaid: error: %s`, err)
	}
	if !btrfsRaidCheck {
		t.Fatalf(`TestCheckBtrfsRaid: %v, want TRUE`, btrfsRaidCheck)
	}
}

// Test CheckBtrfsRaid no raid
func TestCheckBtrfsRaidNoRaid(t *testing.T) {
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
		outputStdout.WriteString("RANDOM OUTPUT")
		return &outputStdout, &outputStderr, nil
	}

	btrfsRaidCheck, err := CheckBtrfsRaid()
	if err != nil {
		t.Fatalf(`TestCheckBtrfsRaidNoRaid: error: %s`, err)
	}
	if btrfsRaidCheck {
		t.Fatalf(`TestCheckBtrfsRaidNoRaid: %v, want FALSE`, btrfsRaidCheck)
	}
}

// Test CheckBtrfsRaid Error
func TestCheckBtrfsRaidError(t *testing.T) {
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
		// Write error to stderr:
		outputStderr.WriteString("RANDOM ERROR")
		return &outputStdout, &outputStderr, nil
	}

	btrfsRaidCheck, err := CheckBtrfsRaid()
	if err == nil {
		t.Fatalf(`TestCheckBtrfsRaidError: error should be != nil`)
	}
	if btrfsRaidCheck {
		t.Fatalf(`TestCheckBtrfsRaidError: %v, want FALSE`, btrfsRaidCheck)
	}
}

// TODO GetBtrfsRaidType, gets binary executor directly

// Test GetBtrfsRaidSize
func TestGetBtrfsRaidSize(t *testing.T) {
	// Generate btrfs raid with disks
	raid := utils.RaidStruct{
		ControllerId: "btrfs-0",
		RaidLevel:    0,
		RaidType:     "RAID10",
		State:        "ONLINE",
		Dg:           "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXX",
	}

	var totalSumDisks float64
	var raidSize float64
	var diskSize float64
	totalSumDisks = 0
	// TiB: multiplierFactor -> 4
	multiplierFactor := 4

	for i := 1; i <= 4; i++ {
		diskSize = float64(i) * math.Pow(float64(1024), float64(multiplierFactor))
		totalSumDisks = totalSumDisks + diskSize

		testSizeValue := strconv.Itoa(i) + ".0 TiB"
		//fmt.Println("testSizeValue: ", testSizeValue)
		btrfsDisk := utils.DiskStruct{
			ControllerId: "btrfs-0",
			Dg:           "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXX",
			State:        "ONLINE",
			Size:         testSizeValue,
			OsDevice:     "XYZ",
		}
		raid.AddDisk(btrfsDisk)

	}
	// RAID10
	raidSize = totalSumDisks / 2
	raidSizeFinal1 := human.Bytes(uint64(raidSize))
	raidSizeFinal1 = raidSizeFinal1 + " Aprox"
	//fmt.Println("raidSizeFinal1: ", raidSizeFinal1)

	// Check raid size
	raidSizeFinal2, _ := GetBtrfsRaidSize(raid)
	//fmt.Println("raidSizeFinal2: ", raidSizeFinal2)

	if raidSizeFinal1 != raidSizeFinal2 {
		t.Fatalf(`TestGetBtrfsRaidSize: raidSizeFinal1: %v must be equal to raidSizeFinal2: %v`, raidSizeFinal1, raidSizeFinal2)
	}
}

// Test ProcessBtrfsRaid
func TestProcessBtrfsRaid(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	getBtrfsRaidTypeOri := GetBtrfsRaidType
	getBtrfsRaidSizeOri := GetBtrfsRaidSize
	getDiskDataOri := utils.GetDiskData
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
		GetBtrfsRaidType = getBtrfsRaidTypeOri
		GetBtrfsRaidSize = getBtrfsRaidSizeOri
		utils.GetDiskData = getDiskDataOri
	}()

	// Mocked functions, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		var outputStdout, outputStderr bytes.Buffer

		// Simulate RAID1
		mockedData := `
			Label: 'defaultbtrfs'  uuid: XXXXXX-XXXX-XXXX-XXXX-XXXXXXXX
			Total devices 2 FS bytes used 1.50TiB
			devid    1 size 3.49TiB used 2.04TiB path /dev/XYZ1
			devid    2 size 3.49TiB used 2.04TiB path /dev/XYZ2
		`
		outputStdout.WriteString(mockedData)
		return &outputStdout, &outputStderr, nil
	}

	// Mocked functions, this way we can run unit tests in servers without hardware raid controller installed.
	GetBtrfsRaidType = func(device string) (string, error) {
		return "RAID1", nil
	}

	// Mocked functions, this way we can run unit tests in servers without hardware raid controller installed.
	GetBtrfsRaidSize = func(raid utils.RaidStruct) (string, error) {
		return "3.49TiB", nil
	}

	// Mocked functions, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetDiskData = func(diskDrive string) (string, string, string, string, error) {
		return "AlfaExploit-SerialNumber", "AlfaExploit-Model", "AlfaExploit-Intf", "AlfaExploit-Medium", nil
	}

	// Test
	var newControllers []utils.ControllerStruct
	var newRaids []utils.RaidStruct
	newControllers, newRaids, err := ProcessBtrfsRaid("btrfs")
	if err != nil {
		t.Fatalf(`TestProcessBtrfsRaid: error: %s`, err)
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

		// fmt.Println("Controller ID: ", controller.Id)
		// fmt.Println("Controller Manufacturer: ", controller.Manufacturer)
		// fmt.Printf("Controller Model: |%s|\n", controller.Model)
		// fmt.Println("Controller Status: ", controller.Status)

		controllerStatus = controller.Status
	}

	controllerIdWanted := "btrfs-0"
	if controllerId != controllerIdWanted {
		t.Fatalf(`TestProcessHWAdaptecRaid: controllerId: %s muts match %v`, controllerId, controllerIdWanted)
	}

	controllerManufacturerWanted := "btrfs"
	if controllerManufacturer != controllerManufacturerWanted {
		t.Fatalf(`TestProcessHWAdaptecRaid: controllerManufacturer: %s muts match %v`, controllerManufacturer, controllerManufacturerWanted)
	}

	controllerModelWanted := "Btrfs"
	if controllerModel != controllerModelWanted {
		t.Fatalf(`TestProcessHWAdaptecRaid: controllerModel: %s muts match %v`, controllerModel, controllerModelWanted)
	}

	controllerStatusWanted := "Good"
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

		raidControllerIdWanted := "btrfs-0"
		if raidControllerId != raidControllerIdWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidControllerId: %s muts match %v`, raidControllerId, raidControllerIdWanted)
		}

		raidRaidLevelWanted := 0
		if raidRaidLevel != raidRaidLevelWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidRaidLevel: %v muts match %v`, raidRaidLevel, raidRaidLevelWanted)
		}

		raidDgWanted := "XXXXXX-XXXX-XXXX-XXXX-XXXXXXXX"
		if raidDg != raidDgWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidDg: %s muts match %v`, raidDg, raidDgWanted)
		}

		raidRaidTypeWanted := "RAID1"
		if raidRaidType != raidRaidTypeWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidRaidType: %s muts match %v`, raidRaidType, raidRaidTypeWanted)
		}

		raidStateWanted := "ONLINE"
		if raidState != raidStateWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidState: %s muts match %v`, raidState, raidStateWanted)
		}

		raidSizeWanted := "3.49TiB"
		if raidSize != raidSizeWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidSize: %s muts match %v`, raidSize, raidSizeWanted)
		}

		// Btrfs considers the first raid device as OS device
		raidOsDeviceWanted := "XYZ1"
		if raidOsDevice != raidOsDeviceWanted {
			t.Fatalf(`TestProcessHWAdaptecRaid: raidOsDevice: %s muts match %v`, raidOsDevice, raidOsDeviceWanted)
		}

		diskControllerId := ""
		diskDg := ""
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
			diskState = disk.State
			diskSize = disk.Size
			diskIntf = disk.Intf
			diskMedium = disk.Medium
			diskModel = disk.Model
			diskSerialNumber = disk.SerialNumber

			// fmt.Println("  Disk ControllerId: ", disk.ControllerId)
			// fmt.Println("  Disk Dg: ", disk.Dg)
			// fmt.Println("  Disk State: ", disk.State)
			// fmt.Println("  Disk Size: ", disk.Size)
			// fmt.Println("  Disk Intf: ", disk.Intf)
			// fmt.Println("  Disk Medium: ", disk.Medium)
			// fmt.Println("  Disk Model: ", disk.Model)
			// fmt.Println("  Disk SerialNumber: ", disk.SerialNumber)
			// fmt.Println("  Disk OsDevice: ", disk.OsDevice)

			diskOsDevice = disk.OsDevice

			diskControllerIdWanted := "btrfs-0"
			if diskControllerId != diskControllerIdWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskControllerId: %s muts match %v`, diskControllerId, diskControllerIdWanted)
			}

			diskDgWanted := "XXXXXX-XXXX-XXXX-XXXX-XXXXXXXX"
			if diskDg != diskDgWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskDg: %v muts match %v`, diskDg, diskDgWanted)
			}

			diskStateWanted := "ONLINE"
			if diskState != diskStateWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskState: %s muts match %v`, diskState, diskStateWanted)
			}

			testSizeWanted := "3.49 TiB"
			if diskSize != testSizeWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskSize: %s muts match %v`, diskSize, testSizeWanted)
			}

			diskIntfWanted := "AlfaExploit-Intf"
			if diskIntf != diskIntfWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskIntf: %s muts match %v`, diskIntf, diskIntfWanted)
			}

			diskMediumWanted := "AlfaExploit-Medium"
			if diskMedium != diskMediumWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskMedium: %s muts match %v`, diskMedium, diskMediumWanted)
			}

			diskModelWanted := "AlfaExploit-Model"
			if diskModel != diskModelWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskModel: %s muts match %v`, diskModel, diskModelWanted)
			}

			diskSerialNumberWanted := "AlfaExploit-SerialNumber"
			if diskSerialNumber != diskSerialNumberWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskSerialNumber: %s muts match %v`, diskSerialNumber, diskSerialNumberWanted)
			}

			diskOsDeviceWanted := "XYZ" + strconv.Itoa(i)
			if diskOsDevice != diskOsDeviceWanted {
				t.Fatalf(`TestProcessHWAdaptecRaid: diskOsDevice: %s muts match %v`, diskOsDevice, diskOsDeviceWanted)
			}
		}
	}
}
