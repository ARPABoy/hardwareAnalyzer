package megaraidpercsas2ircu

import (
	"bufio"
	"fmt"
	"hardwareAnalyzer/hardwarecontrollerscommon"
	"hardwareAnalyzer/utils"
	"strings"

	"github.com/fatih/color"
)

// Hardware controller functions:
// Megaraid: Pure MegaRAID controller
// Dell(Megaraid): PERC and SAS2IRCU controllers

// Function only used from processHWMegaraidPercRaid
// Function as variable in order to be possible to mock it from unitary tests
var GetMegaraidPercDriveSerialNumber = func(manufacturer, controllerId, eidSlot string) (string, error) {
	// fmt.Println("-- getMegaraidPercDriveSerialNumber --")
	// fmt.Println("controllerId: ", controllerId)
	// fmt.Println("eidSlot: ", eidSlot)

	eidSlotData := strings.Split(eidSlot, ":")
	eid := eidSlotData[0]
	slot := eidSlotData[1]

	command := "/c" + controllerId + " /e" + eid + " /s" + slot + " show all"
	//fmt.Println("Command: ", command)
	outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "getMegaraidPercDriveSerialNumber", command)
	if err != nil {
		color.Red("++ ERROR: Something went wrong executing command %s: %v", command, err)
		return "Unknown", fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
	}
	if len(outputStderr.String()) != 0 {
		color.Red("++ ERROR: Something went wrong executing command: %s.", command)
		return "Unknown", fmt.Errorf("Error: Something went wrong executing command: %s.", command)
	}
	//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())

	scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		//fmt.Println("LINE: ", line)
		if len(line) == 0 {
			continue
		}
		if strings.Contains(line, "SN = ") {
			serialNumberData := strings.Split(line, "SN = ")
			serialNumber := serialNumberData[1]
			//fmt.Println("serialNumber: ", serialNumber)
			return serialNumber, nil
		}
	}
	return "Unknown", nil
}

var CheckMegaraidPerc = func(manufacturer string) (bool, error) {
	//fmt.Println("--- CheckMegaraidPerc ---")
	// Execute storcli/perccli
	if manufacturer == "perc" {
		fmt.Println("> Checking PERC controller.")
	} else {
		fmt.Println("> Checking MegaRaid controller.")
	}
	command := "/call show all"
	outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "CheckMegaraidPerc", command)
	if err != nil {
		//color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
		return false, fmt.Errorf("Something went wrong executing command %s: %v", command, err)
	}
	if len(outputStderr.String()) != 0 {
		//color.Red("++ ERROR: Something went wrong executing command: %s.", command)
		return false, fmt.Errorf("Something went wrong executing command: %s.", command)
	}
	//fmt.Println("Stdout:", outputStdout.String(), "Stderr:", outputStderr.String())

	scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
	statusFailure := true
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		//fmt.Println(line)
		// Check first Status output line
		if strings.Contains(line, "Status = ") {
			statusData := strings.Split(line, " = ")
			status := statusData[1]
			//fmt.Println("status: ", status)
			if status == "Success" {
				statusFailure = false
			} else {
				statusFailure = true
			}
			continue
		}

		if strings.Contains(line, "Description = ") {
			descriptionData := strings.Split(line, " = ")
			description := descriptionData[1]
			//fmt.Println("description: ", description)
			capManufacturer := strings.ToUpper(manufacturer)
			if statusFailure && description == "No Controller found" {
				fmt.Printf("> No %v RAID controller detected.\n", capManufacturer)
				return false, nil
			}
			if !statusFailure && description == "None" {
				color.Magenta("> %v RAID controller detected.", capManufacturer)
				return true, nil
			}
		}
	}
	return false, nil
}

var ProcessHWMegaraidPercRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.RaidStruct, []utils.NoRaidDiskStruct, error) {
	//fmt.Println("-- processHWMegaraidPercRaid --")
	var controllers = []utils.ControllerStruct{}
	var raids = []utils.RaidStruct{}
	var noRaidDisks = []utils.NoRaidDiskStruct{}

	// Execute storcli/perccli
	fmt.Println("> Getting current Mega-RAID configuration.")
	command := "/call show all"
	outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "processHWMegaraidPercRaid", command)
	if err != nil {
		color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
		return controllers, raids, noRaidDisks, fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
	}
	if len(outputStderr.String()) != 0 {
		color.Red("++ ERROR: Something went wrong executing command: %s.", command)
		return controllers, raids, noRaidDisks, fmt.Errorf("Error: Something went wrong executing command: %s.", command)
	}
	//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())

	fmt.Println("> Parsing Mega-RAID data.")
	scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))

	insideTopologyList := false
	topologyDataLine := false

	insidePhysicalList := false
	physicalDataLine := false

	tableBarSeparator := ""
	tableBarSeparatorCounter := 0

	// Initialize controller variables
	var controllerId string
	var controllerModel string
	var controllerStatus string

	// Initialize raid variables
	raidLevel := 0
	var raid = utils.RaidStruct{}

	// disks variable to save it to raid
	var disks = []utils.DiskStruct{}
	isCacDrive := false

	wasPreviousLineRaid := false
	wasPreviousLineDisk := false
	wasPreviousLineCac := false
	previousControllerId := ""
	previousTopologyDG := ""
	previoustopologyType := ""

	lineNumber := 1
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		//fmt.Printf("Line: %s\n", line)
		//fmt.Printf("Line: %d ControllerID: %s\n", lineNumber, controllerId)

		// CONTROLLER
		// Controller ID:
		if strings.Contains(line, "Controller = ") && !strings.Contains(line, "Temperature Sensor for Controller") {
			controllerIdData := strings.Split(line, "Controller = ")
			controllerId = controllerIdData[1]
			//fmt.Println("ControllerID: ", controllerId)
			continue
		}

		// Model:
		if strings.Contains(line, "Model = ") {
			controllerModelData := strings.Split(line, "Model = ")
			controllerModel = controllerModelData[1]
			//fmt.Println("Controller Model: ", controllerModel)
			continue
		}

		// Status:
		if strings.Contains(line, "Controller Status = ") {
			controllerStatusData := strings.Split(line, "Controller Status = ")
			controllerStatus = controllerStatusData[1]
			//fmt.Println("Controller Status: ", controllerStatus)
			continue
		}

		// Save controller data only when all fields have been already parsed
		if len(controllerId) > 0 && len(controllerModel) > 0 && len(controllerStatus) > 0 {
			controller := utils.ControllerStruct{
				Id:           manufacturer + "-" + controllerId,
				Manufacturer: manufacturer,
				Model:        controllerModel,
				Status:       controllerStatus,
			}
			controllers = append(controllers, controller)
			// Reset all controller variables since controller is already appended to controllers array
			// Except controllerId variable that is used in other code parts
			controllerModel = ""
			controllerStatus = ""
		}

		// TOPOLOGY
		// Parse Raid topology
		if strings.Contains(line, "TOPOLOGY :") {
			insideTopologyList = true
			//fmt.Println("TOPOLOGY line detected")
			continue
		}

		// tableBarSeparator is dynamic so we have to gues in each server(because separator num of chars depends on showed data that is variable in each server) and section
		if strings.Contains(line, "------------------") && len(line) > 30 {
			tableBarSeparator = line
		}

		// tableBarSeparator detected, increase tableBarSeparatorCounter
		if insideTopologyList && line == tableBarSeparator {
			tableBarSeparatorCounter++
		}

		// We are one line above first topologyDataLine
		if insideTopologyList && tableBarSeparatorCounter == 2 && topologyDataLine == false {
			topologyDataLine = true
			continue
		}

		// We are one line behind last topologyDataLine
		// Disks should be already added to last raid object so add it to raids array
		if tableBarSeparatorCounter == 3 && wasPreviousLineDisk {
			//fmt.Println("Last line detected")
			if !isCacDrive {
				for _, disk := range disks {
					raid.AddDisk(disk)
				}
				//fmt.Println("Disk info transfered to disk")
				raids = append(raids, raid)
				//fmt.Println("Previous raid saved")
				disks = nil
			}
			tableBarSeparatorCounter = 0
			wasPreviousLineDisk = false
			insideTopologyList = false
			topologyDataLine = false
			continue
		}

		// Middle topologyDataLines
		if topologyDataLine && tableBarSeparatorCounter == 2 {
			//fmt.Println("wasPreviousLineRaid: ", wasPreviousLineRaid)
			//fmt.Println("topologyDataLine: ", line)
			topologyDG := strings.Fields(line)[0]
			topologyEIDSlot := strings.Fields(line)[3]
			topologyType := strings.Fields(line)[5]
			topologyState := strings.Fields(line)[6]
			topologySize := strings.Fields(line)[8]
			topologySizeUnit := strings.Fields(line)[9]
			//fmt.Printf("  DG: %s %s(%s) %s %s %s\n", topologyDG, topologyType, topologyEIDSlot, topologyState, topologySize, topologySizeUnit)

			// RAIDs
			// Save raid data only when all fields have been already parsed
			if len(controllerId) > 0 && len(topologyDG) > 0 && len(topologyType) > 0 && len(topologyState) > 0 && len(topologySize) > 0 && len(topologySizeUnit) > 0 && strings.Contains(topologyType, "RAID") {
				// If its a RAID0 of Cac unit, skip it
				if wasPreviousLineCac {
					//fmt.Println("wasPreviousLineCac:", wasPreviousLineCac)
					wasPreviousLineCac = false
					// mark next drive as cacheCade
					isCacDrive = true
					//fmt.Println("Next drive marked as CAC")
					continue
				}

				// If we find two RAID identical lines consecutively, skip it
				if controllerId == previousControllerId && topologyDG == previousTopologyDG && topologyType == previoustopologyType && wasPreviousLineRaid {
					//fmt.Println("Same RAID, skipping")
					continue
				}

				// Second level hirarchy raid detected, alll fields match except topologyType, save previous raid
				if controllerId == previousControllerId && topologyDG == previousTopologyDG && wasPreviousLineRaid {
					//fmt.Println("Second level hirarchy raid detected")
					raids = append(raids, raid)
					//fmt.Println("Previous raid saved")
					raidLevel = 1
				}

				// Current RAID line detected, and previous was DRIVE, so transfer disks info to previous RAID and save it
				if wasPreviousLineDisk && !isCacDrive {
					for _, disk := range disks {
						raid.AddDisk(disk)
					}
					disks = nil
					//fmt.Println("Disk info transfered to disk")
					raids = append(raids, raid)
					//fmt.Println("Previous raid saved")
					wasPreviousLineDisk = false
				}

				// Get RAID OS device
				osDevice, err := hardwarecontrollerscommon.GetRaidOSDevice(manufacturer, controllerId, topologyDG)
				if err != nil {
					color.Red("++ ERROR Getting OS device: %s", err)
					return controllers, raids, noRaidDisks, err
				}

				finalTopologySize := strings.Join([]string{topologySize, topologySizeUnit}, " ")
				//fmt.Println("RaidSize: ", finalTopologySize)
				raid = utils.RaidStruct{
					ControllerId: manufacturer + "-" + controllerId,
					RaidLevel:    raidLevel,
					Dg:           topologyDG,
					RaidType:     topologyType,
					State:        topologyState,
					Size:         finalTopologySize,
					OsDevice:     osDevice,
				}
				//fmt.Println("Raid instance created")

				wasPreviousLineRaid = true
				wasPreviousLineDisk = false

				previousControllerId = controllerId
				previousTopologyDG = topologyDG
				previoustopologyType = topologyType
				continue
			}

			// DRIVEs
			// Save disk  only when all fields have been already parsed
			if len(controllerId) > 0 && len(topologyDG) > 0 && len(topologyEIDSlot) > 0 && len(topologyType) > 0 && len(topologyState) > 0 && len(topologySize) > 0 && len(topologySizeUnit) > 0 && topologyType == "DRIVE" {
				finalTopologySize := strings.Join([]string{topologySize, topologySizeUnit}, " ")

				// Get serial number
				serialNumber, err := GetMegaraidPercDriveSerialNumber(manufacturer, controllerId, topologyEIDSlot)
				//fmt.Println("serialNumber: ", serialNumber)
				if err != nil {
					color.Red("++ ERROR Getting drive serial number: %s", err)
					return controllers, raids, noRaidDisks, err
				}

				if isCacDrive {
					noRaidDisk := utils.NoRaidDiskStruct{
						ControllerId: manufacturer + "-" + controllerId,
						EidSlot:      topologyEIDSlot,
						State:        topologyState,
						Size:         finalTopologySize,
						SerialNumber: serialNumber,
						OsDevice:     "CacheCade",
					}
					noRaidDisks = append(noRaidDisks, noRaidDisk)
					//fmt.Println("CAC drive saved")
				} else {
					disk := utils.DiskStruct{
						ControllerId: manufacturer + "-" + controllerId,
						Dg:           topologyDG,
						EidSlot:      topologyEIDSlot,
						State:        topologyState,
						Size:         finalTopologySize,
						SerialNumber: serialNumber,
					}
					disks = append(disks, disk)
					//fmt.Println("Regular disk instance added to array")
				}
				wasPreviousLineDisk = true
				wasPreviousLineRaid = false
				continue
			}

			// CacheCade drives
			if len(controllerId) > 0 && len(topologyDG) > 0 && len(topologyEIDSlot) > 0 && len(topologyType) > 0 && len(topologyState) > 0 && len(topologySize) > 0 && len(topologySizeUnit) > 0 && strings.Contains(topologyType, "Cac") {
				//fmt.Println("CAC drive detected")
				// Current CAC line detected, and previous was DRIVE, so transfer disks info to previous RAID and save it
				if wasPreviousLineDisk {
					for _, disk := range disks {
						raid.AddDisk(disk)
					}
					disks = nil
					//fmt.Println("Disk info transfered to disk")
					raids = append(raids, raid)
					//fmt.Println("Previous raid saved")
				}
				wasPreviousLineCac = true
				continue
			}
		}

		// PHYSICAL DRIVES
		//fmt.Println("line number: ", lineNumber)
		if strings.Contains(line, "PD LIST :") {
			//fmt.Println("INSIDE PHYSICAL line number: ", lineNumber)
			// Disable previous section variables only to be sure
			topologyDataLine = false
			insideTopologyList = false

			insidePhysicalList = true
			tableBarSeparatorCounter = 0
		}

		// tableBarSeparator is dynamic so we have to gues in each server and section
		if strings.Contains(line, "------------------") && len(line) > 30 {
			tableBarSeparator = line
		}

		if insidePhysicalList && line == tableBarSeparator {
			tableBarSeparatorCounter++
			//fmt.Println("tableBarSeparatorCounter: ", tableBarSeparatorCounter)
		}

		if insidePhysicalList && tableBarSeparatorCounter == 2 && physicalDataLine == false {
			physicalDataLine = true
			//fmt.Println("physicalDataLine: ", physicalDataLine)
			continue
		}

		if physicalDataLine == true && tableBarSeparatorCounter == 2 {
			//fmt.Println("------------------------------")
			//fmt.Println("physicalDataLine: ", line)
			physicalEidSlot := strings.Fields(line)[0]
			physicalState := strings.Fields(line)[2]
			physicalSize := strings.Fields(line)[4]
			physicalSizeUnit := strings.Fields(line)[5]
			finalphysicalSize := strings.Join([]string{physicalSize, physicalSizeUnit}, " ")
			physicalIntf := strings.Fields(line)[6]
			physicalMedium := strings.Fields(line)[7]
			//fmt.Println("len(strings.Fields(line)): ", len(strings.Fields(line)))

			physicalModel := ""
			physicalModelStartField := 11
			//fmt.Println("physicalModelStartField: ", physicalModelStartField)
			physicalModelEndField := len(strings.Fields(line)) - 2
			//fmt.Println("physicalModelEndField: ", physicalModelEndField)

			// Some models lacks Type column
			if physicalModelEndField <= physicalModelStartField {
				physicalModelEndField = len(strings.Fields(line)) - 1
			}

			// Build drive model string
			for i := physicalModelStartField; i < physicalModelEndField; i++ {
				physicalModelField := strings.Fields(line)[i]
				//fmt.Printf("physicalModel: |%s|\n", physicalModel)
				//fmt.Printf("physicalModelField: |%s|\n", physicalModelField)
				// First physicalModelField, dont prepend a space " "
				if i == physicalModelStartField {
					physicalModel = strings.Join([]string{physicalModel, physicalModelField}, "")
				} else {
					physicalModel = strings.Join([]string{physicalModel, physicalModelField}, " ")
				}
			}

			//fmt.Println("physicalEidSlot: ", physicalEidSlot)
			//fmt.Println("physicalIntf: ", physicalIntf)
			//fmt.Println("physicalMedium: ", physicalMedium)
			//fmt.Println("physicalModel: ", physicalModel)

			// Add extra drive information and check if physical drive was previously seen on VolumeGroup section
			eidSlotFound := false
			for _, raid := range raids {
				for i := range raid.Disks {
					// range always copies variable values by copy
					// Its compulsory to alter original values making it by ref
					disk := &raid.Disks[i]
					//fmt.Printf("Checking %s VS %s\n", physicalEidSlot, disk.eidSlot)
					if physicalEidSlot == disk.EidSlot {
						disk.Intf = physicalIntf
						disk.Medium = physicalMedium
						disk.Model = physicalModel
						eidSlotFound = true
						//fmt.Printf("Disk with eidSlot: %s updated\n", physicalEidSlot)
						break
					}
				}
				if eidSlotFound {
					break
				}
			}

			// Also check if disk was previously seen on VolumeGroup section as CAC disk
			for i := range noRaidDisks {
				// range always copies variable values by copy
				// Its compulsory to alter original values making it by ref
				noRaidDisk := &noRaidDisks[i]
				//fmt.Printf("Checking %s VS %s\n", physicalEidSlot, noRaidDisk.eidSlot)
				if physicalEidSlot == noRaidDisk.EidSlot {
					noRaidDisk.Intf = physicalIntf
					noRaidDisk.Medium = physicalMedium
					noRaidDisk.Model = physicalModel
					eidSlotFound = true
					//fmt.Printf("noRaidDisk with eidSlot: %s updated\n", physicalEidSlot)
					break
				}
			}

			// No raid disk detected
			if !eidSlotFound {
				//fmt.Printf("Disk with eidSlot: %s not found in any raid\n", physicalEidSlot)
				// Get serial number
				serialNumber, err := GetMegaraidPercDriveSerialNumber(manufacturer, controllerId, physicalEidSlot)
				if err != nil {
					color.Red("++ ERROR Getting drive serial number: %s", err)
					return controllers, raids, noRaidDisks, err
				}

				// MegaRaid/PERC JBOD query
				osDevice, err := hardwarecontrollerscommon.GetJbodOsDevice(manufacturer, controllerId, physicalEidSlot)
				if err != nil {
					color.Red("++ ERROR Getting OS device: %s", err)
					return controllers, raids, noRaidDisks, err
				}
				// If disk hardware is bogus, change disk state
				//fmt.Println("osDevice: ", osDevice)
				if osDevice == "BogusDisk-OSUnknown" {
					physicalState = "BogusDisk"
				}
				osDevice = "JBOD-" + osDevice
				// We use physical data instead of topology data
				noRaidDisk := utils.NoRaidDiskStruct{
					ControllerId: manufacturer + "-" + controllerId,
					EidSlot:      physicalEidSlot,
					State:        physicalState,
					Size:         finalphysicalSize,
					Intf:         physicalIntf,
					Medium:       physicalMedium,
					Model:        physicalModel,
					SerialNumber: serialNumber,
					OsDevice:     osDevice,
				}
				noRaidDisks = append(noRaidDisks, noRaidDisk)
				//fmt.Printf("Disk with eidSlot: %s added to noRaidDisks array\n", physicalEidSlot)
			}
		}
		//fmt.Println("lineNumber: ", lineNumber)
		lineNumber++
	}

	//fmt.Println("> Done.")

	//fmt.Println("len controllers", len(controllers))
	//fmt.Println("len raids", len(raids))
	// fmt.Println("----- controllers -----")
	// spew.Dump(controllers)
	// fmt.Println("----- raids -----")
	// spew.Dump(raids)
	//fmt.Println("----- noRaidDisks -----")
	//spew.Dump(noRaidDisks)
	return controllers, raids, noRaidDisks, nil
}
