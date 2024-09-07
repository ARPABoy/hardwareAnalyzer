package adaptec

import (
	"bufio"
	"fmt"
	"hardwareAnalyzer/hardwarecontrollerscommon"
	"hardwareAnalyzer/utils"
	"math"
	"regexp"
	"strconv"
	"strings"

	human "github.com/dustin/go-humanize"

	"github.com/fatih/color"
)

var CheckAadaptecRaid = func() (bool, error) {
	fmt.Println("> Checking ADAPTEC RAID controller.")
	command := "LIST"
	outputStdout, outputStderr, err := utils.GetCommandOutput("adaptec", "checkAadaptecRaid", command)
	//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())
	if err != nil {
		//color.Red("++ ERROR: Something went wrong executing command %s: %v", command, err)
		return false, fmt.Errorf("Something went wrong executing command %s: %v", command, err)
	}
	if len(outputStderr.String()) != 0 {
		//color.Red("++ ERROR: Something went wrong executing command: %s.", command)
		return false, fmt.Errorf("Something went wrong executing command: %s.", command)
	}

	scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		//fmt.Println("line: ", line)
		if strings.Contains(line, "Controllers found: ") {
			controllersFoundData := strings.Split(line, ": ")
			controllersFound := controllersFoundData[1]
			if controllersFound == "0" {
				fmt.Println("> No ADAPTEC RAID controllers detected.")
				return false, nil
			} else {
				color.Magenta("> ADAPTEC RAID controller detected.")
				return true, nil
			}
		}
	}
	return false, nil
}

var ProcessHWAdaptecRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.RaidStruct, []utils.NoRaidDiskStruct, error) {
	var controllers = []utils.ControllerStruct{}
	var raids = []utils.RaidStruct{}
	var noRaidDisks = []utils.NoRaidDiskStruct{}

	// Execute arcconf
	fmt.Println("> Getting current Adaptec-RAID configuration.")
	command := "LIST"
	outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "processHWAdaptecRaid", command)
	//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())
	if err != nil {
		color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
		return controllers, raids, noRaidDisks, fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
	}
	if len(outputStderr.String()) != 0 {
		color.Red("++ ERROR: Something went wrong executing command: %s.", command)
		return controllers, raids, noRaidDisks, fmt.Errorf("Error: Something went wrong executing command: %s.", command)
	}

	scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
	// Initialize controller variables
	controllerModel := ""
	controllerStatus := ""
	totalAdaptecControllers := ""

	// Initialize raid variables
	//raidLevel := 0
	var raid = utils.RaidStruct{}

	firstLogicalDevice := true
	logicalDeviceNumber := ""
	logicalDeviceRaidLevel := ""
	logicalDeviceName := ""
	logicalDeviceStatus := ""
	logicalDeviceSize := ""
	logicalDeviceUniqueIdentifier := ""
	osDevice := ""

	// No raid disks
	var disks = []utils.DiskStruct{}
	diskSerialNumber := ""

	// physicalDevice variables
	physicalDeviceState := ""
	physicalDeviceEsd := ""
	physicalDeviceModel := ""
	physicalDeviceSize := ""
	physicalDeviceInterface := ""
	physicalDeviceMedium := ""

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		//fmt.Println("Line: ", line)
		// Controller ID:
		if strings.Contains(line, "Controllers found:") {
			totalAdaptecControllersData := strings.Split(line, ":")
			totalAdaptecControllers = totalAdaptecControllersData[1]
			totalAdaptecControllers = utils.ClearString(totalAdaptecControllers)
			//fmt.Println("totalAdaptecControllers: ", totalAdaptecControllers)
			break
		}
	}
	// Adaptec controllers starts with ID 1
	n, err := strconv.Atoi(totalAdaptecControllers)

	// Adaptec controllers starts with ID 1, so last controller is n-1
	for i := 0; i <= n-1; i++ {
		controllerId := strconv.Itoa(i)
		// Adaptec controllers starts with ID 1, but all other controllers with 0
		// We generate an extra var only for arcconf command execution
		controllerIdArcconf := i + 1
		controllerIdArcconfString := strconv.Itoa(controllerIdArcconf)

		//fmt.Printf("> Getting adaptec controller %s data.\n", controllerId)

		fmt.Println("> Parsing Adaptec-RAID data.")
		command := "GETCONFIG " + controllerIdArcconfString + " AL"
		outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "processHWAdaptecRaid", command)
		//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())
		if err != nil {
			color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
			return controllers, raids, noRaidDisks, fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
		}
		if len(outputStderr.String()) != 0 {
			color.Red("++ ERROR: Something went wrong executing command: %s.", command)
			return controllers, raids, noRaidDisks, fmt.Errorf("Error: Something went wrong executing command: %s.", command)
		}

		scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			//fmt.Println("Line: ", line)

			// Controller ID:
			if strings.Contains(line, "Controller Status") {
				controllerStatusData := strings.Split(line, ":")
				controllerStatus = controllerStatusData[1]
				controllerStatus = utils.ClearString(controllerStatus)
				//fmt.Println("controllerStatus: ", controllerStatus)
				continue
			}

			// Controller model:
			if strings.Contains(line, "Controller Model") {
				controllerModelData := strings.Split(line, ":")
				// Controller model can have spaces
				firstControllerModelIteration := true
				for i := 0; i <= len(controllerModelData)-1; i++ {
					if firstControllerModelIteration {
						controllerModel = strings.Fields(controllerModelData[1])[i]
						firstControllerModelIteration = false
					} else {
						controllerModel = controllerModel + " " + strings.Fields(controllerModelData[1])[i]
					}
				}
				//fmt.Printf("controllerModel: |%s|\n", controllerModel)
				continue
			}

			// Save controller data
			if len(controllerId) > 0 && len(controllerStatus) > 0 && len(controllerModel) > 0 {
				controller := utils.ControllerStruct{
					Id:           manufacturer + "-" + controllerId,
					Manufacturer: manufacturer,
					Model:        controllerModel,
					Status:       controllerStatus,
				}
				controllers = append(controllers, controller)
				// Reset data except controllerId that is used after
				controllerModel = ""
				controllerStatus = ""
			}

			// Logical Device number:
			if strings.Contains(line, "Logical Device number") {
				if firstLogicalDevice {
					logicalDeviceNumber = strings.Fields(line)[3]
					logicalDeviceNumber = utils.ClearString(logicalDeviceNumber)
					//fmt.Println("logicalDeviceNumber: ", logicalDeviceNumber)
					firstLogicalDevice = false
				} else {
					// Current RAID line detected, so transfer pending disks info to previous RAID and save it
					for _, disk := range disks {
						raid.AddDisk(disk)
					}
					disks = nil
					//fmt.Println("Disk info transfered to disk")
					raids = append(raids, raid)
					//fmt.Println("raid.controllerId: ", raid.controllerId)
					//fmt.Println("Previous raid saved")
				}
				continue
			}

			// Last Logical Device detected
			if strings.Contains(line, "Physical Device information") {
				for _, disk := range disks {
					raid.AddDisk(disk)
				}
				disks = nil
				//fmt.Println("Disk info transfered to disk")
				raids = append(raids, raid)
				//fmt.Println("raid.controllerId: ", raid.controllerId)
				//fmt.Println("Previous raid saved")
				continue
			}

			// Logical Device name:
			if strings.Contains(line, "Logical Device name") {
				logicalDeviceNameData := strings.Split(line, ":")
				logicalDeviceName = logicalDeviceNameData[1]
				logicalDeviceName = utils.ClearString(logicalDeviceName)
				//fmt.Println("logicalDeviceName: ", logicalDeviceName)
				continue
			}

			// Logical Device raid level:
			if strings.Contains(line, "RAID level") {
				logicalDeviceRaidLevelData := strings.Split(line, ":")
				logicalDeviceRaidLevel = logicalDeviceRaidLevelData[1]
				logicalDeviceRaidLevel = "RAID" + utils.ClearString(logicalDeviceRaidLevel)
				//fmt.Println("logicalDeviceRaidLevel: ", logicalDeviceRaidLevel)
				continue
			}

			// Logical Device Unique Identifier:
			if strings.Contains(line, "Unique Identifier") {
				logicalDeviceUniqueIdentifierData := strings.Split(line, ":")
				logicalDeviceUniqueIdentifier = logicalDeviceUniqueIdentifierData[1]
				logicalDeviceUniqueIdentifier = utils.ClearString(logicalDeviceUniqueIdentifier)
				//fmt.Println("logicalDeviceUniqueIdentifier: ", logicalDeviceUniqueIdentifier)
				continue
			}

			// Logical Device status:
			if strings.Contains(line, "Status of Logical Device") {
				logicalDeviceStatusData := strings.Split(line, ":")
				logicalDeviceStatus = logicalDeviceStatusData[1]
				logicalDeviceStatus = utils.ClearString(logicalDeviceStatus)
				//fmt.Println("logicalDeviceStatus: ", logicalDeviceStatus)
				continue
			}

			// Logical Device size:
			if strings.Contains(line, "Size") && !strings.Contains(line, "Block Size of member drives") && !strings.Contains(line, "Total Size") {
				logicalDeviceSizeData := strings.Split(line, ":")
				logicalDeviceSize = logicalDeviceSizeData[1]

				// Remove alphas
				re := regexp.MustCompile(`\D`)
				logicalDeviceSizeValue := re.ReplaceAllString(logicalDeviceSize, "")
				logicalDeviceSizeValue = utils.ClearString(logicalDeviceSizeValue)
				//fmt.Println("logicalDeviceSizeValue", logicalDeviceSizeValue)

				// Remove digits
				re = regexp.MustCompile(`\d`)
				logicalDeviceSizeUnit := re.ReplaceAllString(logicalDeviceSize, "")
				logicalDeviceSizeUnit = utils.ClearString(logicalDeviceSizeUnit)
				//fmt.Println("logicalDeviceSizeUnit", logicalDeviceSizeUnit)

				multiplierFactor := 1
				switch logicalDeviceSizeUnit {
				case "KB":
					multiplierFactor = 1
				case "MB":
					multiplierFactor = 2
				case "GB":
					multiplierFactor = 3
				case "TB":
					multiplierFactor = 4
				case "PB":
					multiplierFactor = 5
				case "EB":
					multiplierFactor = 5
				case "ZB":
					multiplierFactor = 6
				}
				//fmt.Println("multiplierFactor: ", multiplierFactor)

				logicalDeviceDiskSizeFloat, _ := strconv.ParseFloat(logicalDeviceSizeValue, 64)
				//fmt.Println("logicalDeviceDiskSizeFloat: ", logicalDeviceDiskSizeFloat)
				logicalDeviceDiskSizeFloat = logicalDeviceDiskSizeFloat * math.Pow(float64(1024), float64(multiplierFactor))
				//fmt.Println("Bytes: ", logicalDeviceDiskSizeFloat)
				logicalDeviceSize = human.Bytes(uint64(logicalDeviceDiskSizeFloat))
				//fmt.Println("logicalDeviceSize: ", logicalDeviceSize)
				continue
			}

			//fmt.Printf("controllerId: %s logicalDeviceName: %s logicalDeviceUniqueIdentifier: %s logicalDeviceNumber: %s logicalDeviceRaidLevel: %s logicalDeviceStatus: %s logicalDeviceSize: %s\n", controllerId, logicalDeviceName, logicalDeviceUniqueIdentifier, logicalDeviceNumber, logicalDeviceRaidLevel, logicalDeviceStatus, logicalDeviceSize)
			// Create raid object
			if len(controllerId) > 0 && len(logicalDeviceName) > 0 && len(logicalDeviceUniqueIdentifier) > 0 && len(logicalDeviceNumber) > 0 && len(logicalDeviceRaidLevel) > 0 && len(logicalDeviceStatus) > 0 && len(logicalDeviceSize) > 0 {
				//fmt.Println("------ Creating RAID object")
				osDevice, err = hardwarecontrollerscommon.GetRaidOSDevice("adaptec", controllerId, logicalDeviceName+"_"+logicalDeviceUniqueIdentifier)
				if err != nil {
					color.Red("++ ERROR Getting OS device: %s", err)
					return controllers, raids, noRaidDisks, err
				}
				// Nested raids couldnt be tested as long as I dont have one to test it, so raidLevel always will be 0
				raid = utils.RaidStruct{
					ControllerId: manufacturer + "-" + controllerId,
					RaidLevel:    0,
					Dg:           logicalDeviceNumber,
					RaidType:     logicalDeviceRaidLevel,
					State:        logicalDeviceStatus,
					Size:         logicalDeviceSize,
					OsDevice:     osDevice,
				}
				//fmt.Println("------ RAID object created")
				// Reset values
				logicalDeviceName = ""
				logicalDeviceStatus = ""
				logicalDeviceSize = ""
			}

			// Logical Device disks:
			if strings.Contains(line, "Segment") {
				//fmt.Println("Line: ", line)

				logicalDeviceDiskSize := strings.Fields(line)[4]
				logicalDeviceDiskSize = strings.ReplaceAll(logicalDeviceDiskSize, "(", "")
				logicalDeviceDiskSize = strings.ReplaceAll(logicalDeviceDiskSize, ",", "")
				//fmt.Println("logicalDeviceDiskSize", logicalDeviceDiskSize)

				logicalDeviceDiskInterface := strings.Fields(line)[5]
				logicalDeviceDiskInterface = strings.ReplaceAll(logicalDeviceDiskInterface, ",", "")

				logicalDeviceDiskMedium := strings.Fields(line)[6]
				logicalDeviceDiskMedium = strings.ReplaceAll(logicalDeviceDiskMedium, ",", "")

				logicalDeviceEnclosure := strings.Fields(line)[7]
				logicalDeviceEnclosure = strings.ReplaceAll(logicalDeviceEnclosure, ",", "")
				logicalDeviceEnclosure = strings.ReplaceAll(logicalDeviceEnclosure, "Enclosure:", "")

				logicalDeviceSlot := strings.Fields(line)[8]
				logicalDeviceSlot = strings.ReplaceAll(logicalDeviceSlot, ")", "")
				logicalDeviceSlot = strings.ReplaceAll(logicalDeviceSlot, "Slot:", "")

				eidSlot := logicalDeviceEnclosure + ":" + logicalDeviceSlot
				//fmt.Println("eidSlot: ", eidSlot)

				// Calling megaraidpercsas2ircu.GetMegaraidPercDriveSerialNumber() as made with storcli is not required as
				// arcconf shows that data directly without need of executing a second command
				logicalDeviceSerialNumber := strings.Fields(line)[9]
				logicalDeviceSerialNumber = utils.ClearString(logicalDeviceSerialNumber)

				// fmt.Println("logicalDeviceDiskSize: ", logicalDeviceDiskSize)
				// fmt.Println("logicalDeviceDiskInterface: ", logicalDeviceDiskInterface)
				// fmt.Println("logicalDeviceDiskMedium: ", logicalDeviceDiskMedium)
				// fmt.Println("logicalDeviceEnclosure: ", logicalDeviceEnclosure)
				// fmt.Println("logicalDeviceSlot: ", logicalDeviceSlot)
				// fmt.Println("logicalDeviceSerialNumber: ", logicalDeviceSerialNumber)
				// fmt.Println("-----------------------------")
				disk := utils.DiskStruct{
					ControllerId: manufacturer + "-" + controllerId,
					Dg:           logicalDeviceNumber,
					EidSlot:      eidSlot,
					Size:         logicalDeviceDiskSize,
					Intf:         logicalDeviceDiskInterface,
					Medium:       logicalDeviceDiskMedium,
					SerialNumber: logicalDeviceSerialNumber,
				}
				disks = append(disks, disk)
				continue
			}

			// Add extra drive information:
			// disk state
			if strings.Contains(line, "State") && !strings.Contains(line, "Power State") && !strings.Contains(line, "Supported Power States") {
				physicalDeviceStateData := strings.Split(line, ":")
				physicalDeviceState = physicalDeviceStateData[1]
				physicalDeviceState = utils.ClearString(physicalDeviceState)
				//fmt.Println("physicalDeviceState: ", physicalDeviceState)
				continue
			}

			// disk interface
			if strings.Contains(line, "Transfer Speed") {
				physicalDeviceInterfaceData := strings.Split(line, ":")
				physicalDeviceInterface = physicalDeviceInterfaceData[1]
				physicalDeviceInterface = strings.ReplaceAll(physicalDeviceInterface, "Gb/s", "")
				physicalDeviceInterface = strings.ReplaceAll(physicalDeviceInterface, ".", "")
				regex := regexp.MustCompile("[0-9]+")
				physicalDeviceInterface = regex.ReplaceAllString(physicalDeviceInterface, "")
				physicalDeviceInterface = utils.ClearString(physicalDeviceInterface)
				//fmt.Println("physicalDeviceInterface: ", physicalDeviceInterface)
				continue
			}

			// disk Esd
			if strings.Contains(line, "Reported Location") {
				physicalDeviceEsdData := strings.Split(line, ":")
				physicalDeviceEsd = physicalDeviceEsdData[1]
				physicalDeviceEsd = strings.ReplaceAll(physicalDeviceEsd, "( Connector Unknown )", "")
				physicalDeviceEsd = strings.ReplaceAll(physicalDeviceEsd, "Enclosure", "")
				physicalDeviceEsd = strings.ReplaceAll(physicalDeviceEsd, "Slot", "")
				physicalDeviceEsd = strings.ReplaceAll(physicalDeviceEsd, ",", ":")
				physicalDeviceEsd = utils.ClearString(physicalDeviceEsd)
				//fmt.Println("physicalDeviceEsd: ", physicalDeviceEsd)
				continue
			}

			// disk model
			if strings.Contains(line, "Model") {
				physicalDeviceModelData := strings.Split(line, ":")
				physicalDeviceModel = physicalDeviceModelData[1]
				physicalDeviceModel = utils.ClearString(physicalDeviceModel)
				//fmt.Println("physicalDeviceModel: ", physicalDeviceModel)
				continue
			}

			// disk size
			if strings.Contains(line, "Total Size") {
				physicalDeviceSizeData := strings.Split(line, ":")
				//fmt.Println("physicalDeviceSizeData: ", physicalDeviceSizeData)
				physicalDeviceSize = physicalDeviceSizeData[1]
				//fmt.Println("physicalDeviceSize: ", physicalDeviceSize)
				continue
			}

			// disk medium
			if strings.Contains(line, "SSD") {
				physicalDeviceMediumData := strings.Split(line, ":")
				physicalDeviceMedium = physicalDeviceMediumData[1]
				physicalDeviceMedium = utils.ClearString(physicalDeviceMedium)
				if physicalDeviceMedium == "No" {
					physicalDeviceMedium = "HDD"
				} else {
					physicalDeviceMedium = "SSD"
				}
				//fmt.Println("physicalDeviceMedium: ", physicalDeviceMedium)
				continue
			}

			//fmt.Printf("physicalDeviceEsd: %s physicalDeviceState: %s physicalDeviceSize: %s physicalDeviceInterface: %s physicalDeviceModel: %s|\n", physicalDeviceEsd, physicalDeviceState, physicalDeviceSize, physicalDeviceInterface, physicalDeviceModel)
			if len(physicalDeviceEsd) > 0 && len(physicalDeviceState) > 0 && len(physicalDeviceSize) > 0 && len(physicalDeviceInterface) > 0 && len(physicalDeviceModel) > 0 {
				eidSlotFound := false
				for _, raid := range raids {
					for i := range raid.Disks {
						// range always copies variable values by copy
						// Its compulsory to alter original values making it by ref
						disk := &raid.Disks[i]
						// Calling megaraidpercsas2ircu.GetMegaraidPercDriveSerialNumber() as made with storcli is not required as
						// sas2ircu shows that data directly without need of executing a second command
						// If disk was already added in raid parsing process, medium should be already set, so we dont take care of it here
						//fmt.Printf("Checking physicalDeviceEsd: %v VS disk.EidSlot: %v\n", physicalDeviceEsd, disk.EidSlot)
						if physicalDeviceEsd == disk.EidSlot {
							//fmt.Println("Raid disk detected")
							disk.State = physicalDeviceState
							// Disk size human
							// Remove alphas
							re := regexp.MustCompile(`\D`)
							physicalDeviceSizeValue := re.ReplaceAllString(physicalDeviceSize, "")
							physicalDeviceSizeValue = utils.ClearString(physicalDeviceSizeValue)
							//fmt.Println("physicalDeviceSizeValue: ", physicalDeviceSizeValue)

							// Remove digits
							re = regexp.MustCompile(`\d`)
							physicalDeviceSizeUnit := re.ReplaceAllString(physicalDeviceSize, "")
							physicalDeviceSizeUnit = utils.ClearString(physicalDeviceSizeUnit)
							//fmt.Println("physicalDeviceSizeUnit: ", physicalDeviceSizeUnit)

							multiplierFactor := 1
							switch physicalDeviceSizeUnit {
							case "KB":
								multiplierFactor = 1
							case "MB":
								multiplierFactor = 2
							case "GB":
								multiplierFactor = 3
							case "TB":
								multiplierFactor = 4
							case "PB":
								multiplierFactor = 5
							case "EB":
								multiplierFactor = 5
							case "ZB":
								multiplierFactor = 6
							}
							//fmt.Println("multiplierFactor: ", multiplierFactor)

							physicalDeviceSizeFloat, _ := strconv.ParseFloat(physicalDeviceSizeValue, 64)
							physicalDeviceSizeFloat = physicalDeviceSizeFloat * math.Pow(float64(1024), float64(multiplierFactor))
							physicalDeviceSize = human.Bytes(uint64(physicalDeviceSizeFloat))
							//fmt.Println("physicalDeviceSize: ", physicalDeviceSize)

							disk.Size = physicalDeviceSize

							//disk.intf = physicalDeviceInterface
							disk.Model = physicalDeviceModel
							//fmt.Printf("diskState: %s diskSize: %s diskInterface: %s diskMedium: %s diskModel: %s diskSerialNumber: %s|\n", disk.state, disk.size, disk.intf, disk.medium, disk.model, disk.serialNumber)

							eidSlotFound = true
							// Reset all drive data
							physicalDeviceState = ""
							physicalDeviceSize = ""
							physicalDeviceInterface = ""
							physicalDeviceModel = ""
							break
						}
					}
					if eidSlotFound {
						break
					}
				}

				if !eidSlotFound {
					//fmt.Println("NO Raid disk detected")
					// Adaptec JBOD query
					osDevice, err := hardwarecontrollerscommon.GetJbodOsDevice(manufacturer, controllerId, physicalDeviceEsd)
					if err != nil {
						color.Red("++ ERROR Getting OS device: %s", err)
						return controllers, raids, noRaidDisks, err
					}
					osDevice = "JBOD-" + osDevice
					noRaidDisk := utils.NoRaidDiskStruct{
						ControllerId: manufacturer + "-" + controllerId,
						EidSlot:      physicalDeviceEsd,
						State:        physicalDeviceState,
						Size:         physicalDeviceSize,
						Intf:         physicalDeviceInterface,
						Medium:       physicalDeviceMedium,
						Model:        physicalDeviceModel,
						SerialNumber: diskSerialNumber,
						OsDevice:     osDevice,
					}
					noRaidDisks = append(noRaidDisks, noRaidDisk)
					// Reset all drive data
					physicalDeviceEsd = ""
					physicalDeviceState = ""
					physicalDeviceSize = ""
					physicalDeviceModel = ""
					diskSerialNumber = ""
				}
			}
		}
	}
	return controllers, raids, noRaidDisks, nil
}
