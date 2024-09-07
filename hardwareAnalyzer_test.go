package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"hardwareAnalyzer/adaptec"
	"hardwareAnalyzer/btrfs"
	"hardwareAnalyzer/lvm"
	"hardwareAnalyzer/megaraidpercsas2ircu"
	"hardwareAnalyzer/regulardisks"
	"hardwareAnalyzer/softraid"
	"hardwareAnalyzer/utils"
	"hardwareAnalyzer/zfs"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
)

// Test main
func TestMain(t *testing.T) {
	// Copy original functions content
	// We cant unmock it using defer because maybe we need to make some prints in console for debugging
	osStdoutOri := os.Stdout
	osStderrOri := os.Stderr
	colorOutputOri := color.Output
	colorErrorOri := color.Error

	// All content written to w pipe, will be copied automatically to r pipe
	r, w, _ := os.Pipe()
	// Make Stdout/Stderr to be written to w pipe
	// Color module defines other Stdout/Stderr, so pipe them to w pipe too
	os.Stdout = w
	os.Stderr = w
	color.Output = w
	color.Error = w

	main()

	// Close w pipe
	w.Close()

	// Restore Stdout/Stderr to normal output
	os.Stdout = osStdoutOri
	os.Stderr = osStderrOri
	color.Output = colorOutputOri
	color.Error = colorErrorOri

	// Read all r pipe content
	out, _ := io.ReadAll(r)
	//fmt.Println("--- out ---")
	//fmt.Println(out)

	scanner := bufio.NewScanner(bytes.NewReader(out))
	bannerFound := false
	for scanner.Scan() {
		line := scanner.Text()
		//fmt.Println("-- LINE: ", line)
		if strings.Contains(line, "CodeName: Sistine Chapel") {
			bannerFound = true
			break
		}
	}

	if !bannerFound {
		t.Fatalf(`TestMain: No banner found`)
	}
}

// Test main -showInfo
func TestMainShowInfo(t *testing.T) {

	// Save original Args and restore on exit function
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	// Configure new Args
	os.Args = []string{"cmd", "-showInfo"}

	// Copy original functions content
	// We cant unmock it using defer because maybe we need to make some prints in console for debugging
	osStdoutOri := os.Stdout
	osStderrOri := os.Stderr
	colorOutputOri := color.Output
	colorErrorOri := color.Error

	// All content written to w pipe, will be copied automatically to r pipe
	r, w, _ := os.Pipe()

	// Make Stdout/Stderr to be written to w pipe
	// Color module defines other Stdout/Stderr, so pipe them to w pipe too
	os.Stdout = w
	os.Stderr = w
	color.Output = w
	color.Error = w

	main()

	// Close w pipe
	w.Close()

	// Restore Stdout/Stderr to normal output
	os.Stdout = osStdoutOri
	os.Stderr = osStderrOri
	color.Output = colorOutputOri
	color.Error = colorErrorOri

	// Read all r pipe content
	out, _ := io.ReadAll(r)
	//fmt.Println("--- out ---")
	//fmt.Println(out)

	scanner := bufio.NewScanner(bytes.NewReader(out))
	bannerFound := false
	for scanner.Scan() {
		line := scanner.Text()
		//fmt.Println("-- LINE: ", line)
		if strings.Contains(line, "alfaexploit.com") {
			bannerFound = true
			break
		}
	}

	if !bannerFound {
		t.Fatalf(`TestMainShowInfo: -showInfo banner not found`)
	}
}

// Test checkHardware
func TestCheckHardware(t *testing.T) {
	// Copy original functions content
	checkMegaraidPercOri := megaraidpercsas2ircu.CheckMegaraidPerc
	checkSas2ircuRaidOri := megaraidpercsas2ircu.CheckSas2ircuRaid
	checkAadaptecRaidOri := adaptec.CheckAadaptecRaid
	checkSoftRaidOri := softraid.CheckSoftRaid
	checkZFSRaidOri := zfs.CheckZFSRaid
	checkBtrfsRaidOri := btrfs.CheckBtrfsRaid
	checkLVMRaidOri := lvm.CheckLVMRaid

	// unmock functions content
	defer func() {
		megaraidpercsas2ircu.CheckMegaraidPerc = checkMegaraidPercOri
		megaraidpercsas2ircu.CheckSas2ircuRaid = checkSas2ircuRaidOri
		adaptec.CheckAadaptecRaid = checkAadaptecRaidOri
		softraid.CheckSoftRaid = checkSoftRaidOri
		zfs.CheckZFSRaid = checkZFSRaidOri
		btrfs.CheckBtrfsRaid = checkBtrfsRaidOri
		lvm.CheckLVMRaid = checkLVMRaidOri
	}()

	// Mocked functions.
	megaraidpercsas2ircu.CheckMegaraidPerc = func(manufacturer string) (bool, error) {
		return true, nil
	}
	megaraidpercsas2ircu.CheckSas2ircuRaid = func() (bool, error) {
		return true, nil
	}
	adaptec.CheckAadaptecRaid = func() (bool, error) {
		return true, nil
	}
	softraid.CheckSoftRaid = func() (bool, error) {
		return true, nil
	}
	zfs.CheckZFSRaid = func() (bool, error) {
		return true, nil
	}
	btrfs.CheckBtrfsRaid = func() (bool, error) {
		return true, nil
	}
	lvm.CheckLVMRaid = func() (bool, error) {
		return true, nil
	}

	megaRaidCheck, percRaidCheck, sas2ircuRaidCheck, adaptecRaidCheck, softRaidCheck, zfsRaidCheck, btrfsRaidCheck, lvmRaidCheck := checkHardware()
	if !megaRaidCheck || !percRaidCheck || !sas2ircuRaidCheck || !adaptecRaidCheck || !softRaidCheck || !zfsRaidCheck || !btrfsRaidCheck || !lvmRaidCheck {
		t.Fatalf(`TestCheckHardware: all check must return TRUE`)
	}
}

// Test checkHardware errors
func TestCheckHardwareErrors(t *testing.T) {
	// Copy original functions content
	checkMegaraidPercOri := megaraidpercsas2ircu.CheckMegaraidPerc
	checkSas2ircuRaidOri := megaraidpercsas2ircu.CheckSas2ircuRaid
	checkAadaptecRaidOri := adaptec.CheckAadaptecRaid
	checkSoftRaidOri := softraid.CheckSoftRaid
	checkZFSRaidOri := zfs.CheckZFSRaid
	checkBtrfsRaidOri := btrfs.CheckBtrfsRaid
	checkLVMRaidOri := lvm.CheckLVMRaid

	// unmock functions content
	defer func() {
		megaraidpercsas2ircu.CheckMegaraidPerc = checkMegaraidPercOri
		megaraidpercsas2ircu.CheckSas2ircuRaid = checkSas2ircuRaidOri
		adaptec.CheckAadaptecRaid = checkAadaptecRaidOri
		softraid.CheckSoftRaid = checkSoftRaidOri
		zfs.CheckZFSRaid = checkZFSRaidOri
		btrfs.CheckBtrfsRaid = checkBtrfsRaidOri
		lvm.CheckLVMRaid = checkLVMRaidOri
	}()

	// Mocked functions.
	megaraidpercsas2ircu.CheckMegaraidPerc = func(manufacturer string) (bool, error) {
		return false, fmt.Errorf("TEST ERROR")
	}

	megaraidpercsas2ircu.CheckSas2ircuRaid = func() (bool, error) {
		return false, fmt.Errorf("TEST ERROR")
	}

	adaptec.CheckAadaptecRaid = func() (bool, error) {
		return false, fmt.Errorf("TEST ERROR")
	}

	softraid.CheckSoftRaid = func() (bool, error) {
		return false, fmt.Errorf("TEST ERROR")
	}

	zfs.CheckZFSRaid = func() (bool, error) {
		return false, fmt.Errorf("TEST ERROR")
	}

	btrfs.CheckBtrfsRaid = func() (bool, error) {
		return false, fmt.Errorf("TEST ERROR")
	}

	lvm.CheckLVMRaid = func() (bool, error) {
		return false, fmt.Errorf("TEST ERROR")
	}

	megaRaidCheck, percRaidCheck, sas2ircuRaidCheck, adaptecRaidCheck, softRaidCheck, zfsRaidCheck, btrfsRaidCheck, lvmRaidCheck := checkHardware()
	if megaRaidCheck || percRaidCheck || sas2ircuRaidCheck || adaptecRaidCheck || softRaidCheck || zfsRaidCheck || btrfsRaidCheck || lvmRaidCheck {
		t.Fatalf(`TestCheckHardware: all check must return FALSE`)
	}
}

// Test inquireHardwareConfiguration
func TestInquireHardwareConfiguration(t *testing.T) {
	// Copy original functions content
	processHWMegaraidPercRaidOri := megaraidpercsas2ircu.ProcessHWMegaraidPercRaid
	processHWSas2ircuRaidOri := megaraidpercsas2ircu.ProcessHWSas2ircuRaid
	processHWAdaptecRaidOri := adaptec.ProcessHWAdaptecRaid
	processSoftRaidOri := softraid.ProcessSoftRaid
	processZFSRaidOri := zfs.ProcessZFSRaid
	processBtrfsRaidOri := btrfs.ProcessBtrfsRaid
	processLVMRaidOri := lvm.ProcessLVMRaid
	processRegularDisksOri := regulardisks.ProcessRegularDisks

	// unmock functions content
	defer func() {
		megaraidpercsas2ircu.ProcessHWMegaraidPercRaid = processHWMegaraidPercRaidOri
		megaraidpercsas2ircu.ProcessHWSas2ircuRaid = processHWSas2ircuRaidOri
		adaptec.ProcessHWAdaptecRaid = processHWAdaptecRaidOri
		softraid.ProcessSoftRaid = processSoftRaidOri
		zfs.ProcessZFSRaid = processZFSRaidOri
		btrfs.ProcessBtrfsRaid = processBtrfsRaidOri
		lvm.ProcessLVMRaid = processLVMRaidOri
		regulardisks.ProcessRegularDisks = processRegularDisksOri
	}()

	// Mocked functions.
	megaraidpercsas2ircu.ProcessHWMegaraidPercRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.RaidStruct, []utils.NoRaidDiskStruct, error) {
		var controllers = []utils.ControllerStruct{}
		var raids = []utils.RaidStruct{}
		var noRaidDisks = []utils.NoRaidDiskStruct{}

		if manufacturer == "mega" {
			controller := utils.ControllerStruct{
				Id:           "mega-0",
				Manufacturer: "mega",
				Model:        "LSI MegaRAID SAS 9271-4i",
				Status:       "Optimal",
			}
			controllers = append(controllers, controller)

			disk1 := utils.DiskStruct{
				EidSlot:      "252:0",
				State:        "Onln",
				Size:         "744.687 GB",
				Intf:         "SATA",
				Medium:       "SSD",
				Model:        "INTEL SSDSC2BB800H4",
				SerialNumber: "BTWH509601KE800CGN",
			}

			disk2 := utils.DiskStruct{
				EidSlot:      "252:1",
				State:        "Onln",
				Size:         "893.750 GB",
				Intf:         "SATA",
				Medium:       "SSD",
				Model:        "INTEL SSDSC2KB960G8",
				SerialNumber: "BTYF950108L5960CGN",
			}

			// Crear instancia de RaidStruct con los dos discos
			raid := utils.RaidStruct{
				ControllerId: "mega-0",
				RaidLevel:    1,
				RaidType:     "RAID1",
				State:        "Optimal",
				Size:         "744.687 GB",
				Disks:        []utils.DiskStruct{disk1, disk2},
				OsDevice:     "sda",
			}
			raids = append(raids, raid)
		}

		return controllers, raids, noRaidDisks, nil
	}

	megaraidpercsas2ircu.ProcessHWSas2ircuRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.RaidStruct, []utils.NoRaidDiskStruct, error) {
		var controllers = []utils.ControllerStruct{}
		var raids = []utils.RaidStruct{}
		var noRaidDisks = []utils.NoRaidDiskStruct{}

		return controllers, raids, noRaidDisks, nil
	}

	adaptec.ProcessHWAdaptecRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.RaidStruct, []utils.NoRaidDiskStruct, error) {
		var controllers = []utils.ControllerStruct{}
		var raids = []utils.RaidStruct{}
		var noRaidDisks = []utils.NoRaidDiskStruct{}

		return controllers, raids, noRaidDisks, nil
	}

	softraid.ProcessSoftRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.RaidStruct, error) {
		var controllers = []utils.ControllerStruct{}
		var raids = []utils.RaidStruct{}

		return controllers, raids, nil
	}

	zfs.ProcessZFSRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.PoolStruct, []utils.RaidStruct, error) {
		var controllers = []utils.ControllerStruct{}
		var pools = []utils.PoolStruct{}
		var vdevs = []utils.RaidStruct{}

		return controllers, pools, vdevs, nil
	}

	btrfs.ProcessBtrfsRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.RaidStruct, error) {
		var controllers = []utils.ControllerStruct{}
		var raids = []utils.RaidStruct{}

		return controllers, raids, nil
	}

	lvm.ProcessLVMRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.VolumeGroupStruct, []utils.RaidStruct, error) {
		var controllers = []utils.ControllerStruct{}
		var volumeGroups = []utils.VolumeGroupStruct{}
		var raids = []utils.RaidStruct{}

		return controllers, volumeGroups, raids, nil
	}

	regulardisks.ProcessRegularDisks = func(raids []utils.RaidStruct, noRaidDisks []utils.NoRaidDiskStruct) ([]utils.ControllerStruct, []utils.RaidStruct, error) {
		regularDiskControllers := []utils.ControllerStruct{}
		regularDiskRaids := []utils.RaidStruct{}

		return regularDiskControllers, regularDiskRaids, nil
	}

	// Call functions
	controllers, pools, volumeGroups, raids, noRaidDisks := inquireHardwareConfiguration(true, true, true, true, true, true, true, true)

	// Since we only mocked one function with real data
	// We must get controllers and raids only, all other structures must be empty
	if len(pools) != 0 || len(volumeGroups) != 0 || len(noRaidDisks) != 0 {
		t.Fatalf(`TestInquireHardwareConfiguration: pools, volumeGroups and noRaidDisks should be empty.`)
	}

	if len(controllers) != 1 || len(raids) != 1 {
		t.Fatalf(`TestInquireHardwareConfiguration: controllers and raids length should be 1.`)
	}

	for _, controller := range controllers {
		//spew.Dump(controller)
		wanted := "mega-0"
		if controller.Id != wanted {
			t.Fatalf(`TestInquireHardwareConfiguration: Bogus controller.Id: %v, wanted: %v.`, controller.Id, wanted)
		}

		wanted = "mega"
		if controller.Manufacturer != wanted {
			t.Fatalf(`TestInquireHardwareConfiguration: Bogus controller.Manufacturer: %v, wanted: %v.`, controller.Manufacturer, wanted)
		}

		wanted = "LSI MegaRAID SAS 9271-4i"
		if controller.Model != wanted {
			t.Fatalf(`TestInquireHardwareConfiguration: Bogus controller.Model: %v, wanted: %v.`, controller.Model, wanted)
		}

		wanted = "Optimal"
		if controller.Status != wanted {
			t.Fatalf(`TestInquireHardwareConfiguration: Bogus controller.Status: %v, wanted: %v.`, controller.Status, wanted)
		}
	}
	for _, raid := range raids {
		//spew.Dump(raid)

		if len(raid.Disks) != 2 {
			t.Fatalf(`TestInquireHardwareConfiguration: raid with incorrect number of disks.`)
		}

		wanted := "mega-0"
		if raid.ControllerId != wanted {
			t.Fatalf(`TestInquireHardwareConfiguration: Bogus raid.ControllerId: %v, wanted: %v.`, raid.ControllerId, wanted)
		}

		intWanted := 1
		if raid.RaidLevel != intWanted {
			t.Fatalf(`TestInquireHardwareConfiguration: Bogus raid.RaidLevel: %v, wanted: %v.`, raid.RaidLevel, intWanted)
		}

		wanted = "RAID1"
		if raid.RaidType != "RAID1" {
			t.Fatalf(`TestInquireHardwareConfiguration: Bogus raid.RaidType: %v, wanted: %v.`, raid.RaidType, wanted)
		}

		wanted = "Optimal"
		if raid.State != "Optimal" {
			t.Fatalf(`TestInquireHardwareConfiguration: Bogus raid.State: %v, wanted: %v.`, raid.State, wanted)
		}

		wanted = "744.687 GB"
		if raid.Size != "744.687 GB" {
			t.Fatalf(`TestInquireHardwareConfiguration: Bogus raid.Size: %v, wanted: %v.`, raid.Size, wanted)
		}

		wanted = "sda"
		if raid.OsDevice != "sda" {
			t.Fatalf(`TestInquireHardwareConfiguration: Bogus raid.OsDevice: %v, wanted: %v.`, raid.OsDevice, wanted)
		}

		for i, disk := range raid.Disks {
			if i == 0 {
				wanted := "252:0"
				if disk.EidSlot != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk1: Bogus disk.EidSlot: %v, wanted: %v.`, disk.EidSlot, wanted)
				}

				wanted = "Onln"
				if disk.State != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk1: Bogus disk.State: %v, wanted: %v.`, disk.State, wanted)
				}

				wanted = "744.687 GB"
				if disk.Size != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk1: Bogus disk.Size: %v, wanted: %v.`, disk.Size, wanted)
				}

				wanted = "SATA"
				if disk.Intf != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk1: Bogus disk.Intf: %v, wanted: %v.`, disk.Intf, wanted)
				}

				wanted = "SSD"
				if disk.Medium != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk1: Bogus disk.Medium: %v, wanted: %v.`, disk.Medium, wanted)
				}

				wanted = "INTEL SSDSC2BB800H4"
				if disk.Model != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk1: Bogus disk.Model: %v, wanted: %v.`, disk.Model, wanted)
				}

				wanted = "BTWH509601KE800CGN"
				if disk.SerialNumber != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk2: Bogus disk.SerialNumber: %v, wanted: %v.`, disk.SerialNumber, wanted)
				}
			} else {
				wanted := "252:1"
				if disk.EidSlot != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk1: Bogus disk.EidSlot: %v, wanted: %v.`, disk.EidSlot, wanted)
				}

				wanted = "Onln"
				if disk.State != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk1: Bogus disk.State: %v, wanted: %v.`, disk.State, wanted)
				}

				wanted = "893.750 GB"
				if disk.Size != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk1: Bogus disk.Size: %v, wanted: %v.`, disk.Size, wanted)
				}

				wanted = "SATA"
				if disk.Intf != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk1: Bogus disk.Intf: %v, wanted: %v.`, disk.Intf, wanted)
				}

				wanted = "SSD"
				if disk.Medium != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk1: Bogus disk.Medium: %v, wanted: %v.`, disk.Medium, wanted)
				}

				wanted = "INTEL SSDSC2KB960G8"
				if disk.Model != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk1: Bogus disk.Model: %v, wanted: %v.`, disk.Model, wanted)
				}

				wanted = "BTYF950108L5960CGN"
				if disk.SerialNumber != wanted {
					t.Fatalf(`TestInquireHardwareConfiguration-disk2: Bogus disk.SerialNumber: %v, wanted: %v.`, disk.SerialNumber, wanted)
				}
			}
		}
	}
}
