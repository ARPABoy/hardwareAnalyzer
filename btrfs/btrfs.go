package btrfs

import (
	"bufio"
	"fmt"
	"hardwareAnalyzer/utils"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/Masterminds/semver"
	human "github.com/dustin/go-humanize"

	"github.com/fatih/color"
)

var CheckBtrfsRaid = func() (bool, error) {
	fmt.Println("> Checking Btrfs RAIDs.")

	// Check Btrfs kernel version support
	minimumVersion, err := semver.NewConstraint(">= 3.0")
	if err != nil {
		color.Red("++ ERROR semver error: %s", err)
		return false, err
	}

	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		color.Red("++ ERROR Unable to get syscall info: %s", err)
		return false, err
	}

	//fmt.Println(utils.Int8ToStr(uname.Sysname[:]), utils.Int8ToStr(uname.Release[:]), utils.Int8ToStr(uname.Version[:]))
	currentKernel := utils.Int8ToStr(uname.Release[:])
	currentKernelSplitted := strings.Split(currentKernel, ".")
	currentKernel = currentKernelSplitted[0] + "." + currentKernelSplitted[1]
	//fmt.Println("currentKernel: ", currentKernel)

	currentVersion, err := semver.NewVersion(currentKernel)
	if err != nil {
		color.Red("++ ERROR semver error: %s", err)
		return false, err
	}

	if validKernel, _ := minimumVersion.Validate(currentVersion); !validKernel {
		fmt.Println("> Kernel too old without Btrfs support.")
		return false, nil
	}

	command := "filesystem show"
	outputStdout, outputStderr, err := utils.GetCommandOutput("btrfs", "checkBtrfsRaid", command)
	//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())
	if err != nil {
		color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
		return false, fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
	}
	if len(outputStderr.String()) != 0 {
		// Check if its a missing device error or a real error
		scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
		missingDeviceError := false
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			if strings.Contains(line, "*** Some devices missing") {
				missingDeviceError = true
				break
			}
		}
		if !missingDeviceError {
			color.Red("++ ERROR: Something went wrong executing command: %s.", command)
			return false, fmt.Errorf("Error: Something went wrong executing command: %s.", command)
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		//fmt.Println("line: ", line)
		if strings.Contains(line, "Label:") {
			color.Magenta("> Btrfs volume detected.")
			return true, nil
		}
	}

	fmt.Println("> No Btrfs volumes detected.")
	return false, nil
}

var GetBtrfsRaidType = func(device string) (string, error) {
	//fmt.Println("-- getBtrfsRaidType --")
	device = "/dev/" + device
	//fmt.Println("Device: ", device)

	// We DONT USE utils.GetCommandOutput because we are parsing command output in real time
	raidBinary, exe, err := utils.GetBinaryExecutor("btrfs", "getBtrfsRaidType")
	if err != nil {
		color.Red("++ ERROR utils.GetBinaryExecutor: %s", err)
		return "ERROR utils.GetBinaryExecutor", err
	}
	if exe != nil {
		// When exitting from hardwarecontrollerscommon.GetRaidOSDevice function, last instruction executed will be exe.Close()
		defer exe.Close()
	}

	// Execute btrfs
	// Kernel < 3.17 detected, we cant execute embeded binaries from RAM, we wirtted it to /tmp/hardwareAnalyzerBin instead.
	var cmd *exec.Cmd
	if raidBinary == "/tmp/hardwareAnalyzerBin" {
		defer utils.RemoveFile("getBtrfsRaidType")
		cmd = exec.Command(raidBinary, "inspect-internal", "dump-tree", device)
	} else {
		cmd = exe.Command("inspect-internal", "dump-tree", device)
	}

	// In this case we use StdoutPipe not having to wait for command execution ;)
	// Stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		color.Red("++ ERROR getBtrfsRaidType cmd.StdoutPipe: %s", err)
	}
	defer stdout.Close()
	stdoutReader := bufio.NewReader(stdout)
	//fmt.Println("StdoutPipe linked with stdoutReader")

	//Stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		color.Red("++ ERROR getBtrfsRaidType cmd.StderrPipe: %s", err)
	}
	defer stderr.Close()
	stderrReader := bufio.NewReader(stderr)
	//fmt.Println("StderrPipe linked with stderrReader")

	// Start command
	//fmt.Printf("> Getting Btrfs info: %s, it can take some time, please wait.\n", device)
	err = cmd.Start()
	if err != nil {
		color.Red("++ ERROR getBtrfsRaidType cmd.Start: %s", err)
		return "ERROR getBtrfsRaidType cmd.Start", nil
	}

	// Parse Stdout
	for {
		line, err := stdoutReader.ReadString('\n')
		//fmt.Println("line: ", line)
		if err != nil {
			break
		}
		// Data raid type can missmatch with metadata raid type, we rely on DATA to get raid type
		matched, err := regexp.MatchString(`.*DATA\|.*`, line)
		if err != nil {
			color.Red("++ ERROR: getBtrfsRaidType Regexp errror %s", err)
		}
		if matched {
			//fmt.Println("DATA found")
			//fmt.Println("line: ", line)
			diskRaidTypeData := strings.Split(line, "|")
			diskRaidType := diskRaidTypeData[1]
			diskRaidType = utils.ClearString(diskRaidType)
			diskRaidType = strings.TrimSuffix(diskRaidType, "\n")
			//fmt.Println("diskRaidType: ", diskRaidType)
			return diskRaidType, nil
		}
	}
	// Parse Stderr
	for {
		line, err := stderrReader.ReadString('\n')
		//fmt.Printf("line: |%s|\n", line)
		if err != nil {
			break
		}
		if line == "ERROR: cannot read chunk root\n" {
			return "Cant read chunk root", nil
		}
	}
	return "Unknown", nil
}

var GetBtrfsRaidSize = func(raid utils.RaidStruct) (string, error) {
	//fmt.Println("-- getBtrfsRaidSize --")
	//spew.Dump(raid)

	// SINGLE: ZFS stripe
	// MIXED: ZFS stripe but mixing data with metadata
	// DUP: 2 copies of the data in the same disk
	// RAID0
	// RAID1: 2 copies of the data in diferent disks
	// RAID1c3: 3 copies of the data in diferent disks
	// RAID1c4: 4 copies of the data in diferent disks
	// RAID10
	// RAID5: Stripped data betwwen disks with 1 faulty disk fault tolerance
	// RAID6: Stripped data betwwen disks with 2 faulty disk fault tolerance

	var totalSumDisks float64
	var raidSize float64
	var diskSize float64
	re := regexp.MustCompile(`^(\d+\.\d+)\s([KMGTPEZ]iB)$`)
	for i, disk := range raid.Disks {
		//fmt.Println("disk.Size", disk.Size)
		multiplierFactor := 1
		matches := re.FindStringSubmatch(disk.Size)
		diskSizeString := matches[1]
		diskSizeUnit := matches[2]
		diskSize, _ = strconv.ParseFloat(diskSizeString, 64)

		switch diskSizeUnit {
		case "KiB":
			multiplierFactor = 1
		case "MiB":
			multiplierFactor = 2
		case "GiB":
			multiplierFactor = 3
		case "TiB":
			multiplierFactor = 4
		case "PiB":
			multiplierFactor = 5
		case "EiB":
			multiplierFactor = 5
		case "ZiB":
			multiplierFactor = 6
		}
		diskSize = diskSize * math.Pow(float64(1024), float64(multiplierFactor))
		//fmt.Println("diskSize bytes: ", diskSize)
		if i == 0 {
			totalSumDisks = diskSize
		} else {
			totalSumDisks = totalSumDisks + diskSize
		}
	}

	//fmt.Printf("raid.raidType: |%s|\n", raid.raidType)
	// Sometimes raidType is uppercase, sometimes lowercase
	switch raid.RaidType {
	case "SINGLE", "single":
		raidSize = totalSumDisks
	case "MIXED", "mixed":
		raidSize = totalSumDisks
	case "DUP", "dup":
		raidSize = totalSumDisks / 2
	case "RAID0", "raid0":
		raidSize = totalSumDisks
	case "RAID1", "raid1":
		raidSize = totalSumDisks / 2
	case "RAID1c3", "raid1c3":
		raidSize = totalSumDisks / 3
	case "RAID1c4", "raid1c4":
		raidSize = totalSumDisks / 4
	case "RAID10", "raid10":
		raidSize = totalSumDisks / 2
	case "RAID5", "raid5":
		raidSize = totalSumDisks - diskSize
	case "RAID6", "raid6":
		raidSize = totalSumDisks - (diskSize * 2)
	default:
		return "Unknown", fmt.Errorf("Incorrect RAID type")
	}
	raidSizeFinal := human.Bytes(uint64(raidSize))
	raidSizeFinal = raidSizeFinal + " Aprox"
	//fmt.Println("raidSizeFinal: ", raidSizeFinal)
	return raidSizeFinal, nil
}

var ProcessBtrfsRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.RaidStruct, error) {
	// We collect Btrfs information without requiring to mount filesystem, this program will work even filesystem isnt or cant be mounted
	var controllers = []utils.ControllerStruct{}
	var raids = []utils.RaidStruct{}

	controller := utils.ControllerStruct{
		Id:           "btrfs-0",
		Manufacturer: "btrfs",
		Model:        "Btrfs",
		Status:       "Good",
	}
	controllers = append(controllers, controller)

	fmt.Println("> Getting current Btrfs configuration.")
	command := "filesystem show"
	outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "processBtrfsRaid", command)
	//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())

	// If error happened, check the source
	if err != nil {
		color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
		return controllers, raids, fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
	}
	if len(outputStderr.String()) != 0 {
		// Check if its a missing device error or a real error
		scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
		missingDeviceError := false
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			if strings.Contains(line, "*** Some devices missing") {
				missingDeviceError = true
				break
			}
		}
		if !missingDeviceError {
			color.Red("++ ERROR: Something went wrong executing command: %s.", command)
			return controllers, raids, fmt.Errorf("Error: Something went wrong executing command: %s.", command)
		}
	}

	fmt.Println("> Parsing Btrfs data.")
	// Get total command output lines
	scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
	lastLine := 0
	for scanner.Scan() {
		scanner.Text()
		// Dont skip blank lines, it are userful for controlling edge of each raid
		// line := scanner.Text()
		// line = strings.TrimSpace(line)
		// if len(line) == 0 {
		// 	continue
		// }
		lastLine++
	}

	scanner = bufio.NewScanner(strings.NewReader(outputStdout.String()))
	uuid := "Unknown"
	osDevice := "Unknown"
	diskSize := "Unknown"
	var raid = utils.RaidStruct{}
	firstDisk := true
	var btrfsDisk = utils.DiskStruct{}
	diskModel := "Unknown"
	diskIntf := "Unknown"
	diskMedium := "Unknown"
	diskSerialNumber := "Unknown"
	firstOutputLine := true
	currentLine := 0
	for scanner.Scan() {
		line := scanner.Text()
		// Dont skip blank lines, it are userful for controlling edge of each raid
		// line = strings.TrimSpace(line)
		// if len(line) == 0 {
		// 	continue
		// }
		//fmt.Println("-- LINE: ", line)
		//fmt.Println("line: ", line)
		//fmt.Printf("Processing line: %d/%d\n", currentLine, lastLine)
		currentLine++

		// End of raid detected, fill extra raid data and append raid to raids array
		if currentLine == lastLine || (strings.Contains(line, "uuid:") && !firstOutputLine) {
			//fmt.Println("Raid end detected")
			// Get raid type
			raidType := "Unknown"
			previousRaidType := "Unknown"
			allRaidTypesMatches := false
			for i, btrfsDisk := range raid.Disks {
				if i == 0 {
					raidType, _ = GetBtrfsRaidType(btrfsDisk.OsDevice)
					allRaidTypesMatches = true
				} else {
					raidType, _ = GetBtrfsRaidType(btrfsDisk.OsDevice)
					if raidType == previousRaidType {
						allRaidTypesMatches = true
					} else {
						color.Red("++ ERROR: raid type missmatch: %s != %s", raidType, previousRaidType)
						allRaidTypesMatches = false
					}
				}
				//fmt.Println("raidType: ", raidType)
				previousRaidType = raidType
			}
			if allRaidTypesMatches {
				raid.RaidType = raidType
			} else {
				raid.RaidType = "Unknown"
			}
			//fmt.Println("raid.raidType: ", raid.raidType)

			// Calculate raid size knowing devices size and raid type
			if raid.RaidType != "Unknown" {
				raidSize, _ := GetBtrfsRaidSize(raid)
				raid.Size = raidSize
			} else {
				raid.Size = "Unknown"
			}

			raids = append(raids, raid)
			//fmt.Println("Raid appended")
			firstDisk = true
		}

		// Add raid
		if strings.Contains(line, "uuid:") {
			firstOutputLine = false
			uuidData := strings.Split(line, ":")
			uuid = uuidData[2]
			uuid = utils.ClearString(uuid)
			//fmt.Println("uuid: ", uuid)
			raid = utils.RaidStruct{
				ControllerId: "btrfs-0",
				RaidLevel:    0,
				State:        "ONLINE",
				Dg:           uuid,
			}
			//fmt.Println("Raid object created")
			continue
		}
		// Add disks to raid
		if strings.Contains(line, "devid") {
			//fmt.Println("devid detected")
			//fmt.Println(strings.Fields(line))
			diskSize = strings.Fields(line)[3]
			//fmt.Println("diskSize: ", diskSize)
			// All data is always shown with Num Unit syntax, adapt Btrfs to standar output
			// Ex: 931.51GiB -> 931.51 GiB
			re := regexp.MustCompile(`^(\d+\.\d+)([KMGTPEZ]iB)$`)
			matches := re.FindStringSubmatch(diskSize)
			diskSizeString := matches[1]
			//fmt.Println("diskSizeString: ", diskSizeString)
			diskSizeUnit := matches[2]
			//fmt.Println("diskSizeUnit: ", diskSizeUnit)
			diskSize = diskSizeString + " " + diskSizeUnit
			//fmt.Println("diskSize: ", diskSize)
			osDevice = strings.Fields(line)[7]
			osDevice = strings.ReplaceAll(osDevice, "/dev/", "")
			//fmt.Println("osDevice: ", osDevice)
			btrfsDisk = utils.DiskStruct{
				ControllerId: "btrfs-0",
				Dg:           uuid,
				State:        "ONLINE",
				Size:         diskSize,
				OsDevice:     osDevice,
			}
			//fmt.Println("Btrfs disk object created")

			diskSerialNumber, diskModel, diskIntf, diskMedium, err = utils.GetDiskData(btrfsDisk.OsDevice)
			if err != nil {
				color.Red("++ ERROR: utils.GetDiskData: %s", err)
			}

			btrfsDisk.SerialNumber = diskSerialNumber
			btrfsDisk.Intf = diskIntf
			btrfsDisk.Medium = diskMedium
			btrfsDisk.Model = diskModel

			raid.AddDisk(btrfsDisk)
			//fmt.Println("Disk added to raid")

			if firstDisk {
				raid.OsDevice = osDevice
				firstDisk = false
			}
			continue
		}

		// Missing device detected
		if strings.Contains(line, "*** Some devices missing") {
			raid.State = "Missing devices"
			controllers[0].Status = "Bad"
			continue
		}
		if strings.Contains(line, "is missing") {
			raid.State = "Missing devices"
			controllers[0].Status = "Bad"
			continue
		}

		// When there are mising devices it can be found any "random" error messages, skip it
		if strings.Contains(line, "bad tree block") || strings.Contains(line, "cannot read chunk root") {
			continue
		}
	}
	return controllers, raids, nil
}
