package softraid

import (
	"bufio"
	"fmt"
	"hardwareAnalyzer/utils"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// Function as variable in order to be able to mock it from unit tests
var GetSoftraids = func() (*bufio.Scanner, *os.File, error) {
	readFile, err := os.Open("/proc/mdstat")
	//var scanner bufio.Scanner
	scanner := bufio.NewScanner(readFile)
	return scanner, readFile, err
}

var CheckSoftRaid = func() (bool, error) {
	fmt.Println("> Checking Soft RAID units.")
	scanner, readFile, err := GetSoftraids()
	if err != nil {
		fmt.Println("> No SoftRaid kernel support.")
		return false, nil
	}
	// When GetSoftraids gets mocked from unit tests, theres no file pointer returned
	// So it will be only closed when called from real code in that case it will be != nil
	if readFile != nil {
		defer readFile.Close()
	}

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		//fmt.Println("-- LINE: ", line)
		matched, err := regexp.MatchString(`md\d* : .*`, line)
		if err != nil {
			color.Red("++ ERROR: SoftRaid Regexp errror %s", err)
		}
		if matched {
			color.Magenta("> Soft RAID detected.")
			return true, nil
		}
	}
	fmt.Println("> No SoftRaids detected.")
	return false, nil
}

var ProcessSoftRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.RaidStruct, error) {
	var controllers = []utils.ControllerStruct{}
	var raids = []utils.RaidStruct{}

	fmt.Println("> Getting current softraid configuration.")
	scanner, readFile, err := GetSoftraids()
	if err != nil {
		color.Red("++ ERROR: Something went wrong quering softraid info.")
		return controllers, raids, fmt.Errorf("Error: Something went wrong quering softraid info.")
	}
	// When GetSoftraids gets mocked from unit tests, theres no file pointer returned
	// So it will be only closed when called from real code in that case it will be != nil
	if readFile != nil {
		defer readFile.Close()
	}

	controller := utils.ControllerStruct{
		Id:           "softraid-0",
		Manufacturer: "mdadm",
		Model:        "MDADM",
		Status:       "Good",
	}
	controllers = append(controllers, controller)

	var raid = utils.RaidStruct{}
	var diskState string
	var tempVar []string
	var raidName string
	var raidState string
	var raidType string
	driveStateLine := false

	fmt.Println("> Parsing softraid data.")
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		//fmt.Println("line: ", line)

		// md line
		matched, err := regexp.MatchString(`md\d* : .*`, line)
		if err != nil {
			color.Red("++ ERROR: Regexp errror %s", err)
		}
		if matched {
			raidName = strings.Fields(line)[0]
			raidState = strings.Fields(line)[2]
			var start int
			switch raidState {
			case "active":
				raidState = "Okay"
				raidType = strings.Fields(line)[3]
				// Somre weird raid states(ex: auto-read-only) can be detected this way:
				if !strings.Contains(raidType, "raid") {
					raidState = "Degraded: " + strings.Fields(line)[3]
					raidType = strings.Fields(line)[4]
					start = 5
					controllers[0].Status = "Bad"
				} else {
					start = 4
				}
			default:
				raidState = strings.ToTitle(raidState)
				raidType = "Unknown"
				controllers[0].Status = "Bad"
				start = 3
			}
			// Create Raid
			raid = utils.RaidStruct{
				ControllerId: "softraid-0",
				RaidLevel:    0,
				Dg:           raidName,
				State:        raidState,
				RaidType:     raidType,
				OsDevice:     raidName,
			}

			// Add disks
			n := len(strings.Fields(line))
			//fmt.Println("n: ", n)
			for i := start; i <= n-1; i++ {
				diskDriveData := strings.Fields(line)[i]

				// Check for Failed drives
				matched, err := regexp.MatchString(`^.*\[\d+\]\(F\).*$`, diskDriveData)
				if err != nil {
					color.Red("++ ERROR: Regexp errror %s", err)
				}
				if matched {
					//fmt.Println("FAILED drive detected")
					tempVar = strings.Split(diskDriveData, "(F)")
					diskDriveData = tempVar[0]
					diskState = "Failed"
					raid.State = "Degraded"
					controllers[0].Status = "Bad"
				} else {
					//fmt.Println("ONLINE drive detected")
					diskState = "ONLINE"
				}

				tempVar = strings.Split(diskDriveData, "[")
				diskDrive := tempVar[0]
				diskDrive = utils.ClearString(diskDrive)
				//fmt.Println("diskDrive: ", diskDrive)
				diskSize, err := utils.GetDiskPartitionSize(diskDrive)
				if err != nil {
					color.Red("++ ERROR: utils.GetDiskPartitionSize: %s", err)
				}

				diskSerialNumber := "Unknown"
				diskIntf := "Unknown"
				diskMedium := "Unknown"
				diskModel := "Unknown"

				diskSerialNumber, diskModel, diskIntf, diskMedium, err = utils.GetDiskData(diskDrive)
				if err != nil {
					color.Red("++ ERROR: utils.GetDiskData: %s", err)
				}
				disk := utils.DiskStruct{
					ControllerId: "softraid-0",
					Dg:           raidName,
					State:        diskState,
					Size:         diskSize,
					Intf:         diskIntf,
					Medium:       diskMedium,
					Model:        diskModel,
					SerialNumber: diskSerialNumber,
					OsDevice:     diskDrive,
				}
				//fmt.Println("disk: ", disk)
				raid.AddDisk(disk)
			}

			//fmt.Println("raidName: ", raidName)
			//fmt.Println("raidState: ", raidState)
			//fmt.Println("raidType: ", raidType)
			driveStateLine = true
			continue
		}

		if driveStateLine {
			// driveState line
			n := len(strings.Fields(line))
			diskStateData := strings.Fields(line)[n-1]
			// Some disk failed state
			matched, err = regexp.MatchString(`^\[U*_+.*\]$`, diskStateData)
			if err != nil {
				color.Red("++ ERROR: Regexp errror %s", err)
			}
			if matched {
				raid.State = "Degraded"
				controllers[0].Status = "Bad"
			}

			// Raid size
			raidSize, err := utils.GetDiskPartitionSize(raid.Dg)
			if err != nil {
				color.Red("++ ERROR: utils.GetDiskPartitionSize: %s", err)
			}

			raidSize = strings.TrimSpace(raidSize)
			raid.Size = raidSize

			// Save Raid
			raids = append(raids, raid)
			driveStateLine = false
		}
	}

	// fmt.Println("")
	// for _, raid := range raids {
	// 	fmt.Printf("RaidDg: %s  RaidState: %s RaidType: %s\n", raid.dg, raid.state, raid.raidType)
	// 	for _, disk := range raid.disks {
	// 		fmt.Printf("Disk: %s State: %s\n", disk.osDevice, disk.state)
	// 	}
	// 	fmt.Println("--------")
	// }
	return controllers, raids, nil
}

// ZFS/Btrfs/LVM over SoftRaid disks
func CheckSoftRaidDisks(newRaids []utils.RaidStruct, raids []utils.RaidStruct) error {
	//fmt.Println("-- checkSoftRaidDisks --")
	//fmt.Println("------------ newRaids --------------")
	//spew.Dump(newRaids)
	//fmt.Println("------------ raids --------------")
	//spew.Dump(raids)
	// Check new disks against MD disks, if any of them match, append string to indicate that it is being used by a ZFS/Btrfs/LVM raid
	for _, newRaid := range newRaids {
		for i := range newRaid.Disks {
			newDisk := &newRaid.Disks[i]
			newDiskOsDevice := newDisk.OsDevice
			//fmt.Printf("newDiskOsDevice: |%s|\n", newDiskOsDevice)
			// Previous Raids
			for i := range raids {
				raid := &raids[i]
				//fmt.Printf("raid.osDevice: |%s|\n", raid.osDevice)
				if newDiskOsDevice == raid.OsDevice {
					switch newDisk.ControllerId {
					case "zfs-0":
						raid.OsDevice = raid.OsDevice + " ZFS"
					case "btrfs-0":
						raid.OsDevice = raid.OsDevice + " Btrfs"
					case "lvm-0":
						raid.OsDevice = raid.OsDevice + " LVM"
					default:
						return nil
					}
					// Btrfs over MD device, check MD device info for model/medium/serialNumber info
					newDisk.Model = "Check " + strings.ToUpper(raid.OsDevice) + " disks."
					newDisk.Intf = ""
					newDisk.Medium = "Check " + strings.ToUpper(raid.OsDevice) + " disks."
					newDisk.SerialNumber = "Check " + strings.ToUpper(raid.OsDevice) + " disks."
				}
			}
		}
	}
	return nil
}
