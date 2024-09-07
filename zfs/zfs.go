package zfs

import (
	"bufio"
	"fmt"
	"hardwareAnalyzer/utils"
	"io/fs"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// Function as variable in order to be able to mock it from unit tests
var GetZFSPoolSize = func(poolName string) (string, error) {
	command := "list"
	outputStdout, outputStderr, err := utils.GetCommandOutput("zfs", "getZFSPoolSize", command)
	if err != nil {
		color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
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
		//fmt.Println("LINE: ", line)
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		//fmt.Printf("strings.Fields(line)[0]: %s - poolName: %s\n", strings.Fields(line)[0], poolName)
		if strings.Fields(line)[0] == poolName {
			poolSize := strings.Fields(line)[1]

			// Remove alphas
			re := regexp.MustCompile(`[^0-9.]`)
			poolSizeValue := re.ReplaceAllString(poolSize, "")
			poolSizeValue = utils.ClearString(poolSizeValue)
			//fmt.Println("poolSizeValue: ", poolSizeValue)

			// Remove digits
			re = regexp.MustCompile(`[^a-zA-Z]`)
			poolSizeUnit := re.ReplaceAllString(poolSize, "")
			poolSizeUnit = utils.ClearString(poolSizeUnit)

			// The standar followed to show sizes are "Value Unit"
			poolSize = poolSizeValue + " " + poolSizeUnit + "B"

			//fmt.Println("poolSize: ", poolSize)
			return poolSize, nil
		}
	}
	return "Unknown", nil
}

// Function as variable in order to be able to mock it from unit tests
var GetZFSs = func() ([]fs.DirEntry, error) {
	readDir, err := os.ReadDir("/proc/spl/kstat/zfs/")
	return readDir, err
}

var CheckZFSRaid = func() (bool, error) {
	fmt.Println("> Checking ZFS RAIDs.")
	files, err := GetZFSs()
	if err != nil {
		fmt.Println("> No ZFS kernel support.")
		return false, nil
	}

	// Pools are represented as directories
	for _, file := range files {
		//fmt.Println(file.Name(), file.IsDir())
		if file.IsDir() {
			color.Magenta("> ZFS pool detected.")
			return true, nil
		}
	}

	fmt.Println("> No ZFS pools detected.")
	return false, nil
}

// ZFS: Drives are in VDEVs with determine the RAID level.
var ProcessZFSRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.PoolStruct, []utils.RaidStruct, error) {
	var controllers = []utils.ControllerStruct{}
	// pools and vdevs are ralated using: pool.name <-> vdev.dg
	var pools = []utils.PoolStruct{}
	var vdevs = []utils.RaidStruct{}

	controller := utils.ControllerStruct{
		Id:           "zfs-0",
		Manufacturer: "zfs",
		Model:        "ZFS",
		Status:       "Good",
	}
	controllers = append(controllers, controller)

	fmt.Println("> Getting current zpool configuration.")
	command := "status"
	outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "processZFSRaid", command)
	if err != nil {
		color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
		return controllers, pools, vdevs, fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
	}
	if len(outputStderr.String()) != 0 {
		color.Red("++ ERROR: Something went wrong executing command: %s.", command)
		return controllers, pools, vdevs, fmt.Errorf("Error: Something went wrong executing command: %s.", command)
	}
	//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())

	fmt.Println("> Parsing zpool data.")
	scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
	poolName := "Unknown"
	poolState := "Unknown"
	vdevType := "Unknown"
	vdevState := "Unknown"
	insideVdevList := false
	var vdev utils.RaidStruct
	var pool utils.PoolStruct
	previousLineWasDrive := false
	for scanner.Scan() {
		line := scanner.Text()
		// DONT trim line, \n lines are useful for determining end sections
		//fmt.Println("-- LINE: ", line)

		// Last pool line detected
		if insideVdevList && len(line) == 0 {
			insideVdevList = false
			//fmt.Println("End of vdev list detected.")
			vdevs = append(vdevs, vdev)
			// fmt.Println("-----------")
			// fmt.Println("VDEVS-FINAL:")
			// for _, vdev := range vdevs {
			// 	spew.Dump(vdev)
			// }
			// fmt.Println("-----------")
			previousLineWasDrive = false
			continue
		}

		if strings.Contains(line, "pool:") {
			poolName = strings.Fields(line)[1]
			//fmt.Println("poolName: ", poolName)
			continue
		}

		if strings.Contains(line, "state:") {
			poolState = strings.Fields(line)[1]
			if poolState != "ONLINE" {
				controllers[0].Status = "Bad"
			}
			//fmt.Println("poolState: ", poolState)

			poolSize, err := GetZFSPoolSize(poolName)
			if err != nil {
				color.Red("++ ERROR getting poolSize: %s: %s", poolName, err)
			}
			//fmt.Println("poolSize: ", poolSize)

			pool = utils.PoolStruct{
				ControllerId: "zfs-0",
				Name:         poolName,
				State:        poolState,
				Size:         poolSize,
				OsDevice:     "/" + poolName,
			}
			//fmt.Println("pool created: ", poolName)
			pools = append(pools, pool)
			continue
		}

		// When we find poolName it can be a Stripe vdev raid or only the header of the real vdev raid
		if strings.Contains(line, poolName) {
			vdevState = strings.Fields(line)[1]
			insideVdevList = true
			vdevType = "STRIPE"
			// Create vdev:
			vdev = utils.RaidStruct{
				ControllerId: "zfs-0",
				RaidLevel:    0,
				Dg:           poolName,
				RaidType:     vdevType,
				State:        vdevState,
			}
			continue
		}

		// Dont be concerned about strings.Contains, pool name cant match because raidz and mirror are reserved words, not allowing to create raids starting with these names.
		// If previous line was header, its overwritten
		if strings.Contains(line, "mirror") || strings.Contains(line, "raidz") {
			// If we find another vdev in the same pool, save previous vdev
			if previousLineWasDrive {
				vdevs = append(vdevs, vdev)
				previousLineWasDrive = false
			}

			lineData := strings.Split(line, "-")
			vdevType = lineData[0]
			vdevType = utils.ClearString(vdevType)
			//fmt.Printf("vdevType: %s\n", vdevType)
			vdevState = strings.Fields(line)[1]
			//fmt.Printf("vdevState: %s\n", vdevState)
			// Create vdev:
			vdev = utils.RaidStruct{
				ControllerId: "zfs-0",
				RaidLevel:    0,
				Dg:           poolName,
				RaidType:     vdevType,
				State:        vdevState,
			}
			continue
		}

		driveSerialNumber := "Unknown"
		driveModel := "Unknown"
		driveIntf := "Unknown"
		driveMedium := "Unknown"
		if insideVdevList {
			previousLineWasDrive = true
			drive := strings.Fields(line)[0]
			//fmt.Println("Drive: ", drive)
			driveState := strings.Fields(line)[1]
			//fmt.Println("driveState: ", driveState)
			if driveState != "ONLINE" {
				controllers[0].Status = "Bad"
			}

			driveSerialNumber, driveModel, driveIntf, driveMedium, err = utils.GetDiskData(drive)
			if err != nil {
				color.Red("++ ERROR: utils.GetDiskData: %s", err)
			}

			driveSize, _ := utils.GetDiskPartitionSize(drive)
			diskDrive := utils.DiskStruct{
				ControllerId: "zfs-0",
				Dg:           poolName,
				State:        driveState,
				Size:         driveSize,
				Intf:         driveIntf,
				Medium:       driveMedium,
				Model:        driveModel,
				SerialNumber: driveSerialNumber,
				OsDevice:     drive,
			}
			vdev.AddDisk(diskDrive)
		}
	}
	return controllers, pools, vdevs, nil
}
