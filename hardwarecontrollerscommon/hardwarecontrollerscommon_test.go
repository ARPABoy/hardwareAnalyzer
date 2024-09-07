package hardwarecontrollerscommon

import (
	"bytes"
	"fmt"
	"hardwareAnalyzer/utils"
	"math/big"
	"os"
	"strings"
	"testing"
)

// Test GetJbodOsDevice
func TestGetJbodOsDevice(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	// It will search for first wwn-0x drive listed under /dev/disk/by-id/ and return SAS address, thats what storcli would do but with JBOD disks
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function")
		var outputStdout, outputStderr bytes.Buffer

		// Search an existent device regardless received data from function call parameters
		files, err := os.ReadDir("/dev/disk/by-id/")
		if err != nil {
			t.Fatalf(`TestGetJbodOsDevice - getCommandOutput Could not ReadDir: %s"`, err)
		}

		// Get WNN
		osWnn := "Unknown"
		for _, file := range files {
			//fmt.Println("Checking: ", file.Name())
			if strings.Contains(file.Name(), "wwn-0x") {
				fileNameData := strings.Split(file.Name(), "wwn-0x")
				osWnn = fileNameData[1]
				//fmt.Println("osWnn: ", osWnn)
				break
			}
		}

		switch manufacturer {
		case "mega", "perc":
			// Convert osWnn to decimal
			osWnnDecimal := new(big.Int)
			osWnnDecimal.SetString(osWnn, 16)
			//fmt.Println("osWnnDecimal: ", osWnnDecimal)

			// Add 1 to osWnn
			decimalOne := big.NewInt(1)
			sasAddress := new(big.Int)
			sasAddress.Add(osWnnDecimal, decimalOne)
			//fmt.Println("sasAddress: ", sasAddress)

			// Convert osWnn+1 to Hex value: sasAddress
			sasAddressHex := sasAddress.Text(16)
			//fmt.Println("sasAddressHex: ", sasAddressHex)
			outputStdout.WriteString("  0 Active 6.0Gb/s   0x" + sasAddressHex)
			return &outputStdout, &outputStderr, nil
		case "sas2ircu":
			diskString := `
				Enclosure #                             : 1
				Slot #                                  : 0
				GUID                                    : ` + osWnn
			outputStdout.WriteString(diskString)
			return &outputStdout, &outputStderr, nil
		case "adaptec":
			// I dont have any server with this configuration, so i cant test it
			return &outputStdout, &outputStderr, nil
		default:
			return &outputStdout, &outputStderr, fmt.Errorf("Error: Something went wrong getting binary executor, obtained: Unknown")
		}
	}

	// mocked function getCommandOutput always returns SAS address of first wwn-0x drive listed under /dev/disk/by-id/
	// Get OS device to check it against obtained via getCommandOutput output
	files, err := os.ReadDir("/dev/disk/by-id/")
	if err != nil {
		t.Fatalf(`TestGetJbodOsDevice - Could not ReadDir: %s"`, err)
	}

	osDevice := "Unknown"
	for _, file := range files {
		if strings.Contains(file.Name(), "wwn-0x") {
			diskPath := "/dev/disk/by-id/" + file.Name()
			//fmt.Println("diskPath", diskPath)
			osDevice, err = os.Readlink(diskPath)
			if err != nil {
				t.Fatalf(`TestGetJbodOsDevice - Could not ReadDir: %s"`, err)
			}
			osDevice = strings.ReplaceAll(osDevice, "../", "")
			break
		}
	}
	//fmt.Println("osDevice: ", osDevice)

	// MEGA check
	// Unreal info, only to call getJbodOsDevice function with required parameters
	manufacturer := "mega"
	controllerId := "XX"
	eidslot := "XX:YY"
	osDeviceGetJbodOsDevice, err := GetJbodOsDevice(manufacturer, controllerId, eidslot)
	//fmt.Println("osDeviceGetJbodOsDevice: ", osDeviceGetJbodOsDevice)

	if err != nil {
		t.Fatalf(`TestGetJbodOsDevice MEGA, osDeviceGetJbodOsDevice returned error: %s`, err)
	}
	if osDeviceGetJbodOsDevice != osDevice || osDeviceGetJbodOsDevice == "Unknown" || osDevice == "Unknown" {
		t.Fatalf(`TestGetJbodOsDevice MEGA, osDeviceGetJbodOsDevice: %s should match with osDevice: %s and both of them to be != Unknown`, osDeviceGetJbodOsDevice, osDevice)
	}

	// SAS2IRCU check
	manufacturer = "sas2ircu"
	controllerId = "XX"
	// Our mocked function will return a string with this EID:SLOT
	eidslot = "1:0"
	osDeviceGetJbodOsDevice, err = GetJbodOsDevice(manufacturer, controllerId, eidslot)

	if err != nil {
		t.Fatalf(`TestGetJbodOsDevice SAS2IRCU, osDeviceGetJbodOsDevice returned error: %s`, err)
	}
	if osDeviceGetJbodOsDevice != osDevice || osDeviceGetJbodOsDevice == "Unknown" || osDevice == "Unknown" {
		t.Fatalf(`TestGetJbodOsDevice SAS2IRCU, osDeviceGetJbodOsDevice: %s should match with osDevice: %s and both of them to be != Unknown`, osDeviceGetJbodOsDevice, osDevice)
	}

	// Adaptec check, it must return: "Unknown", nil because I dont have this hardware configuration available to develop the linked code
	manufacturer = "adaptec"
	controllerId = "XX"
	eidslot = "XX:YY"
	osDeviceGetJbodOsDevice, err = GetJbodOsDevice(manufacturer, controllerId, eidslot)

	if err != nil {
		t.Fatalf(`TestGetJbodOsDevice ADAPTEC, osDeviceGetJbodOsDevice returned error: %s`, err)
	}
	if osDeviceGetJbodOsDevice != "Unknown" {
		t.Fatalf(`TestGetJbodOsDevice ADAPTEC, osDeviceGetJbodOsDevice: %s should match Unknown`, osDeviceGetJbodOsDevice)
	}

	// Any other manufacturer must return: "Unknown", Error
	manufacturer = "AlfaExploitTechnologies"
	controllerId = "XX"
	eidslot = "XX:YY"
	osDeviceGetJbodOsDevice, err = GetJbodOsDevice(manufacturer, controllerId, eidslot)

	if osDeviceGetJbodOsDevice != "Unknown" || err == nil {
		t.Fatalf(`TestGetJbodOsDevice OTHER, osDeviceGetJbodOsDevice: %s should match Unknown and err != nil: %s`, osDeviceGetJbodOsDevice, err)
	}
}

// Test GetRaidOSDevice
func TestGetRaidOSDevice(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	// It will search for first wwn-0x drive listed under /dev/disk/by-id/ and return SCSI NAA Id, thats what storcli would do
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function")
		var outputStdout, outputStderr bytes.Buffer

		// Search an existent device regardless received data from function call parameters
		files, err := os.ReadDir("/dev/disk/by-id/")
		if err != nil {
			t.Fatalf(`TestGetJbodOsDevice - getCommandOutput Could not ReadDir: %s"`, err)
		}

		// Get WNN
		osWnn := "Unknown"
		for _, file := range files {
			//fmt.Println("Checking: ", file.Name())
			if strings.Contains(file.Name(), "wwn-0x") {
				fileNameData := strings.Split(file.Name(), "wwn-0x")
				osWnn = fileNameData[1]
				//fmt.Println("osWnn: ", osWnn)
				break
			}
		}

		switch manufacturer {
		case "mega", "perc":
			outputStdout.WriteString("SCSI NAA Id = " + osWnn)
			return &outputStdout, &outputStderr, nil
		case "sas2ircu":
			diskString := `
				Enclosure #                             : 1
				Slot #                                  : 0
				GUID                                    : ` + osWnn
			outputStdout.WriteString(diskString)
			return &outputStdout, &outputStderr, nil
		case "adaptec":
			// I dont have any server with this configuration, so i cant test it
			return &outputStdout, &outputStderr, nil
		default:
			return &outputStdout, &outputStderr, fmt.Errorf("Error: Something went wrong getting binary executor, obtained: Unknown")
		}
	}

	// mocked function getCommandOutput always returns SCSI NAA Id of first wwn-0x drive listed under /dev/disk/by-id/
	// Get OS device to check it against obtained via getCommandOutput output
	files, err := os.ReadDir("/dev/disk/by-id/")
	if err != nil {
		t.Fatalf(`TestGetJbodOsDevice - Could not ReadDir: %s"`, err)
	}

	osDevice := "Unknown"
	for _, file := range files {
		if strings.Contains(file.Name(), "wwn-0x") {
			diskPath := "/dev/disk/by-id/" + file.Name()
			//fmt.Println("diskPath", diskPath)
			osDevice, err = os.Readlink(diskPath)
			if err != nil {
				t.Fatalf(`TestGetJbodOsDevice - Could not ReadDir: %s"`, err)
			}
			osDevice = strings.ReplaceAll(osDevice, "../", "")
			break
		}
	}
	//fmt.Println("osDevice: ", osDevice)

	// MEGA check
	// Unreal info, only to call getJbodOsDevice function with required parameters
	manufacturer := "mega"
	controllerId := "XX"
	dg := "XX:YY"
	osDeviceGetRaidOSDevice, err := GetRaidOSDevice(manufacturer, controllerId, dg)
	//fmt.Println("osDeviceGetRaidOSDevice: ", osDeviceGetRaidOSDevice)

	if err != nil {
		t.Fatalf(`TestGetRaidOSDevice MEGA, osDeviceGetRaidOSDevice returned error: %s`, err)
	}
	if osDeviceGetRaidOSDevice != osDevice || osDeviceGetRaidOSDevice == "Unknown" || osDevice == "Unknown" {
		t.Fatalf(`TestGetRaidOSDevice MEGA, osDeviceGetRaidOSDevice: %s should match with osDevice: %s and both of the to be != Unknown`, osDeviceGetRaidOSDevice, osDevice)
	}

	// SAS2IRCU check, cant be tested on a system without SAS2IRCU hardware controller installed
	// It checks /dev/disk/by-id/wwn-0x600508e000000000 + reversedWwid, wwn-0x600508e000000000 is only generated by OS when a SAS2IRCU controller is present
	// When its not present function returns Unknown, nil
	manufacturer = "sas2ircu"
	controllerId = "XX"
	dg = "XX:YY"
	osDeviceGetRaidOSDevice, err = GetRaidOSDevice(manufacturer, controllerId, dg)

	if err != nil {
		t.Fatalf(`TestGetRaidOSDevice SAS2IRCU, osDeviceGetRaidOSDevice returned error: %s`, err)
	}
	if osDeviceGetRaidOSDevice != "Unknown" {
		t.Fatalf(`TestGetRaidOSDevice SAS2IRCU, osDeviceGetRaidOSDevice: %s should match Unknown`, osDeviceGetRaidOSDevice)
	}

	// Adaptec check, cant be tested on a system without Adaptec hardware controller installed
	// It checks /dev/disk/by-id/scsi-SAdaptec_ + dg, scsi-SAdaptec_ is only generated by OS when a Adaptec controller is present
	// When its not present function returns Unknown, err
	manufacturer = "adaptec"
	controllerId = "XX"
	dg = "XX:YY"
	osDeviceGetRaidOSDevice, err = GetRaidOSDevice(manufacturer, controllerId, dg)

	if err == nil {
		t.Fatalf(`TestGetRaidOSDevice Adaptec, osDeviceGetRaidOSDevice returned nil error: %s`, err)
	}
	if osDeviceGetRaidOSDevice != "Unknown" {
		t.Fatalf(`TestGetRaidOSDevice Adaptec, osDeviceGetRaidOSDevice: %s should match Unknown`, osDeviceGetRaidOSDevice)
	}

	// Any other manufacturer must return: "Unknown", Error
	manufacturer = "AlfaExploitTechnologies"
	controllerId = "XX"
	dg = "XX:YY"
	osDeviceGetRaidOSDevice, err = GetRaidOSDevice(manufacturer, controllerId, dg)

	if err == nil {
		t.Fatalf(`TestGetRaidOSDevice Other, osDeviceGetRaidOSDevice returned nil error: %s`, err)
	}
	if osDeviceGetRaidOSDevice != "Unknown" {
		t.Fatalf(`TestGetRaidOSDevice Other, osDeviceGetRaidOSDevice: %s should match Unknown`, osDeviceGetRaidOSDevice)
	}
}

// Test CheckJbodDisks
func TestCheckJbodDisks(t *testing.T) {

	// -------- noRaidDisks -------------
	var noRaidDisks = []utils.NoRaidDiskStruct{}

	// Mega controllers X1
	manufacturer := "mega"
	controllerId := "0"
	diskUnits := []string{"sdb", "sdc", "sdd"}

	for _, diskUnit := range diskUnits {
		noRaidDisk := utils.NoRaidDiskStruct{
			ControllerId: manufacturer + "-" + controllerId,
			OsDevice:     "JBOD-" + diskUnit,
		}
		noRaidDisks = append(noRaidDisks, noRaidDisk)
	}

	// Mega controllers X2
	controllerId = "1"
	diskUnits = []string{"sde", "sdf", "sdg"}

	for _, diskUnit := range diskUnits {
		noRaidDisk := utils.NoRaidDiskStruct{
			ControllerId: manufacturer + "-" + controllerId,
			OsDevice:     "JBOD-" + diskUnit,
		}
		noRaidDisks = append(noRaidDisks, noRaidDisk)
	}

	// Adaptec controller X1
	manufacturer = "adaptec"
	controllerId = "0"
	diskUnits = []string{"sdh", "sdi", "sdj"}

	for _, diskUnit := range diskUnits {
		noRaidDisk := utils.NoRaidDiskStruct{
			ControllerId: manufacturer + "-" + controllerId,
			OsDevice:     "JBOD-" + diskUnit,
		}
		noRaidDisks = append(noRaidDisks, noRaidDisk)
	}

	// Adaptec controller X2
	manufacturer = "adaptec"
	controllerId = "1"
	diskUnits = []string{"sdk", "sdl", "sdm"}

	for _, diskUnit := range diskUnits {
		noRaidDisk := utils.NoRaidDiskStruct{
			ControllerId: manufacturer + "-" + controllerId,
			OsDevice:     "JBOD-" + diskUnit,
		}
		noRaidDisks = append(noRaidDisks, noRaidDisk)
	}

	// -------- newRaids -------------
	var newRaids = []utils.RaidStruct{}

	// Softraid
	newRaid := utils.RaidStruct{
		ControllerId: "softraid-0",
	}

	diskUnits = []string{"sdb", "sdc", "sdd"}
	for _, diskUnit := range diskUnits {
		diskDrive := utils.DiskStruct{
			ControllerId: "softraid-0",
			OsDevice:     diskUnit,
		}
		newRaid.AddDisk(diskDrive)
	}

	newRaids = append(newRaids, newRaid)

	// ZFS
	vdev := utils.RaidStruct{
		ControllerId: "zfs-0",
	}

	diskUnits = []string{"sde", "sdf", "sdg"}
	for _, diskUnit := range diskUnits {
		diskDrive := utils.DiskStruct{
			ControllerId: "zfs-0",
			OsDevice:     diskUnit,
		}
		vdev.AddDisk(diskDrive)
	}

	newRaids = append(newRaids, vdev)

	// Btrfs
	newRaid = utils.RaidStruct{
		ControllerId: "btrfs-0",
	}

	diskUnits = []string{"sdh", "sdi", "sdj"}
	for _, diskUnit := range diskUnits {
		diskDrive := utils.DiskStruct{
			ControllerId: "btrfs-0",
			OsDevice:     diskUnit,
		}
		newRaid.AddDisk(diskDrive)
	}

	newRaids = append(newRaids, newRaid)

	// LVM
	newRaid = utils.RaidStruct{
		ControllerId: "lvm-0",
	}

	diskUnits = []string{"sdk", "sdl", "sdm"}
	for _, diskUnit := range diskUnits {
		diskDrive := utils.DiskStruct{
			ControllerId: "lvm-0",
			OsDevice:     diskUnit,
		}
		newRaid.AddDisk(diskDrive)
	}

	newRaids = append(newRaids, newRaid)

	// test it
	// fmt.Println("---------------- noRaidDisks -------------------------")
	// spew.Dump(noRaidDisks)
	// fmt.Println("---------------- newRaids -------------------------")
	// spew.Dump(newRaids)
	// fmt.Println("-----------------------------------------")
	if err := CheckJbodDisks(newRaids, noRaidDisks); err != nil {
		t.Fatalf(`TestCheckJbodDisks: error: %s`, err)
	}
	// fmt.Println("---------------- noRaidDisks -------------------------")
	// spew.Dump(noRaidDisks)
	// fmt.Println("-----------------------------------------")
	for i, noRaidDisk := range noRaidDisks {
		//fmt.Println("")
		noRaidDiskNotFound := true
		noRaidDiskOsDeviceWanted := ""

		// Mega-Softraid disks
		if i == 0 || i == 1 || i == 2 {
			diskUnits := []string{"sdb", "sdc", "sdd"}
			for _, diskUnit := range diskUnits {
				noRaidDiskOsDeviceWanted = "JBOD-" + diskUnit + " SoftRaid"
				if noRaidDisk.OsDevice == noRaidDiskOsDeviceWanted {
					noRaidDiskNotFound = false
					break
				}
			}
			if noRaidDiskNotFound {
				t.Fatalf(`TestCheckJbodDisks SoftRaid noRaidDisk.OsDevice: %v should match: %v SoftRaid`, noRaidDisk.OsDevice, noRaidDisk.OsDevice)
			}
		}

		// Mega-ZFS disks
		if i == 3 || i == 4 || i == 5 {
			diskUnits := []string{"sde", "sdf", "sdg"}
			for _, diskUnit := range diskUnits {
				noRaidDiskOsDeviceWanted = "JBOD-" + diskUnit + " ZFS"
				if noRaidDisk.OsDevice == noRaidDiskOsDeviceWanted {
					noRaidDiskNotFound = false
					break
				}
			}
			if noRaidDiskNotFound {
				t.Fatalf(`TestCheckJbodDisks ZFS noRaidDisk.OsDevice: %v should match: %v ZFS`, noRaidDisk.OsDevice, noRaidDisk.OsDevice)
			}
		}

		// Adaptec-Btrfs disks
		if i == 6 || i == 7 || i == 8 {
			diskUnits := []string{"sdh", "sdi", "sdj"}
			for _, diskUnit := range diskUnits {
				noRaidDiskOsDeviceWanted = "JBOD-" + diskUnit + " Btrfs"
				if noRaidDisk.OsDevice == noRaidDiskOsDeviceWanted {
					noRaidDiskNotFound = false
					break
				}
			}
			if noRaidDiskNotFound {
				t.Fatalf(`TestCheckJbodDisks Btrfs noRaidDisk.OsDevice: %v should match: %v Btrfs`, noRaidDisk.OsDevice, noRaidDisk.OsDevice)
			}
		}

		// Adaptec-LVM disks
		if i == 9 || i == 10 || i == 11 {
			diskUnits := []string{"sdk", "sdl", "sdm"}
			for _, diskUnit := range diskUnits {
				noRaidDiskOsDeviceWanted = "JBOD-" + diskUnit + " LVM"
				if noRaidDisk.OsDevice == noRaidDiskOsDeviceWanted {
					noRaidDiskNotFound = false
					break
				}
			}
			if noRaidDiskNotFound {
				t.Fatalf(`TestCheckJbodDisks LVM noRaidDisk.OsDevice: %v should match: %v LVM`, noRaidDisk.OsDevice, noRaidDisk.OsDevice)
			}
		}
	}
}

// Test CheckHardRaidDisks
func TestCheckHardRaidDisks(t *testing.T) {
	var raids = []utils.RaidStruct{}
	var newRaids = []utils.RaidStruct{}

	// MegaRaid raid
	manufacturer := "mega"
	controllerId := "0"
	osDevice := "sdb"
	raid := utils.RaidStruct{
		ControllerId: manufacturer + "-" + controllerId,
		OsDevice:     osDevice,
	}
	raids = append(raids, raid)

	// ZFS raid
	vdev := utils.RaidStruct{
		ControllerId: "zfs-0",
	}
	// ZFS raid - disk same OsDevice as MegaRaid raid
	diskDrive := utils.DiskStruct{
		ControllerId: "zfs-0",
		OsDevice:     osDevice,
	}
	vdev.AddDisk(diskDrive)
	newRaids = append(newRaids, vdev)

	// Adaptec raid
	manufacturer = "adaptec"
	controllerId = "0"
	osDevice = "sdc"
	raid = utils.RaidStruct{
		ControllerId: manufacturer + "-" + controllerId,
		OsDevice:     osDevice,
	}
	raids = append(raids, raid)

	// Btrfs raid, same OsDevice: sdc
	raid = utils.RaidStruct{
		ControllerId: "btrfs-0",
	}
	// Btrfs raid - disk same OsDevice as Adaptec raid
	diskDrive = utils.DiskStruct{
		ControllerId: "btrfs-0",
		OsDevice:     osDevice,
	}
	raid.AddDisk(diskDrive)
	newRaids = append(newRaids, raid)

	// SAS2IRCU raid
	manufacturer = "sas2ircu"
	controllerId = "0"
	osDevice = "sdd"
	raid = utils.RaidStruct{
		ControllerId: manufacturer + "-" + controllerId,
		OsDevice:     osDevice,
	}
	raids = append(raids, raid)

	// Btrfs raid, same OsDevice: sdc
	raid = utils.RaidStruct{
		ControllerId: "lvm-0",
	}
	// LVM raid - disk same OsDevice as SAS2IRCU raid
	diskDrive = utils.DiskStruct{
		ControllerId: "lvm-0",
		OsDevice:     osDevice,
	}
	raid.AddDisk(diskDrive)
	newRaids = append(newRaids, raid)

	// fmt.Println("------------ BEFORE ------------------")
	// spew.Dump(raids)
	// fmt.Println("------------------------------")

	// test it
	err := CheckHardRaidDisks(newRaids, raids)
	if err != nil {
		t.Fatalf(`TestCheckHardRaidDisks error: %s`, err)
	}

	// fmt.Println("------------ AFTER ------------------")
	// spew.Dump(raids)
	// fmt.Println("------------------------------")

	for i, raid := range raids {
		if i == 0 && raid.OsDevice != "sdb ZFS" {
			t.Fatalf(`TestCheckHardRaidDisks ZFS raid using whole diks sdb, yet HW controller doesnt reflect info: %v, it should be sdb ZFS`, raid.OsDevice)
		}
		if i == 1 && raid.OsDevice != "sdc Btrfs" {
			t.Fatalf(`TestCheckHardRaidDisks Btrfs raid using whole diks sdc, yet HW controller doesnt reflect info: %v, it should be sdc Btrfs`, raid.OsDevice)
		}
		if i == 2 && raid.OsDevice != "sdd LVM" {
			t.Fatalf(`TestCheckHardRaidDisks LVM raid using whole diks sdd, yet HW controller doesnt reflect info: %v, it should be sdd LVM`, raid.OsDevice)
		}
	}
}
