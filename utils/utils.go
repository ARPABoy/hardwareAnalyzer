package utils

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/Masterminds/semver"
	"github.com/amenzhinsky/go-memexec"
	human "github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/shirou/gopsutil/disk"
)

// Using the //go:embed comment to your code, the compiler will include files in the resulting static binary.
//
//go:embed binaries/storcli
var Storcli []byte

//go:embed binaries/perccli
var Perccli []byte

//go:embed binaries/sas2ircu
var Sas2ircu []byte

//go:embed binaries/arcconf
var Arcconf []byte

//go:embed binaries/arcconfdynamic
var Arcconfdynamic []byte

//go:embed binaries/zpool
var Zpool []byte

//go:embed binaries/zpooldynamic
var Zpooldynamic []byte

//go:embed binaries/btrfs
var Btrfs []byte

//go:embed binaries/btrfsdynamic
var Btrfsdynamic []byte

//go:embed binaries/lvm
var Lvm []byte

// Clear spaces and tabs from string
func ClearString(stringToClear string) string {
	stringToClear = strings.ReplaceAll(stringToClear, " ", "")
	stringToClear = strings.ReplaceAll(stringToClear, "\t", "")
	return stringToClear
}

// memexec requires Kernel >= 3.17 and glibc >= 2.27: syscall_319 (errno 38)
// In these cases we write binary to disk and execute it from this location instead of directly from RAM
func WriteExecutableFile(binaryFile []byte) error {
	err := os.WriteFile("/tmp/hardwareAnalyzerBin", binaryFile, 0700)
	if err != nil {
		color.Red("++ ERROR Could not write file: %s", err)
		return err
	} else {
		return nil
	}
}

// Copy system binary to /tmp/hardwareAnalyzerBin
func CopySystemBinary(srcFile string) error {
	RemoveFile("CopySystemBinary")

	// Open srcFile
	sourceFile, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("++ ERROR opening source file: %v", err)
	}
	defer sourceFile.Close()

	// Create dst
	destFile, err := os.Create("/tmp/hardwareAnalyzerBin")
	if err != nil {
		return fmt.Errorf("++ ERROR creating destination file: %v", err)
	}
	defer destFile.Close()

	// Copy srs -> dst
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("++ ERROR copying: %v", err)
	}

	// Sync dst
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("++ ERROR syncing file: %v", err)
	}

	// Assign dst permissions
	permissions := os.FileMode(0755)
	err = os.Chmod("/tmp/hardwareAnalyzerBin", permissions)
	if err != nil {
		return fmt.Errorf("++ ERROR assigning file permissions: %v", err)
	}

	return nil
}

// Remove WriteExecutableFile binary
func RemoveFile(callingFunction string) error {
	//fmt.Println("-- removeFile called by: ", callingFunction)

	// Sometimes file has been already deleted by functions called inside functions
	// The first function generates the file, the second function generates it and deletes
	// when code returns to firs function, tries to delete it failing because it has been already deleted.
	fileToRemove := "/tmp/hardwareAnalyzerBin"
	_, err := os.Stat(fileToRemove)
	if os.IsNotExist(err) {
		return nil
	}

	err = os.Remove(fileToRemove)
	if err != nil {
		color.Red("++ ERROR Could not remove file: %s", err)
		return err
	} else {
		return nil
	}
}

// Convert string to uint64
func Str2uint64(str string) (uint64, error) {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		color.Red("++ ERROR str2uint64 error: %s", err)
		return uint64(i), err
	}
	return uint64(i), nil
}

// Convert int8 to string
func Int8ToStr(arr []int8) string {
	//fmt.Println("arr:", arr)
	b := make([]byte, 0, len(arr))
	for _, v := range arr {
		if v == 0x00 {
			break
		}
		b = append(b, byte(v))
	}
	//fmt.Println("string(b): ", string(b))
	return string(b)
}

// Check if user is root
func IsRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		color.Red("++ ERROR Unable to get current user: %s", err)
		return false
	}
	//fmt.Println(currentUser)
	return currentUser.Username == "root"
}

// Check if OS is supported
func SupportedOS() (bool, error) {
	// Static binaries seems to work in FreeBSD too, but I dont have any RAID controller system to test it
	OS := runtime.GOOS
	switch OS {
	case "linux":
		return true, nil
	default:
		//color.Red("Sorry, this program only supports Linux systems")
		return false, fmt.Errorf("Sorry, this program only supports Linux systems")
	}
}

// Check if kernel version support memexec
func CheckMemExecKernelSupport() (bool, string, error) {
	minimumVersion, err := semver.NewConstraint(">= 3.17")
	if err != nil {
		color.Red("++ ERROR semver error: %s", err)
		return false, "", err
	}

	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		color.Red("++ ERROR Unable to get syscall info: %s", err)
		return false, "", err
	}
	// extract members:
	// type Utsname struct {
	//  Sysname    [65]int8
	//  Nodename   [65]int8
	//  Release    [65]int8
	//  Version    [65]int8
	//  Machine    [65]int8
	//  Domainname [65]int8

	//fmt.Println(int8ToStr(uname.Sysname[:]), int8ToStr(uname.Release[:]), int8ToStr(uname.Version[:]))
	currentKernel := Int8ToStr(uname.Release[:])
	currentKernelSplitted := strings.Split(currentKernel, ".")
	currentKernel = currentKernelSplitted[0] + "." + currentKernelSplitted[1]
	//fmt.Println("currentKernel: ", currentKernel)

	currentVersion, err := semver.NewVersion(currentKernel)
	if err != nil {
		color.Red("++ ERROR semver error: %s", err)
		return false, "", err
	}

	if validKernel, _ := minimumVersion.Validate(currentVersion); validKernel {
		//fmt.Println("VALID KERNEL VERSION")
		return true, currentKernel, nil
	} else {
		//fmt.Println("INVALID KERNEL VERSION")
		return false, currentKernel, fmt.Errorf("Minimum kernel version required >= 3.17")
	}
}

// Function as variable in order to be possible to be mocked from unit tests
var GetDiskPartitionSize = func(diskDrive string) (string, error) {
	//fmt.Println("-- getDiskPartitionSize --")
	//fmt.Println("diskDrive: ", diskDrive)
	diskSize := "Unknown"
	readFile, err := os.Open("/proc/partitions")
	if err != nil {
		color.Red("++ ERROR Could not read /proc/partitions file: %s", err)
		return "", err
	}

	defer readFile.Close()
	scanner := bufio.NewScanner(readFile)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Line contains diskDrive, if its a whole device, it can have multiple matches, thats because we check exact match two lines below
		if strings.Contains(line, diskDrive) {
			partitionDevice := strings.Fields(line)[3]
			// Check exact match
			if partitionDevice == diskDrive {
				//fmt.Println("Disk found in /proc/partitions")
				partitionBlocks := strings.Fields(line)[2]
				partitionBlocksInt, _ := Str2uint64(partitionBlocks)
				// blockSize is always 1024
				// man proc: Contains the major and minor numbers of each partition as well as the number of 1024-byte blocks and the partition name.
				blockSize, _ := Str2uint64("1024")
				partitionBlocksInt = partitionBlocksInt * blockSize
				diskSize = human.Bytes(partitionBlocksInt)
				//fmt.Println("diskSize: ", diskSize)
				return diskSize, nil
			}
		}
	}
	return diskSize, nil
}

func GetDiskPartitionInterface(diskDataRaw string) (string, error) {
	//fmt.Println("-- getDiskPartitionInterface --")
	//fmt.Println("diskDataRaw: ", diskDataRaw)
	intf := "Unknown"
	files, err := os.ReadDir("/dev/disk/by-id/")
	if err != nil {
		color.Red("++ ERROR Could not getDiskPartitionInterface: %s", err)
		return intf, fmt.Errorf("Could not getDiskPartitionInterface: %s", err)
	}

	for _, file := range files {
		//fmt.Println("Checking: ", file.Name())
		if strings.Contains(file.Name(), diskDataRaw) {
			//fmt.Println("Matched")
			intfData := strings.Split(file.Name(), "-")
			intf = intfData[0]
			//fmt.Println("intf: ", intf)
			return intf, nil
		}
	}
	return intf, nil
}

// Function as variable in order to be possible to mock it from unitary tests
var GetDiskData = func(diskDrive string) (string, string, string, string, error) {
	diskSerialNumber := "Unknown"
	diskIntf := "Unknown"
	diskMedium := "Unknown"
	diskModel := "Unknown"
	// diskSerialNumber: gopsutil/disk GetDiskSerialNumber function call
	diskDataRaw := disk.GetDiskSerialNumber("/dev/" + diskDrive)
	diskDataRaw = strings.ReplaceAll(diskDataRaw, " ", "_")
	//fmt.Println("diskDataRaw: ", diskDataRaw)
	diskData := strings.Split(diskDataRaw, "_")
	if len(diskData) == 1 && diskData[0] != "" {
		diskSerialNumber = diskDataRaw
	} else {
		// Sometimes disk.GetDiskSerialNumber doesnt get serialnumber
		if len(diskData) > 1 {
			diskSerialNumber = diskData[len(diskData)-1]
			diskSerialNumberCounter := 1
			for len(diskSerialNumber) < 5 && diskSerialNumberCounter < 3 {
				diskSerialNumber = diskData[len(diskData)-diskSerialNumberCounter]
				diskSerialNumberCounter++
			}
			//fmt.Println("diskSerialNumber: ", diskSerialNumber)
			// diskModel
			for j := 0; j <= len(diskData)-2; j++ {
				if j == 0 {
					diskModel = diskData[j]
				} else {
					diskModel = diskModel + " " + diskData[j]
				}
			}
			// diskIntf
			diskIntf, _ = GetDiskPartitionInterface(diskDataRaw)
			//fmt.Println("diskIntf: ", diskIntf)
			if diskIntf != "nvme" && diskIntf != "ata" && diskIntf != "scsi" && diskIntf != "SAS" && diskIntf != "SATA" && diskIntf != "md" && diskIntf != "dm" {
				diskIntf = "Unknown"
			}
			// diskMedium
			diskMedium = diskData[1]
			//fmt.Println("diskMedium: ", diskMedium)
			if diskMedium != "HDD" && diskMedium != "SATA_HDD" && diskMedium != "SSD" && diskMedium != "NVME" {
				diskMediumAssigned := false
				if strings.Contains(diskModel, "SSD") || strings.Contains(diskModel, "ssd") {
					diskMedium = "SSD"
					diskMediumAssigned = true
				}
				if strings.Contains(diskModel, "NVME") || strings.Contains(diskModel, "nvme") {
					diskMedium = "NVME"
					diskMediumAssigned = true
				}
				if !diskMediumAssigned {
					diskMedium = diskIntf
				}
			}
		}
	}
	//fmt.Printf("diskSerialNumber: %v diskModel: %v diskIntf: %v diskMedium: %v\n", diskSerialNumber, diskModel, diskIntf, diskMedium)
	return diskSerialNumber, diskModel, diskIntf, diskMedium, nil
}

// Get binary executor depending of the manufacturer
func GetBinaryExecutor(manufacturer string, callingFunction string) (string, *memexec.Exec, error) {
	//fmt.Printf("-- getBinaryExecutor manufacturer: %s, called by: %s --\n", manufacturer, callingFunction)

	raidBinaryName := "Unknown"
	checkCommand := []string{}
	var raidBinary []byte
	switch manufacturer {
	case "mega":
		raidBinaryName = "storcli"
		raidBinary = Storcli
		checkCommand = []string{"/call", "show", "all"}
	case "perc":
		raidBinaryName = "perccli"
		raidBinary = Perccli
		checkCommand = []string{"/call", "show", "all"}
	case "sas2ircu":
		raidBinaryName = "sas2ircu"
		raidBinary = Sas2ircu
		checkCommand = []string{"LIST"}
	case "adaptec":
		raidBinaryName = "arcconf"
		raidBinary = Arcconf
		checkCommand = []string{"LIST"}
	case "zfs":
		raidBinaryName = "zpool"
		raidBinary = Zpool
		checkCommand = []string{"list"}
	case "btrfs":
		raidBinaryName = "btrfs"
		raidBinary = Btrfs
		checkCommand = []string{"fi", "show"}
	case "lvm":
		raidBinaryName = "lvm"
		raidBinary = Lvm
		checkCommand = []string{"lvs"}
	default:
		return raidBinaryName, nil, fmt.Errorf("Unknown manufacturer.")
	}

	// We try 4 execution methods:
	// - Memory execution: last version tool.
	// - Disk file execution: last version tool, static version.
	// - Disk file execution: last version tool, dynamic version.
	// - System tool.

	raidBinaryFile := "/tmp/hardwareAnalyzerBin"

	//fmt.Println("Trying memory execution")
	// memexec requires Kernel >= 3.17 and glibc >= 2.27: syscall_319 (errno 38)
	// If we detect previous versions, copy binary to temp directory and execute it
	//supportedKernel, currentKernel, err = CheckKernelVersion()
	memExecSupport, _, _ := CheckMemExecKernelSupport()
	//fmt.Println("memExecSupport: ", memExecSupport)
	//fmt.Println("currentKernel: ", currentKernel)

	// MEMORY:
	if memExecSupport {
		//fmt.Println("Generating exec from memexec.")
		// Generate exe from embedded storcli/perccli/sas2ircu/arcconf/zpool/btrfs/lvm
		exe, err := memexec.New(raidBinary)
		if err != nil {
			color.Red("++ ERROR memexec error: %s", err)
		}
		//fmt.Println("Memexec returned.")
		return raidBinaryName, exe, nil
	} else {
		//fmt.Println("Trying disk static execution")

		// DISK STATIC
		_ = WriteExecutableFile(raidBinary)
		cmd := exec.Command(raidBinaryFile, checkCommand...)
		var outputStdout, outputStderr bytes.Buffer
		cmd.Stdout = &outputStdout
		cmd.Stderr = &outputStderr
		err := cmd.Run()
		// if raidBinaryName == "sas2ircu" {
		// 	fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())
		// }
		// fmt.Printf("raidBinaryName: |%v|\n", raidBinaryName)
		// fmt.Printf("ERROR: |%v|\n", err)
		// fmt.Println("ERR OUTPUT:", outputStderr.String())

		if err != nil {
			// sas2ircu fake output error, ignore it
			if raidBinaryName == "sas2ircu" && err.Error() == "exit status 1" && strings.Contains(outputStdout.String(), "SAS2IRCU: MPTLib2 Error 1") {
				err = nil
			}
		}

		if err == nil && len(outputStderr.String()) == 0 {
			// When we use file execution, we only must return file path
			//fmt.Println("Static binary executed successfuly")
			return raidBinaryFile, nil, nil
		}

		//fmt.Println(err)
		//fmt.Println("err:", outputStderr.String())

		// DISK DYNAMIC
		// Some really old kernels(ex:2.6.32) works better with dynamically linked binary version.
		// Check dynamic version, not all binaries have dynamic version
		if raidBinaryName == "arcconf" || raidBinaryName == "btrfs" || raidBinaryName == "zpool" {
			//fmt.Println("Trying disk dynamic execution")
			switch raidBinaryName {
			case "arcconf":
				raidBinary = Arcconfdynamic
			case "zpool":
				raidBinary = Zpooldynamic
			case "btrfs":
				raidBinary = Btrfsdynamic
			}

			_ = WriteExecutableFile(raidBinary)
			cmd := exec.Command(raidBinaryFile, checkCommand...)
			var outputStdout, outputStderr bytes.Buffer
			cmd.Stdout = &outputStdout
			cmd.Stderr = &outputStderr
			err := cmd.Run()
			//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())
			// fmt.Printf("raidBinaryName: |%v|\n", raidBinaryName)
			// fmt.Printf("ERROR: |%v|\n", err)
			// fmt.Println("ERR OUTPUT:", outputStderr.String())

			// sas2ircu fake output error, ignore it
			if raidBinaryName == "sas2ircu" && err.Error() == "exit status 1" && strings.Contains(outputStdout.String(), "SAS2IRCU: MPTLib2 Error 1") {
				err = nil
			}

			if err == nil && len(outputStderr.String()) == 0 {
				//fmt.Println("Dynamic binary executed successfuly")
				// When we use file execution, we only must return file path
				return raidBinaryFile, nil, nil
			}
		}

		// SYSTEM VERSION
		// lvm cant be renamed, so if we copy it to /tmp/hardwareAnalyzerBin, we will get an execution error
		if raidBinaryName != "lvm" {
			//fmt.Println("Trying system binary execution")
			var commonPaths = []string{
				"/bin",
				"/usr/bin",
				"/sbin",
				"/usr/sbin",
				"/usr/local/bin",
			}

			fullPath := ""
			for _, dir := range commonPaths {
				fullPath = filepath.Join(dir, raidBinaryName)
				//fmt.Println("Checking tool path: ", fullPath)
				if _, err := os.Stat(fullPath); err == nil {
					//fmt.Println("Tool found: ", fullPath)
					if err := CopySystemBinary(fullPath); err != nil {
						return raidBinaryFile, nil, err
					}
					return raidBinaryFile, nil, nil
				}
			}
		}
		//fmt.Println("Tool not found")
	}
	return raidBinaryFile, nil, fmt.Errorf("Cant execute required binaries, maybe your system is too old.")
}

// Function as variable in order to be possible to mock it from unitary tests
// Get command output
var GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
	//fmt.Printf("-- getCommandOutput  from: %s - manufacturer: %s command: %s --\n", callingFunction, manufacturer, command)

	raidBinary, exe, err := GetBinaryExecutor(manufacturer, callingFunction)
	if err != nil {
		return nil, nil, err
	}
	//fmt.Println("raidBinary: ", raidBinary)

	// When code is exitting from getCommandScanner function, last instruction executed will be exe.Close()
	if exe != nil {
		defer exe.Close()
	}

	args := strings.Split(command, " ")
	var cmd *exec.Cmd
	// Kernel < 3.17 detected, we cant execute embeded binaries from RAM, we writted it to /tmp/hardwareAnalyzerBin instead.
	if raidBinary == "/tmp/hardwareAnalyzerBin" {
		defer RemoveFile(callingFunction)
		cmd = exec.Command(raidBinary, args...)
	} else {
		cmd = exe.Command(args...)
	}
	var outputStdout, outputStderr bytes.Buffer
	cmd.Stdout = &outputStdout
	cmd.Stderr = &outputStderr
	err = cmd.Run()
	//fmt.Println("Command: ", command)
	//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())

	// Some error happened while executing command
	if err != nil {
		// Dont warn already, analyze stderr in origin function before considering a real command execution failure
		return &outputStdout, &outputStderr, err
	}
	if len(outputStderr.String()) != 0 {
		// Dont warn already, analyze stderr in origin function before considering a real command execution failure
		return &outputStdout, &outputStderr, nil
	}

	// No errors
	return &outputStdout, &outputStderr, nil
}

// Try to fill the most information about disks
func CheckPreviousDiskData(newRaids []RaidStruct, raids []RaidStruct) error {
	// fmt.Println("-- checkPreviousDiskData --")
	// fmt.Println("------------ newRaids -----------------")
	// spew.Dump(newRaids)
	// fmt.Println("------------ raids -----------------")
	// spew.Dump(raids)
	// range always copies variable values by copy
	for i := range newRaids {
		newRaid := &newRaids[i]
		for j := range newRaid.Disks {
			newDisk := &newRaid.Disks[j]
			diskFound := false
			// Already saved Raids and disks are accessed in RO mode
			for _, raid := range raids {
				if diskFound {
					break
				}
				for _, disk := range raid.Disks {
					if diskFound {
						break
					}
					//fmt.Printf("New: %s -> %s              Existent: %s -> %s\n", newDisk.controllerId, newDisk.osDevice, disk.controllerId, disk.osDevice)
					// Remove digits
					re := regexp.MustCompile(`\d`)
					newDiskOsDevice := re.ReplaceAllString(newDisk.OsDevice, "")
					diskOsDevice := re.ReplaceAllString(disk.OsDevice, "")
					raidOsDevice := re.ReplaceAllString(raid.OsDevice, "")

					// Check against Hardware Raids or other disks
					if newDiskOsDevice == raidOsDevice || newDiskOsDevice == diskOsDevice {
						//fmt.Println("Disk match")
						if newDisk.Model == "Unknown" {
							newDisk.Model = disk.Model
						}
						if newDisk.Intf == "Unknown" {
							newDisk.Intf = disk.Intf
						}
						if newDisk.Medium == "Unknown" {
							newDisk.Medium = disk.Medium
						}
						if newDisk.SerialNumber == "Unknown" {
							newDisk.SerialNumber = disk.SerialNumber
						}
						diskFound = true
					}
				}
			}
		}
	}
	return nil
}

func ShowGatheredData(controllers []ControllerStruct, pools []PoolStruct, volumeGroups []VolumeGroupStruct, raids []RaidStruct, noRaidDisks []NoRaidDiskStruct) error {
	//Show gathered data
	// fmt.Println("-- showGatheredData --")
	// fmt.Println("-- controllers --")
	// spew.Dump(controllers)
	// fmt.Println("-- pools --")
	// spew.Dump(pools)
	// fmt.Println("-- raids --")
	// spew.Dump(raids)
	//fmt.Println("-- noRaidDisks --")
	//spew.Dump(noRaidDisks)

	if len(raids) == 0 && len(noRaidDisks) == 0 {
		fmt.Println("> No RAIDs detected.")
	} else {
		for _, controller := range controllers {
			// Code commented due to HW controllers without RAIDs/Disks or only JBOD disks attached
			// Even when theres no RAID/Disk attached its worth to show it to advise that the controller is present in the system
			// controllerDiskCount := 0
			// for _, raid := range raids {
			// 	if raid.ControllerId == controller.Id {
			// 		controllerDiskCount = len(raid.Disks)
			// 	}
			// }
			// if controllerDiskCount == 0 {
			// 	//fmt.Println("Empty controller detected.")
			// 	continue
			// }

			// Search bogus disks and mark controller as Bad
			for _, raid := range raids {
				if raid.ControllerId == controller.Id {
					for _, disk := range raid.Disks {
						if disk.Size == "Unknown" && disk.Model == "Unknown" && disk.Intf == "Unknown" && disk.Medium == "Unknown" && disk.SerialNumber == "Unknown" {
							controller.Status = "Bad"
						}
					}
				}
			}

			fmt.Println("")
			if controller.Status == "Good" || controller.Status == "Optimal" || controller.Status == "OK" {
				color.Yellow("-- ControllerID: %s - %s: %s", controller.Id, controller.Model, controller.Status)
			} else {
				color.Red("-- ControllerID: %s - %s: %s", controller.Id, controller.Model, controller.Status)
			}

			// Show raids and disks
			zfsPoolListOfShownPools := []string{}
			volumeGroupListOfShownVolumeGroups := []string{}
			shownLvmsHeader := false
			for _, raid := range raids {
				//fmt.Println("raid: ", raid)
				if raid.ControllerId == controller.Id {
					raidLevelTabs := strings.Repeat("  ", raid.RaidLevel)
					// Search bogus disks and mark raid as Bad
					for _, disk := range raid.Disks {
						if disk.Size == "Unknown" && disk.Model == "Unknown" && disk.Intf == "Unknown" && disk.Medium == "Unknown" && disk.SerialNumber == "Unknown" {
							raid.State = "Bad"
						}
					}
					//fmt.Printf("raid.state: |%s|\n", raid.state)
					if raid.State == "Okay(OKY)" || raid.State == "Okay" || raid.State == "Optl" || raid.State == "Optimal" || raid.State == "Good" || raid.State == "ONLINE" || raid.State == "available" {
						switch controller.Manufacturer {
						case "mdadm":
							color.Blue("   %s%s: %s   Size: %s   => %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), raid.State, raid.Size, strings.ToUpper(raid.OsDevice))
						case "zfs":
							// Show pool info
							for _, pool := range pools {
								if !slices.Contains(zfsPoolListOfShownPools, pool.Name) {
									if raid.Dg == pool.Name {
										color.Blue("   Pool: %s  %s - %s  => %s", pool.Name, pool.State, pool.Size, pool.OsDevice)
										zfsPoolListOfShownPools = append(zfsPoolListOfShownPools, pool.Name)
										break
									}
								}
							}
							// Show vdev info
							color.Blue("     %s%s: %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), raid.State)
						case "btrfs":
							color.Blue("   %s%s: %s   Size: %s   => %s - %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), raid.State, raid.Size, strings.ToUpper(raid.Dg), strings.ToUpper(raid.OsDevice))
						case "lvm":
							// Show volumeGroup info
							for _, volumeGroup := range volumeGroups {
								if !slices.Contains(volumeGroupListOfShownVolumeGroups, volumeGroup.Name) {
									if strings.Split(raid.OsDevice, "/")[0] == volumeGroup.Name {
										color.Blue("   Volume Group: %s  %s Size: %s", strings.ToUpper(volumeGroup.Name), volumeGroup.State, volumeGroup.Size)
										color.Blue("     Disks:")
										volumeGroupListOfShownVolumeGroups = append(volumeGroupListOfShownVolumeGroups, volumeGroup.Name)
										// LVM disks are part of the VG not RAID as usually, so we show disks when VG is shown
										for _, disk := range raid.Disks {
											if (disk.State == "Optimal(OPT)" || disk.State == "Onln" || disk.State == "Online" || disk.State == "Good" || disk.State == "ONLINE") && disk.OsDevice != "[UNKNOWN]" {
												color.Green("       %s%s   Size: %s   Model: %s - %s/%s -> SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
											} else {
												color.Red("       %s%s   Size: %s   Model: %s - %s/%s -> SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
											}
										}
										shownLvmsHeader = false
									}
								}
							}
							if !shownLvmsHeader {
								color.Blue("     LVMs:")
								shownLvmsHeader = true
							}
							if raid.OsDevice == "NONE" {
								color.Blue("        %s%s: %s   Size: %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), strings.ToUpper(raid.State), raid.Size)
							} else {
								color.Blue("        %s%s: %s   Size: %s   => %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), strings.ToUpper(raid.State), raid.Size, strings.ToUpper(raid.OsDevice))
							}
						case "motherboard":
							// NOOP
						// HW Raid
						default:
							if raid.RaidLevel > 0 {
								color.Blue("   %s%s: %s   Size: %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), raid.State, raid.Size)
							} else {
								color.Blue("   %s%s: %s   Size: %s   => %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), raid.State, raid.Size, strings.ToUpper(raid.OsDevice))
							}
						}
					} else {
						switch controller.Manufacturer {
						case "mdadm":
							color.Red("   %s%s: %s   Size: %s   => %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), raid.State, raid.Size, strings.ToUpper(raid.OsDevice))
						case "zfs":
							// Show pool info
							for _, pool := range pools {
								if !slices.Contains(zfsPoolListOfShownPools, pool.Name) {
									if raid.Dg == pool.Name {
										color.Red("   Pool: %s  %s - %s  => %s", pool.Name, pool.State, pool.Size, pool.OsDevice)
										zfsPoolListOfShownPools = append(zfsPoolListOfShownPools, pool.Name)
										break
									}
								}
							}
							// Show vdev info
							color.Red("     %s%s: %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), raid.State)
						case "btrfs":
							color.Red("   %s%s: %s   Size: %s   => %s - %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), raid.State, raid.Size, strings.ToUpper(raid.Dg), strings.ToUpper(raid.OsDevice))
						case "lvm":
							// Show volumeGroup info
							for _, volumeGroup := range volumeGroups {
								if !slices.Contains(volumeGroupListOfShownVolumeGroups, volumeGroup.Name) {
									if strings.Split(raid.OsDevice, "/")[0] == volumeGroup.Name {
										color.Red("   Volume Group: %s  %s Size: %s", strings.ToUpper(volumeGroup.Name), volumeGroup.State, volumeGroup.Size)
										color.Blue("     Disks:")
										volumeGroupListOfShownVolumeGroups = append(volumeGroupListOfShownVolumeGroups, volumeGroup.Name)
										// LVM disks are part of the VG not RAID as usually, so we show disks when VG is shown
										for _, disk := range raid.Disks {
											if (disk.State == "Optimal(OPT)" || disk.State == "Onln" || disk.State == "Online" || disk.State == "Good" || disk.State == "ONLINE") && disk.OsDevice != "[unknown]" {
												color.Green("       %s%s   Size: %s   Model: %s - %s/%s -> SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
											} else {
												color.Red("       %s%s   Size: %s   Model: %s - %s/%s -> SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
											}
										}
										shownLvmsHeader = false
									}
								}
							}
							if !shownLvmsHeader {
								color.Blue("     LVMs:")
								shownLvmsHeader = true
							}
							if raid.OsDevice == "NONE" {
								color.Red("        %s%s: %s   Size: %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), strings.ToUpper(raid.State), raid.Size)
							} else {
								color.Red("        %s%s: %s   Size: %s   => %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), strings.ToUpper(raid.State), raid.Size, strings.ToUpper(raid.OsDevice))
							}
						case "motherboard":
							// NOOP
						// HW Raid
						default:
							if raid.RaidLevel > 0 {
								color.Red("   %s%s: %s   Size: %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), raid.State, raid.Size)
							} else {
								color.Red("   %s%s: %s   Size: %s   => %s\n", raidLevelTabs, strings.ToUpper(raid.RaidType), raid.State, raid.Size, strings.ToUpper(raid.OsDevice))
							}
						}
					}

					for _, disk := range raid.Disks {
						if disk.State == "Optimal(OPT)" || disk.State == "Onln" || disk.State == "Online" || disk.State == "Good" || disk.State == "ONLINE" {
							switch controller.Manufacturer {
							case "mega":
								color.Green("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber)
							case "perc":
								color.Green("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber)
							case "sas2ircu":
								color.Green("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber)
							case "adaptec":
								color.Green("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber)
							case "mdadm":
								// Bogus disk
								if disk.Size == "Unknown" && disk.Model == "Unknown" && disk.Intf == "Unknown" && disk.Medium == "Unknown" && disk.SerialNumber == "Unknown" {
									color.Red("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s => %s Disk seems to be bogus.\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
								} else {
									color.Green("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
								}
							case "zfs":
								color.Green("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
							case "btrfs":
								color.Green("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
							case "lvm":
								// Disks show in raid check due to disks owning to VG not LVMs
								//color.Green("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s => %s\n", raidLevelTabs, disk.State, disk.size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
								break
							case "motherboard":
								color.Green("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
							default:
								color.Green("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
							}
						} else {
							switch controller.Manufacturer {
							case "mega":
								color.Red("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber)
							case "perc":
								color.Red("       %s%s   Size: %s   Model: %s - %s/%s  - SN: %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber)
							case "sas2ircu":
								color.Red("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber)
							case "adaptec":
								color.Red("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber)
							case "mdadm":
								color.Red("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
							case "zfs":
								color.Red("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
							case "btrfs":
								color.Red("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
							case "lvm":
								// Disks show in raid check due to disks owning to VG not LVMs
								//color.Red("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
								break
							case "motherboard":
								color.Red("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
							default:
								color.Red("       %s%s   Size: %s   Model: %s - %s/%s - SN: %s => %s\n", raidLevelTabs, disk.State, disk.Size, disk.Model, disk.Intf, disk.Medium, disk.SerialNumber, strings.ToUpper(disk.OsDevice))
							}
						}
					}
				}
			}

			// Check if current controller has any noRaidDisk:
			noRaidDisksFound := false
			if len(noRaidDisks) > 0 {
				for _, noRaidDisk := range noRaidDisks {
					if noRaidDisk.ControllerId == controller.Id {
						noRaidDisksFound = true
					}
				}
			}

			// Show NO-RAID disks
			if noRaidDisksFound {
				color.Blue("   NO-RAID disks:")
				for _, noRaidDisk := range noRaidDisks {
					if noRaidDisk.State == "Optimal (OPT)" || noRaidDisk.State == "Ready(RDY)" || noRaidDisk.State == "UGood" || noRaidDisk.State == "JBOD" {
						color.Green("       %s   Size: %s   Model: %s - %s/%s -> SN: %s => %s\n", noRaidDisk.State, noRaidDisk.Size, noRaidDisk.Model, noRaidDisk.Intf, noRaidDisk.Medium, noRaidDisk.SerialNumber, strings.ToUpper(noRaidDisk.OsDevice))
					} else {
						color.Red("       %s   Size: %s   Model: %s - %s/%s -> SN: %s => %s\n", noRaidDisk.State, noRaidDisk.Size, noRaidDisk.Model, noRaidDisk.Intf, noRaidDisk.Medium, noRaidDisk.SerialNumber, strings.ToUpper(noRaidDisk.OsDevice))
					}
				}
			}
		}
	}
	return nil
}
