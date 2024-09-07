package megaraidpercsas2ircu

import (
	"bufio"
	"fmt"
	"hardwareAnalyzer/hardwarecontrollerscommon"
	"hardwareAnalyzer/utils"
	"strconv"
	"strings"

	human "github.com/dustin/go-humanize"

	"github.com/fatih/color"
)

// Hardware controller functions:
// Megaraid: Pure MegaRAID controller
// Dell(Megaraid): PERC and SAS2IRCU controllers

var CheckSas2ircuRaid = func() (bool, error) {
	//fmt.Println("-- checkSas2ircuRaid --")

	fmt.Println("> Checking SAS2IRCU RAID controller.")
	command := "LIST"
	outputStdout, outputStderr, err := utils.GetCommandOutput("sas2ircu", "checkSas2ircuRaid", command)
	if err != nil {
		// When theres no SAS2IRCU controller in the system, "exit status 1" is returned, its not an error
		if err.Error() != "exit status 1" {
			//color.Red("++ ERROR: Something went wrong executing command %s: %v", command, err)
			return false, fmt.Errorf("Something went wrong executing command %s: %v", command, err)
		} else {
			fmt.Println("> No SAS2IRCU RAID controller detected.")
			return false, nil
		}
	}
	if len(outputStderr.String()) != 0 {
		//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())
		// When theres no sas2ircu controller installed in the system, tool returns: SAS2IRCU: MPTLib2 Error 1
		if strings.Contains(outputStdout.String(), "MPTLib2 Error 1") {
			fmt.Println("> No SAS2IRCU RAID controller detected.")
			return false, nil
		} else {
			//color.Red("++ ERROR: Something went wrong executing command: %s.", command)
			return false, fmt.Errorf("Something went wrong executing command: %s.", command)
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		//fmt.Println(line)
		// Check first Status output line
		if strings.Contains(line, "SAS2IRCU: ") {
			statusData := strings.Split(line, ": ")
			status := statusData[1]
			//fmt.Println("status: ", status)
			if status == "Utility Completed Successfully." {
				color.Magenta("> SAS2IRCU RAID controller detected.")
				return true, nil
			} else {
				fmt.Println("> No SAS2IRCU RAID controller detected.")
				return false, nil
			}
		}
	}
	return false, nil
}

var ProcessHWSas2ircuRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.RaidStruct, []utils.NoRaidDiskStruct, error) {
	//fmt.Println("-- processHWSas2ircuRaid --")
	var controllers = []utils.ControllerStruct{}
	var raids = []utils.RaidStruct{}
	var noRaidDisks = []utils.NoRaidDiskStruct{}

	// Execute sas2ircu
	fmt.Println("> Getting current SAS2IRCU-RAID configuration.")
	command := "LIST"
	outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "processHWSas2ircuRaid", command)
	if err != nil {
		color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
		return controllers, raids, noRaidDisks, fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
	}
	if len(outputStderr.String()) != 0 {
		color.Red("++ ERROR: Something went wrong executing command: %s.", command)
		return controllers, raids, noRaidDisks, fmt.Errorf("Error: Something went wrong executing command: %s.", command)
	}
	//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())

	// Get controller IDs
	scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
	totalSas2ircuControllers := ""
	previousLine := ""
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		//fmt.Println("line: ", line)
		if line == "SAS2IRCU: Utility Completed Successfully." {
			totalSas2ircuControllers = strings.Fields(previousLine)[0]
			totalSas2ircuControllers = utils.ClearString(totalSas2ircuControllers)
		}
		previousLine = line
	}

	//fmt.Printf("totalSas2ircuControllers: |%s|\n", totalSas2ircuControllers)
	n, err := strconv.Atoi(totalSas2ircuControllers)
	//fmt.Printf("n: |%d|\n", n)

	// Initialize raid variables
	var raid = utils.RaidStruct{}

	var volumeId string
	var raidType string
	var raidState string
	var raidSize string
	var osDevice string

	// disks variable to save it to raid
	var disks = []utils.DiskStruct{}
	diskEID := ""
	diskSlot := ""
	diskState := ""
	diskSize := ""
	diskInterface := ""
	diskMedium := ""
	diskModel := ""
	diskSerialNumber := ""

	insideVolumeData := false
	insideDriveData := false
	tableBarSeparator := ""

	// Parse sas2ircu controller data
	for i := 0; i <= n; i++ {
		controllerId := strconv.Itoa(i)
		//fmt.Printf("> Getting sas2ircu controller %s data.\n", controllerId)

		fmt.Println("> Parsing SAS2IRCU data.")
		command := strconv.Itoa(i) + " DISPLAY"
		outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "processHWSas2ircuRaid", command)
		if err != nil {
			color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
			return controllers, raids, noRaidDisks, fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
		}
		if len(outputStderr.String()) != 0 {
			color.Red("++ ERROR: Something went wrong executing command: %s.", command)
			return controllers, raids, noRaidDisks, fmt.Errorf("Error: Something went wrong executing command: %s.", command)
		}
		//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())

		scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
		controllerStatus := "Good"
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			//fmt.Println("line: ", line)

			// tableBarSeparator is dynamic so we have to gues in each server and section
			if strings.Contains(line, "------------------") && len(line) > 30 {
				tableBarSeparator = line
			}

			// CONTROLLER
			// Controller ID:
			if strings.Contains(line, "Controller type") {
				controllerModelData := strings.Split(line, ":")
				controllerModel := controllerModelData[1]
				controllerModel = utils.ClearString(controllerModel)
				//fmt.Println("Controller Model: ", controllerModel)
				controller := utils.ControllerStruct{
					Id:           manufacturer + "-" + controllerId,
					Manufacturer: manufacturer,
					Model:        controllerModel,
					Status:       controllerStatus,
				}
				controllers = append(controllers, controller)
				//fmt.Println("sas2ircu controller appended.")
				continue
			}

			// Data section indicators
			if strings.Contains(line, "IR volume") {
				insideVolumeData = true
				//fmt.Println("--- insideVolumeData: ", insideVolumeData)
				continue
			}
			if strings.Contains(line, "Device is a Hard disk") {
				insideDriveData = true
				//fmt.Println("--- insideDriveData: ", insideDriveData)
			}

			// 2 Volumes example output: https://jp.fujitsu.com/platform/server/sparc/manual/en/c120-e679/f.3.html
			// Volume information
			if insideVolumeData {
				// Volume ID
				if strings.Contains(line, "Volume ID") {
					volumeIdData := strings.Split(line, ": ")
					volumeId = volumeIdData[1]
					volumeId = utils.ClearString(volumeId)
					//fmt.Println("volumeId: ", volumeId)
					continue
				}
				// Volume state
				if strings.Contains(line, "Status of volume") {
					raidStateData := strings.Split(line, ":")
					raidState = raidStateData[1]
					raidState = utils.ClearString(raidState)
					//fmt.Println("raidState: ", raidState)
					continue
				}
				// Get RAID OS device
				if strings.Contains(line, "Volume wwid") {
					wwidData := strings.Split(line, ":")
					wwid := wwidData[1]
					wwid = utils.ClearString(wwid)
					osDevice, err = hardwarecontrollerscommon.GetRaidOSDevice("sas2ircu", controllerId, wwid)
					if err != nil {
						color.Red("++ ERROR Getting OS device: %s", err)
						return controllers, raids, noRaidDisks, err
					}
					//fmt.Println("osDevice: ", osDevice)
					continue
				}
				// RAID type
				if strings.Contains(line, "RAID level") {
					raidTypeData := strings.Split(line, ":")
					raidType = raidTypeData[1]
					raidType = utils.ClearString(raidType)
					//fmt.Println("raidType: ", raidType)
					continue
				}
				// RAID size
				if strings.Contains(line, "Size (in MB)") {
					raidSizeData := strings.Split(line, ":")
					raidSize = raidSizeData[1]
					//fmt.Println("raidSize: ", raidSize)
					raidSize = utils.ClearString(raidSize)
					raidSizeInt, _ := utils.Str2uint64(raidSize)
					raidSizeInt = raidSizeInt * 1024 * 1024
					raidSize = human.Bytes(raidSizeInt)
					//fmt.Println("raidSize: ", raidSize)
					continue
				}
				// RAID drives
				if strings.Contains(line, "PHY") {
					raidEIDSlotData := strings.Split(line, ":")
					raidEID := raidEIDSlotData[1]
					raidEID = utils.ClearString(raidEID)
					raidSlot := raidEIDSlotData[2]
					raidSlot = utils.ClearString(raidSlot)
					raidEIDSlot := raidEID + ":" + raidSlot
					//fmt.Println("controllerId: ", controllerId)
					//fmt.Println("dg: ", volumeId)
					//fmt.Println("raidEIDSlot: ", raidEIDSlot)
					disk := utils.DiskStruct{
						ControllerId: manufacturer + "-" + controllerId,
						Dg:           volumeId,
						EidSlot:      raidEIDSlot,
					}

					//fmt.Println("Raid disks: ", len(raid.disks))
					// Save disks info to array in order to save it to transfer it to raid
					//fmt.Println("disk: ", disk)
					disks = append(disks, disk)
					continue
				}

				// Another volume detected or End of volume data, save RAID to raids array
				if insideVolumeData && (strings.Contains(line, "IR volume") || line == tableBarSeparator) {
					// sas2ircu only shows plain herarchy raids, all of them are raidLevel: 0
					raid = utils.RaidStruct{
						ControllerId: manufacturer + "-" + controllerId,
						RaidLevel:    0,
						Dg:           volumeId,
						RaidType:     raidType,
						State:        raidState,
						Size:         raidSize,
						Disks:        disks,
						OsDevice:     osDevice,
					}
					//fmt.Println("raid: ", raid)
					raids = append(raids, raid)
					//fmt.Printf("Adding RAID controllerId: %s raidLevel: 0 dg: %s raidType: %s raidState: %s raidSize: %s disks: %s osDevice: %s|\n", controllerId, volumeId, raidType, raidState, raidSize, disks, osDevice)
				}
				if insideVolumeData && line == tableBarSeparator {
					insideVolumeData = false
				}
			}

			// Drives information
			if insideDriveData {
				if strings.Contains(line, "Enclosure #") {
					diskEIDData := strings.Split(line, ":")
					diskEID = diskEIDData[1]
					diskEID = utils.ClearString(diskEID)
					//fmt.Println("diskEID: ", diskEID)
					continue
				}
				if strings.Contains(line, "Slot #") {
					diskSlotData := strings.Split(line, ":")
					diskSlot = diskSlotData[1]
					diskSlot = utils.ClearString(diskSlot)
					//fmt.Println("diskSlot: ", diskSlot)
					continue
				}

				if strings.Contains(line, "State") {
					diskStateData := strings.Split(line, ":")
					diskState = diskStateData[1]
					diskState = utils.ClearString(diskState)
					//fmt.Println("diskState: ", diskState)
					continue
				}
				if strings.Contains(line, "Size (in MB)/(in sectors)") {
					diskSizeData := strings.Split(line, ":")
					diskSize = strings.Split(diskSizeData[1], "/")[0]
					diskSize = utils.ClearString(diskSize)
					//fmt.Println("diskSize: ", diskSize)
					diskSizeInt, _ := utils.Str2uint64(diskSize)
					// human.Bytes function expect Bytes: MB -> Bytes
					diskSizeInt = diskSizeInt * 1024 * 1024
					//fmt.Println("diskSizeInt: ", diskSizeInt)
					diskSize = human.Bytes(diskSizeInt)
					//fmt.Println("diskSize: ", diskSize)
					continue
				}
				if strings.Contains(line, "Model Number") {
					diskModelData := strings.Split(line, ":")
					diskModel = diskModelData[1]
					diskModel = utils.ClearString(diskModel)
					// Special fake disk
					if diskModel == "BACKPLANE" {
						//fmt.Println("BACKPLANE detected")
						continue
					}
					//fmt.Println("diskModel: ", diskModel)
					continue
				}
				if strings.Contains(line, "Serial No") {
					diskSerialNumberData := strings.Split(line, ":")
					diskSerialNumber = diskSerialNumberData[1]
					diskSerialNumber = utils.ClearString(diskSerialNumber)
					//fmt.Println("diskSerialNumber: ", diskSerialNumber)
					continue
				}
				if strings.Contains(line, "Protocol") {
					diskInterfaceData := strings.Split(line, ":")
					diskInterface = diskInterfaceData[1]
					diskInterface = utils.ClearString(diskInterface)
					//fmt.Println("diskInterface: ", diskInterface)
					continue
				}
				if strings.Contains(line, "Drive Type") {
					diskMediumData := strings.Split(line, ":")
					diskMedium = diskMediumData[1]
					diskMedium = utils.ClearString(diskMedium)
					//fmt.Println("diskMedium: ", diskMedium)
					continue
				}

				// Update raid disks with extra information, also check if drive was previously seen as any RAID drive
				if len(diskState) > 0 && len(diskSize) > 0 && len(diskInterface) > 0 && len(diskMedium) > 0 && len(diskModel) > 0 && len(diskSerialNumber) > 0 {
					eidSlot := diskEID + ":" + diskSlot
					//fmt.Println("eidSlot: ", eidSlot)
					eidSlotFound := false
					for _, raid := range raids {
						for i := range raid.Disks {
							// range always copies variable values by copy
							// Its compulsory to alter original values making it by ref
							disk := &raid.Disks[i]
							//fmt.Printf("Checking device list: %s VS raid list: %s\n", eidSlot, disk.eidSlot)
							// Calling GetMegaraidPercDriveSerialNumber() as made with storcli is not required as
							// sas2ircu shows that data directly without need of executing a second command
							if eidSlot == disk.EidSlot {
								//fmt.Println("Raid disk detected")
								disk.State = diskState
								disk.Size = diskSize
								disk.Intf = diskInterface
								disk.Medium = diskMedium
								disk.Model = diskModel
								disk.SerialNumber = diskSerialNumber
								//fmt.Printf("diskState: %s diskSize: %s diskInterface: %s diskMedium: %s diskModel: %s diskSerialNumber: %s|\n", diskState, diskSize, diskInterface, diskMedium, diskModel, diskSerialNumber)
								eidSlotFound = true
								// Reset all drive data
								diskEID = ""
								diskSlot = ""
								diskState = ""
								diskSize = ""
								diskInterface = ""
								diskMedium = ""
								diskModel = ""
								diskSerialNumber = ""
								break
							}
						}
						if eidSlotFound {
							break
						}
					}

					if !eidSlotFound {
						//fmt.Println("NO Raid disk detected")
						// SAS2IRCU JBOD query
						osDevice, err := hardwarecontrollerscommon.GetJbodOsDevice(manufacturer, controllerId, eidSlot)
						if err != nil {
							color.Red("++ ERROR Getting OS device: %s", err)
							return controllers, raids, noRaidDisks, err
						}
						osDevice = "JBOD-" + osDevice
						noRaidDisk := utils.NoRaidDiskStruct{
							ControllerId: manufacturer + "-" + controllerId,
							EidSlot:      eidSlot,
							State:        diskState,
							Size:         diskSize,
							Intf:         diskInterface,
							Medium:       diskMedium,
							Model:        diskModel,
							SerialNumber: diskSerialNumber,
							OsDevice:     osDevice,
						}
						noRaidDisks = append(noRaidDisks, noRaidDisk)
						//fmt.Printf("Disk with eidSlot: %v added to noRaidDisks array: %v\n", eidSlot, osDevice)
						// Reset all drive data
						diskEID = ""
						diskSlot = ""
						diskState = ""
						diskSize = ""
						diskInterface = ""
						diskMedium = ""
						diskModel = ""
						diskSerialNumber = ""
					}
				}
			}
			// End of disks data, save RAID to raids array
			if insideDriveData && line == tableBarSeparator {
				insideDriveData = false
				break
			}
		}
	}
	return controllers, raids, noRaidDisks, nil
}
