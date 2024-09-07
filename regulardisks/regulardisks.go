package regulardisks

import (
	"fmt"
	"hardwareAnalyzer/utils"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// Function as a variable in order to be mocked from unitary tests
var GetSystemDisks = func() ([]string, error) {
	var diskArray []string
	files, err := os.ReadDir("/sys/block/")
	if err != nil {
		color.Red("++ ERROR: processRegularDisks, error readin /sys/block/: %s", err)
		return diskArray, err
	}

	// Get system disks
	// Ignore DM(LVM), loop, md(MD) devices
	for _, file := range files {
		//fmt.Println(file.Name())
		regularDisk := file.Name()
		// Device mapper
		matched, err := regexp.MatchString(`dm-\d+`, regularDisk)
		if err != nil {
			color.Red("++ ERROR: processRegularDisks Regexp errror %s", err)
		}
		if matched {
			//fmt.Printf("%s discarded.\n", regularDisk)
			continue
		}
		// MTD devices
		matched, err = regexp.MatchString(`mtdblock\d+`, regularDisk)
		if err != nil {
			color.Red("++ ERROR: processRegularDisks Regexp errror %s", err)
		}
		if matched {
			//fmt.Printf("%s discarded.\n", regularDisk)
			continue
		}
		// NBD devices
		matched, err = regexp.MatchString(`nbd\d+`, regularDisk)
		if err != nil {
			color.Red("++ ERROR: processRegularDisks Regexp errror %s", err)
		}
		if matched {
			//fmt.Printf("%s discarded.\n", regularDisk)
			continue
		}
		// loop devices
		matched, err = regexp.MatchString(`loop\d+`, regularDisk)
		if err != nil {
			color.Red("++ ERROR: processRegularDisks Regexp errror %s", err)
		}
		if matched {
			//fmt.Printf("%s discarded.\n", regularDisk)
			continue
		}
		// ram devices
		matched, err = regexp.MatchString(`ram\d+`, regularDisk)
		if err != nil {
			color.Red("++ ERROR: processRegularDisks Regexp errror %s", err)
		}
		if matched {
			//fmt.Printf("%s discarded.\n", regularDisk)
			continue
		}
		// softraid devices
		matched, err = regexp.MatchString(`md\d+`, regularDisk)
		if err != nil {
			color.Red("++ ERROR: processRegularDisks Regexp errror %s", err)
		}
		if matched {
			//fmt.Printf("%s discarded.\n", regularDisk)
			continue
		}
		// CD-ROM devices
		matched, err = regexp.MatchString(`sr\d+`, regularDisk)
		if err != nil {
			color.Red("++ ERROR: processRegularDisks Regexp errror %s", err)
		}
		if matched {
			//fmt.Printf("%s discarded.\n", regularDisk)
			continue
		}
		// ZFS virtualization dataset devices
		matched, err = regexp.MatchString(`zd\d+`, regularDisk)
		if err != nil {
			color.Red("++ ERROR: processRegularDisks Regexp errror %s", err)
		}
		if matched {
			//fmt.Printf("%s discarded.\n", regularDisk)
			continue
		}
		diskArray = append(diskArray, regularDisk)
		//fmt.Println("Disk added")
	}

	if len(diskArray) > 0 {
		return diskArray, nil
	} else {
		return diskArray, fmt.Errorf("Error: GetSystemDisks, diskArray empty len(diskArray): %v.", len(diskArray))
	}
}

var ProcessRegularDisks = func(raids []utils.RaidStruct, noRaidDisks []utils.NoRaidDiskStruct) ([]utils.ControllerStruct, []utils.RaidStruct, error) {
	fmt.Println("> Getting current regular disks configuration.")

	// When addressing with motherboard, only one controller will exist
	// But we return an array in order to be more standar compared to whet returns other HW controllers
	regularDiskControllers := []utils.ControllerStruct{}
	regularDiskController := utils.ControllerStruct{}
	// Same case for raids, it will only exist one
	regularDiskRaids := []utils.RaidStruct{}
	regularDiskRaid := utils.RaidStruct{}

	diskArray, err := GetSystemDisks()
	if err != nil {
		color.Red("++ ERROR: ProcessRegularDisks, error getting system disks: %v.", err)
		return regularDiskControllers, regularDiskRaids, err
	}

	fmt.Println("> Parsing disks.")
	for _, regularDisk := range diskArray {
		// Compare disks with already saved Raid disks
		diskAlreadyFound := false
		// We can label loops in order to break it from other nested loop
	RaidLoop:
		// Check agains raid disks
		for _, raid := range raids {
			//spew.Dump(raid)
			// Discard detected disks that are virtual units presented by HW raids to OS
			//fmt.Printf("regularDisk: %v - Virtual unit: %v\n", regularDisk, raid.OsDevice)
			if regularDisk == raid.OsDevice {
				//fmt.Println("Disk already found1, discarding it: ", regularDisk)
				diskAlreadyFound = true
				break
			}
			for _, disk := range raid.Disks {
				// Discard detected disks that are attached to HW raid controllers
				//fmt.Printf("regularDisk: %v - Controller attached unit: %v\n", regularDisk, disk.OsDevice)
				if regularDisk == disk.OsDevice {
					//fmt.Println("Disk already found2, discarding it: ", regularDisk)
					diskAlreadyFound = true
					break RaidLoop
				}
			}
		}
		// Check agains JBOD disks
		if !diskAlreadyFound {
			for _, disk := range noRaidDisks {
				// Discard JBOD disks
				//fmt.Printf("regularDisk: %v - JBOD unit: %v\n", regularDisk, disk.OsDevice)
				if strings.Contains("JBOD-"+disk.OsDevice, regularDisk) {
					//fmt.Println("Disk already found3, discarding it: ", regularDisk)
					diskAlreadyFound = true
					break
				}
			}
		}

		if !diskAlreadyFound {
			// Get info regularDisk
			//fmt.Println("Standalone disk detected: ", regularDisk)
			diskSize, err := utils.GetDiskPartitionSize(regularDisk)
			if err != nil {
				color.Red("++ ERROR: GetDiskPartitionSize: %s", err)
			}

			//fmt.Println("diskSize: ", diskSize)

			diskSerialNumber := "Unknown"
			diskIntf := "Unknown"
			diskMedium := "Unknown"
			diskModel := "Unknown"
			diskSerialNumber, diskModel, diskIntf, diskMedium, err = utils.GetDiskData(regularDisk)
			if err != nil {
				color.Red("++ ERROR: GetDiskData: %s", err)
			}

			physicalDisk := utils.DiskStruct{
				ControllerId: "motherBoard-0",
				State:        "ONLINE",
				Size:         diskSize,
				Intf:         diskIntf,
				Medium:       diskMedium,
				Model:        diskModel,
				SerialNumber: diskSerialNumber,
				OsDevice:     regularDisk,
			}
			regularDiskRaid.AddDisk(physicalDisk)
		}
	}

	if len(regularDiskRaid.Disks) > 0 {
		// At the start of the function we created empty regularDiskController and regularDiskRaid structures
		// Now that we know it has disks, fill the other structure fields and append to returned function arrays
		regularDiskController.Id = "motherBoard-0"
		regularDiskController.Manufacturer = "motherboard"
		regularDiskController.Model = "MOTHERBOARD"
		regularDiskController.Status = "Good"
		regularDiskControllers = append(regularDiskControllers, regularDiskController)

		regularDiskRaid.ControllerId = "motherBoard-0"
		regularDiskRaid.State = "Good"
		regularDiskRaids = append(regularDiskRaids, regularDiskRaid)
	}

	return regularDiskControllers, regularDiskRaids, nil
}
