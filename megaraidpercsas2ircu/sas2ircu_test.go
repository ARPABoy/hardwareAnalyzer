package megaraidpercsas2ircu

import (
	"bytes"
	"fmt"
	"hardwareAnalyzer/hardwarecontrollerscommon"
	"hardwareAnalyzer/utils"
	"testing"
)

// Test CheckSas2ircuRaid
func TestCheckSas2ircuRaid(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestCheckSas2ircuRaid function")
		var outputStdout, outputStderr bytes.Buffer
		outputStdout.WriteString("SAS2IRCU: Utility Completed Successfully.")
		return &outputStdout, &outputStderr, nil
	}

	sas2ircuRaidCheck, err := CheckSas2ircuRaid()
	if err != nil {
		t.Fatalf(`TestCheckSas2ircuRaid returned error: %s`, err)
	}

	if !sas2ircuRaidCheck {
		t.Fatalf(`TestCheckSas2ircuRaid sas2ircuRaidCheck: %v should match TRUE`, sas2ircuRaidCheck)
	}
}

// Test CheckSas2ircuRaid incorrect command output
func TestCheckSas2ircuRaidIncorrectCommandOutput(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestCheckSas2ircuRaidIncorrectCommandOutput function")
		var outputStdout, outputStderr bytes.Buffer
		outputStdout.WriteString("SAS2IRCU: RANDOM MESSAGE")
		return &outputStdout, &outputStderr, nil
	}

	sas2ircuRaidCheck, err := CheckSas2ircuRaid()
	if err != nil {
		t.Fatalf(`TestCheckSas2ircuRaidIncorrectCommandOutput returned error: %s`, err)
	}

	if sas2ircuRaidCheck {
		t.Fatalf(`TestCheckSas2ircuRaidIncorrectCommandOutput sas2ircuRaidCheck: %v should match FALSE`, sas2ircuRaidCheck)
	}
}

// Test CheckSas2ircuRaid MPTLib2 error
func TestCheckSas2ircuRaidMPTLib2Error(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestCheckSas2ircuRaidMPTLib2Error function")
		var outputStdout, outputStderr bytes.Buffer
		outputStdout.WriteString("MPTLib2 Error 1")
		return &outputStdout, &outputStderr, nil
	}

	sas2ircuRaidCheck, err := CheckSas2ircuRaid()
	if err != nil {
		t.Fatalf(`TestCheckSas2ircuRaidMPTLib2Error returned error: %s`, err)
	}

	if sas2ircuRaidCheck {
		t.Fatalf(`TestCheckSas2ircuRaidMPTLib2Error sas2ircuRaidCheck: %v should match FALSE`, sas2ircuRaidCheck)
	}
}

// Test CheckSas2ircuRaid outputStderr
func TestCheckSas2ircuRaidOutputStderr(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestCheckSas2ircuRaidOutputStderr function")
		var outputStdout, outputStderr bytes.Buffer
		outputStderr.WriteString("RANDOM MESSAGE")
		return &outputStdout, &outputStderr, nil
	}

	sas2ircuRaidCheck, err := CheckSas2ircuRaid()
	if err == nil {
		t.Fatalf(`TestCheckSas2ircuRaidOutputStderr returned nil error, it should be != nil`)
	}

	if sas2ircuRaidCheck {
		t.Fatalf(`TestCheckSas2ircuRaidOutputStderr sas2ircuRaidCheck: %v should match FALSE`, sas2ircuRaidCheck)
	}
}

// Test ProcessHWSas2ircuRaid
func TestProcessHWSas2ircuRaid(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	getRaidOSDeviceOri := hardwarecontrollerscommon.GetRaidOSDevice
	getJbodOsDeviceOri := hardwarecontrollerscommon.GetJbodOsDevice // unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
		hardwarecontrollerscommon.GetRaidOSDevice = getRaidOSDeviceOri
		hardwarecontrollerscommon.GetJbodOsDevice = getJbodOsDeviceOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestProcessHWSas2ircuRaid function")
		var outputStdout, outputStderr bytes.Buffer
		if command == "LIST" {
			outputStdout.WriteString(`
				0     SAS2008     1000h    72h   00h:03h:00h:00h      1028h   1f1eh
				SAS2IRCU: Utility Completed Successfully.
			`)
		} else if command == "0 DISPLAY" {
			outputStdout.WriteString(`
				Controller type                         : SAS2008
				------------------------------------------------------------------------
				IR volume 1
				Volume ID                               : 79
				Status of volume                        : Okay (OKY)
				Volume wwid                             : 08e444a2ffdbcba6
				RAID level                              : RAID1
				Size (in MB)                            : 476416
				PHY[0] Enclosure#/Slot#                 : 1:0
  				PHY[1] Enclosure#/Slot#                 : 1:1
				------------------------------------------------------------------------
				Device is a Hard disk
				Enclosure #                             : 1
				Slot #                                  : 0
				State                                   : Optimal (OPT)
				Size (in MB)/(in sectors)               : 476940/976773167
				Model Number                            : WDC WD5003ABYX-1
				Serial No                               : WDWMAYP1333673
				Protocol                                : SATA
				Drive Type                              : SATA_HDD

				Device is a Hard disk
				Enclosure #                             : 1
				Slot #                                  : 1
				State                                   : Optimal (OPT)
				Size (in MB)/(in sectors)               : 476940/976773167
				Model Number                            : WDC WD5003ABYX-1
				Serial No                               : WDWMAYP1324246
				Protocol                                : SATA
				Drive Type                              : SATA_HDD

				Device is a Hard disk
				Enclosure #                             : 1
				Slot #                                  : 2
				SAS Address                             : 4433221-1-0500-0000
				State                                   : Ready (RDY)
				Size (in MB)/(in sectors)               : 953869/1953525167
				Manufacturer                            : ATA
				Model Number                            : WDC WD10EFRX-68F
				Firmware Revision                       : 0A82
				Serial No                               : WDWCC4J5FLYN2F
				GUID                                    : 50014ee2b7133ef6
				Protocol                                : SATA
				Drive Type                              : SATA_HDD

				Device is a Hard disk
				Enclosure #                             : 1
				Slot #                                  : 3
				SAS Address                             : 4433221-1-0400-0000
				State                                   : Ready (RDY)
				Size (in MB)/(in sectors)               : 953869/1953525167
				Manufacturer                            : ATA
				Model Number                            : ST31000524AS
				Firmware Revision                       : JC45
				Serial No                               : 5VP7T3Z0
				GUID                                    : 5000c5002f864d33
				Protocol                                : SATA
				Drive Type                              : SATA_HDD
				------------------------------------------------------------------------
			`)
		}
		return &outputStdout, &outputStderr, nil
	}

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	hardwarecontrollerscommon.GetRaidOSDevice = func(manufacturer, controllerId, dg string) (string, error) {
		//fmt.Println("-- Executing mocked TestProcessHWSas2ircuRaid function")
		osDevice := "TESTOSRAIDDEVICE"
		return osDevice, nil
	}

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	hardwarecontrollerscommon.GetJbodOsDevice = func(manufacturer, controllerId, eidslot string) (string, error) {
		osDevice := "TESTOSDISKDEVICE"
		return osDevice, nil
	}

	newControllers, newRaids, newNoRaidDisks, err := ProcessHWSas2ircuRaid("sas2ircu")
	if err != nil {
		t.Fatalf(`TestProcessHWSas2ircuRaid returned error: %s`, err)
	}

	//fmt.Println("------------ newControllers -------------")
	//spew.Dump(newControllers)

	for _, newController := range newControllers {
		newControllerIdWanted := "sas2ircu-0"
		if newController.Id != newControllerIdWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid newController.Id: %v should match: %v`, newController.Id, newControllerIdWanted)
		}
	}

	//fmt.Println("------------ newRaids -------------")
	//spew.Dump(newRaids)
	for _, newRaid := range newRaids {
		newRaidControllerIdWanted := "sas2ircu-0"
		if newRaid.ControllerId != newRaidControllerIdWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid newRaid.ControllerId: %v should match: %v`, newRaid.ControllerId, newRaidControllerIdWanted)
		}

		newRaidRaidLevelWanted := 0
		if newRaid.RaidLevel != newRaidRaidLevelWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid newRaid.RaidLevel: %v should match: %v`, newRaid.RaidLevel, newRaidRaidLevelWanted)
		}

		newRaidDgWanted := "79"
		if newRaid.Dg != newRaidDgWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid newRaid.Dg: %v should match: %v`, newRaid.Dg, newRaidDgWanted)
		}

		newRaidRaidTypeWanted := "RAID1"
		if newRaid.RaidType != newRaidRaidTypeWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid newRaid.RaidType: %v should match: %v`, newRaid.RaidType, newRaidRaidTypeWanted)
		}

		newRaidStateWanted := "Okay(OKY)"
		if newRaid.State != newRaidStateWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid newRaid.State: %v should match: %v`, newRaid.State, newRaidStateWanted)
		}

		newRaidSizeWanted := "500 GB"
		if newRaid.Size != newRaidSizeWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid newRaid.Size: %v should match: %v`, newRaid.Size, newRaidSizeWanted)
		}

		for i, disk := range newRaid.Disks {
			diskControllerIdWanted := "sas2ircu-0"
			if disk.ControllerId != diskControllerIdWanted {
				t.Fatalf(`TestProcessHWSas2ircuRaid disk.ControllerId: %v should match: %v`, disk.ControllerId, diskControllerIdWanted)
			}

			diskDgWanted := "79"
			if disk.Dg != diskDgWanted {
				t.Fatalf(`TestProcessHWSas2ircuRaid disk.Dg: %v should match: %v`, disk.Dg, diskDgWanted)
			}

			diskEidSlotWanted := ""
			if i == 0 {
				diskEidSlotWanted = "1:0"
			} else {
				diskEidSlotWanted = "1:1"
			}
			if disk.EidSlot != diskEidSlotWanted {
				t.Fatalf(`TestProcessHWSas2ircuRaid disk.EidSlot: %v should match: %v`, disk.EidSlot, diskEidSlotWanted)
			}

			diskStateWanted := "Optimal(OPT)"
			if disk.State != diskStateWanted {
				t.Fatalf(`TestProcessHWSas2ircuRaid disk.State: %v should match: %v`, disk.State, diskStateWanted)
			}

			diskSizeWanted := "500 GB"
			if disk.Size != diskSizeWanted {
				t.Fatalf(`TestProcessHWSas2ircuRaid disk.Size: %v should match: %v`, disk.Size, diskSizeWanted)
			}

			diskIntfWanted := "SATA"
			if disk.Intf != diskIntfWanted {
				t.Fatalf(`TestProcessHWSas2ircuRaid disk.Intf: %v should match: %v`, disk.Intf, diskIntfWanted)
			}

			diskMediumWanted := "SATA_HDD"
			if disk.Medium != diskMediumWanted {
				t.Fatalf(`TestProcessHWSas2ircuRaid disk.Medium: %v should match: %v`, disk.Medium, diskMediumWanted)
			}

			diskModelWanted := "WDCWD5003ABYX-1"
			if disk.Model != diskModelWanted {
				t.Fatalf(`TestProcessHWSas2ircuRaid disk.Model: %v should match: %v`, disk.Model, diskModelWanted)
			}

			diskSerialNumberWanted := ""
			if i == 0 {
				diskSerialNumberWanted = "WDWMAYP1333673"
			} else {
				diskSerialNumberWanted = "WDWMAYP1324246"
			}
			if disk.SerialNumber != diskSerialNumberWanted {
				t.Fatalf(`TestProcessHWSas2ircuRaid disk.SerialNumber: %v should match %v`, disk.SerialNumber, diskSerialNumberWanted)
			}

			diskOsDeviceWanted := ""
			if disk.OsDevice != diskOsDeviceWanted {
				t.Fatalf(`TestProcessHWSas2ircuRaid disk.OsDevice: %v should be nil`, disk.OsDevice)
			}
		}

		newRaidOsDeviceWanted := "TESTOSRAIDDEVICE"
		if newRaid.OsDevice != newRaidOsDeviceWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid newRaid.OsDevice: %v should match: %v`, newRaid.OsDevice, newRaidOsDeviceWanted)
		}
	}

	//fmt.Println("------------ newNoRaidDisks -------------")
	//spew.Dump(newNoRaidDisks)

	for i, disk := range newNoRaidDisks {
		diskControllerIdWanted := "sas2ircu-0"
		if disk.ControllerId != diskControllerIdWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid disk.ControllerId: %v should match: %v`, disk.ControllerId, diskControllerIdWanted)
		}

		diskEidSlotWanted := ""
		if i == 0 {
			diskEidSlotWanted = "1:2"
		} else {
			diskEidSlotWanted = "1:3"
		}
		if disk.EidSlot != diskEidSlotWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid disk.EidSlot: %v should match: %v`, disk.EidSlot, diskEidSlotWanted)
		}

		diskStateWanted := "Ready(RDY)"
		if disk.State != diskStateWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid disk.State: %v should match: %v`, disk.State, diskStateWanted)
		}

		diskSizeWanted := "1.0 TB"
		if disk.Size != diskSizeWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid disk.Size: %v should match: %v`, disk.Size, diskSizeWanted)
		}

		diskIntfWanted := "SATA"
		if disk.Intf != diskIntfWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid disk.Intf: %v should match: %v`, disk.Intf, diskIntfWanted)
		}

		diskMediumWanted := "SATA_HDD"
		if disk.Medium != diskMediumWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid disk.Medium: %v should match: %v`, disk.Medium, diskMediumWanted)
		}

		diskModelWanted := ""
		if i == 0 {
			diskModelWanted = "WDCWD10EFRX-68F"
		} else {
			diskModelWanted = "ST31000524AS"
		}
		if disk.Model != diskModelWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid disk.Model: %v should match: %v`, disk.Model, diskModelWanted)
		}

		diskSerialNumberWanted := ""
		if i == 0 {
			diskSerialNumberWanted = "WDWCC4J5FLYN2F"
		} else {
			diskSerialNumberWanted = "5VP7T3Z0"
		}
		if disk.SerialNumber != diskSerialNumberWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid disk.SerialNumber: %v should match %v`, disk.SerialNumber, diskSerialNumberWanted)
		}

		diskOsDeviceWanted := "JBOD-TESTOSDISKDEVICE"
		if disk.OsDevice != diskOsDeviceWanted {
			t.Fatalf(`TestProcessHWSas2ircuRaid disk.OsDevice: %v should be nil`, disk.OsDevice)
		}
	}
}

// Test ProcessHWSas2ircuRaid outputStderr
func TestProcessHWSas2ircuRaidOutputStderr(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked TestProcessHWSas2ircuRaidOutputStderr function")
		var outputStdout, outputStderr bytes.Buffer

		outputStderr.WriteString("RANDOM ERROR")

		return &outputStdout, &outputStderr, nil
	}

	_, _, _, err := ProcessHWSas2ircuRaid("sas2ircu")
	if err == nil {
		t.Fatalf(`TestProcessHWSas2ircuRaidOutputStderr should return err != nil`)
	}
}

// Test ProcessHWSas2ircuRaid GetRaidOSDevice error
func TestProcessHWSas2ircuRaidGetRaidOSDeviceError(t *testing.T) {
	// Copy original functions content
	getRaidOSDeviceOri := hardwarecontrollerscommon.GetRaidOSDevice
	// unmock functions content
	defer func() {
		hardwarecontrollerscommon.GetRaidOSDevice = getRaidOSDeviceOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	hardwarecontrollerscommon.GetRaidOSDevice = func(manufacturer, controllerId, dg string) (string, error) {
		//fmt.Println("-- Executing mocked TestProcessHWSas2ircuRaidGetRaidOSDeviceError function")
		return "Unknown", fmt.Errorf("Error: RANDOM ERROR")
	}

	_, _, _, err := ProcessHWSas2ircuRaid("sas2ircu")
	if err == nil {
		t.Fatalf(`TestProcessHWSas2ircuRaidGetRaidOSDeviceError should return err != nil`)
	}
}

// Test ProcessHWSas2ircuRaid GetJbodOsDevice error
func TestProcessHWSas2ircuRaidGetJbodOsDeviceError(t *testing.T) {
	// Copy original functions content
	getJbodOsDeviceOri := hardwarecontrollerscommon.GetJbodOsDevice
	// unmock functions content
	defer func() {
		hardwarecontrollerscommon.GetJbodOsDevice = getJbodOsDeviceOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	hardwarecontrollerscommon.GetJbodOsDevice = func(manufacturer, controllerId, eidslot string) (string, error) {
		//fmt.Println("-- Executing mocked TestProcessHWSas2ircuRaidGetJbodOsDeviceError function")
		return "Unknown", fmt.Errorf("Error: RANDOM ERROR")
	}

	_, _, _, err := ProcessHWSas2ircuRaid("sas2ircu")
	if err == nil {
		t.Fatalf(`TestProcessHWSas2ircuRaidGetJbodOsDeviceError should return err != nil`)
	}
}
