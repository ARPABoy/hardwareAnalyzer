package utils

import (
	"bufio"
	"bytes"
	_ "embed"
	"io"
	"os"
	"os/user"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/fatih/color"

	"github.com/Masterminds/semver"
)

// Test ClearString() function with input string:
func TestClearString(t *testing.T) {
	testString := "Alfaexploit\tis fucking awesome"
	msg := ClearString(testString)
	msgWanted := "Alfaexploitisfuckingawesome"
	if msg != msgWanted {
		t.Fatalf(`TestClearString: %v should match for %v`, msg, msgWanted)
	}
}

// Test ClearString() function with empty input string:
func TestClearStringEmpty(t *testing.T) {
	testString := ""
	msg := ClearString(testString)
	msgWanted := ""
	if msg != msgWanted {
		t.Fatalf(`TestClearStringEmpty: %v, want match for %v`, msg, msgWanted)
	}
}

// Test WriteExecutableFile
func TestWriteExecutableFile(t *testing.T) {
	raidBinaryPath := "/tmp/hardwareAnalyzerBin"
	RemoveFile("TestWriteExecutableFile")

	// Write file
	err := WriteExecutableFile(Storcli)
	if err != nil {
		t.Fatalf(`TestWriteExecutableFile: WriteExecutableFile returned error: %v`, err)
	}

	// Check if written file exists
	fileInfo, err := os.Stat(raidBinaryPath)
	if os.IsNotExist(err) {
		t.Fatalf(`TestWriteExecutableFile: File %s was not written.`, raidBinaryPath)
	}

	fileMode := fileInfo.Mode().Perm()
	if fileMode != 0700 {
		t.Fatalf(`TestWriteExecutableFile: File %s with wrong permissions: %v.`, raidBinaryPath, fileMode)
	}

	// Remove file
	if err := os.Remove(raidBinaryPath); err != nil {
		t.Fatalf(`TestWriteExecutableFile: Error removing file %s: %s.`, raidBinaryPath, err)
	}

	// Check it has been removed
	if _, err := os.Stat(raidBinaryPath); err == nil {
		t.Fatalf(`TestWriteExecutableFile: File %s was not removed.`, raidBinaryPath)
	}
}

// Test CopySystemBinary
func TestCopySystemBinary(t *testing.T) {
	hostsFile := "/etc/hosts"
	dstFile := "/tmp/hardwareAnalyzerBin"

	if err := CopySystemBinary(hostsFile); err != nil {
		t.Fatalf(`TestCopySystemBinary: Error executing CopySystemBinary file: %v .`, hostsFile)
	}

	f1, err := os.ReadFile(hostsFile)
	if err != nil {
		t.Fatalf(`TestCopySystemBinary: Error reading file: %v .`, hostsFile)
	}

	f2, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf(`TestCopySystemBinary: Error reading file: %v .`, dstFile)
	}

	// Compare f1 VS f2
	if !bytes.Equal(f1, f2) {
		t.Fatalf(`TestCopySystemBinary: CopySystemBinary failed, files are not equal.`)
	}
}

// Test CopySystemBinary inexistent srcFile
func TestCopySystemBinaryInexistentFile(t *testing.T) {
	srcFile := "/etc/AAA"

	if err := CopySystemBinary(srcFile); err == nil {
		t.Fatalf(`TestCopySystemBinaryInexistentFile: CopySystemBinary must return an error: "++ ERROR opening source file:".`)
	}
}

// Test RemoveFile
func TestRemoveFile(t *testing.T) {
	raidBinaryFile := "/tmp/hardwareAnalyzerBin"
	raidBinary := Storcli

	// Remove file, just in case. If file doesnt exists, error is returned, so ignore it
	_ = os.Remove(raidBinaryFile)
	// Remove inexistent file because it was removed in the previous code line
	if err := RemoveFile("TestRemoveFile"); err != nil {
		t.Fatalf(`TestRemoveFile returned: %v != nil`, err)
	}

	// Create file
	if err := WriteExecutableFile(raidBinary); err != nil {
		t.Fatalf(`TestRemoveFile error writting file: %v`, err)
	}
	// Check if written file exists
	_, err := os.Stat(raidBinaryFile)
	if os.IsNotExist(err) {
		t.Fatalf(`TestRemoveFile: File %s was not written.`, raidBinaryFile)
	}
	// Remove file
	if err := RemoveFile("TestRemoveFile"); err != nil {
		t.Fatalf(`TestRemoveFile returned: %v != nil`, err)
	}
}

// Test Str2uint64
func TestStr2uint64(t *testing.T) {
	testString := "69"
	msg, err := Str2uint64(testString)
	if err != nil {
		t.Fatalf(`Teststr2uint64: Error: %v.`, err)
	}

	var msgWanted uint64 = 69
	if msg != msgWanted {
		t.Fatalf(`Teststr2uint64: %v should match %v`, msg, msgWanted)
	}
}

// Test Str2uint64 incorrect arg
func TestStr2uint64IncorrectArg(t *testing.T) {
	testString := "69000000000000000000000000000000000000000000000000000000000000000"
	_, err := Str2uint64(testString)
	if err == nil {
		t.Fatalf(`Teststr2uint64: Returned nil err: %v.`, err)
	}
}

// Test Int8ToStr
func TestInt8ToStr(t *testing.T) {
	// ASCII values of: AlfaExploit
	asciiValues := []int8{65, 108, 102, 97, 69, 120, 112, 108, 111, 105, 116}
	stringConvertedArray := Int8ToStr(asciiValues)

	stringConvertedArrayWanted := "AlfaExploit"
	if stringConvertedArray != stringConvertedArrayWanted {
		t.Fatalf(`TestInt8ToStr - int8ToStr returned string %v != %v.`, stringConvertedArray, stringConvertedArrayWanted)
	}
}

// Test IsRoot
func TestIsRoot(t *testing.T) {
	currentUser, err := user.Current()

	if err != nil {
		t.Fatalf(`TestIsRoot: Unable to get current user: %v.`, currentUser)
	}

	checkIfRoot := IsRoot()

	if currentUser.Username == "root" && !checkIfRoot {
		t.Fatalf(`TestIsRoot: User detection failed, current user: %v - checkIfRoot: %v.`, currentUser.Username, checkIfRoot)
	}

	if currentUser.Username != "root" && checkIfRoot {
		t.Fatalf(`TestIsRoot: User detection failed, current user: %v - checkIfRoot: %v.`, currentUser.Username, checkIfRoot)
	}
}

// Test SupportedOS
func TestSupportedOS(t *testing.T) {
	// Get OS
	OS := runtime.GOOS
	supported, err := SupportedOS()
	if err != nil {
		t.Fatalf(`TestSupportedOS: Error calling supportedOS: %v.`, err)
	}
	if OS == "linux" && !supported {
		t.Fatalf(`TestSupportedOS: OS detection failed, current OS: %v - supported: %v.`, OS, supported)
	}
	if OS != "linux" && supported {
		t.Fatalf(`TestSupportedOS: OS detection failed, current OS: %v - supported: %v.`, OS, supported)
	}
}

// Test CheckMemExecKernelSupport
func TestCheckMemExecKernelSupport(t *testing.T) {
	minimumVersion, err := semver.NewConstraint(">= 3.17")
	if err != nil {
		t.Fatalf(`TestCheckMemExecKernelSupport: Error calling semver.NewConstraint: %v.`, err)
	}

	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		t.Fatalf(`TestCheckMemExecKernelSupport: Error calling syscall.Uname: %s.`, err)
	}

	currentKernel := Int8ToStr(uname.Release[:])
	currentKernelSplitted := strings.Split(currentKernel, ".")
	currentKernel = currentKernelSplitted[0] + "." + currentKernelSplitted[1]

	currentVersion, err := semver.NewVersion(currentKernel)
	if err != nil {
		t.Fatalf(`TestCheckMemExecKernelSupport: Error calling semver.NewVersio: %e.`, err)
	}

	validKernel, _ := minimumVersion.Validate(currentVersion)

	validKernelFunction, currentKernelFunction, err := CheckMemExecKernelSupport()
	if err != nil {
		t.Fatalf(`TestCheckMemExecKernelSupport: Error calling checkKernelVersion: %v`, err)
	}

	if validKernel != validKernelFunction {
		t.Fatalf(`TestCheckMemExecKernelSupport - checkKernelVersion: valid kernel values missmatch: %v - %v`, validKernel, validKernelFunction)
	}

	if currentKernel != currentKernelFunction {
		t.Fatalf(`TestCheckMemExecKernelSupport - checkKernelVersion: kernel version missmatch: %v - %v`, currentKernel, currentKernelFunction)
	}
}

// Test GetDiskPartitionSize
func TestGetDiskPartitionSize(t *testing.T) {
	// Search for first existent drive device
	readFile, err := os.Open("/proc/partitions")
	if err != nil {
		t.Fatalf(`TestGetDiskPartitionSize: Cant open /proc/partitions.`)
	}

	defer readFile.Close()
	scanner := bufio.NewScanner(readFile)
	driveToTest := "Unknown"
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		partitionDevice := strings.Fields(line)[3]
		matched, err := regexp.MatchString(`sd[a-z]+\d*`, partitionDevice)
		if err != nil {
			t.Fatalf(`TestGetDiskPartitionSize: Regexp error.`)
		}
		if matched {
			driveToTest = partitionDevice
			break
		}
	}

	// No suitable drive
	if driveToTest == "Unknown" {
		t.Fatalf(`TestGetDiskPartitionSize: Couldnt get a suitable drive to test.`)
	}

	// Get device size
	diskSize, err := GetDiskPartitionSize(driveToTest)
	if err != nil || diskSize == "Unknown" {
		t.Fatalf(`TestGetDiskPartitionSize: Couldnt get drive: %v size.`, driveToTest)
	}
}

// Test GetDiskPartitionSize incorrect drive name
func TestGetDiskPartitionSizeIncorrectDriveName(t *testing.T) {
	driveToTest := "nonExistentDrive"

	// Query device size
	diskSize, err := GetDiskPartitionSize(driveToTest)

	if err != nil {
		t.Fatalf(`TestGetDiskPartitionSizeIncorrectDriveName: Error happened while executing getDiskPartitionSize: %v.`, err)
	}

	if diskSize != "Unknown" {
		t.Fatalf(`TestGetDiskPartitionSizeIncorrectDriveName: getDiskPartitionSize returned value != Unknown with inexistent drive: %v -> %v size.`, driveToTest, diskSize)
	}
}

// Test GetDiskPartitionInterface
func TestGetDiskPartitionInterface(t *testing.T) {
	// Get a valid diskDataRaw string from /dev/disk/by-id/
	files, err := os.ReadDir("/dev/disk/by-id/")
	if err != nil {
		t.Fatalf(`TestGetDiskPartitionInterface: Error happened while reading /dev/disk/by-id/: %v`, err)
	}

	diskDataRaw := ""
	for _, file := range files {
		diskDataRaw = file.Name()
		break
	}

	intf, err := GetDiskPartitionInterface(diskDataRaw)
	if err != nil {
		t.Fatalf(`TestGetDiskPartitionInterface: Error calling getDiskPartitionInterface: %v`, err)
	}

	if intf != "nvme" && intf != "ata" && intf != "scsi" && intf != "SAS" && intf != "SATA" && intf != "md" && intf != "dm" {
		t.Fatalf(`TestGetDiskPartitionInterface - getDiskPartitionInterface returned unknown interface type: %v.`, intf)
	}
}

// Test GetDiskPartitionInterface incorrect disk DataRaw
func TestGetDiskPartitionInterfaceIncorrectDiskDataRaw(t *testing.T) {
	diskDataRaw := "InexistentIntf-InexistentSerialNumber"
	intf, err := GetDiskPartitionInterface(diskDataRaw)

	if err != nil {
		t.Fatalf(`TestGetDiskPartitionInterface: Error calling getDiskPartitionInterface: %v`, err)
	}

	if intf != "Unknown" {
		t.Fatalf(`TestGetDiskPartitionInterface - getDiskPartitionInterface returned value != Unknown with inexistent diskDataRaw: %v -> %v`, diskDataRaw, intf)
	}
}

// Test GetDiskData
func TestGetDiskData(t *testing.T) {
	// Search for first existent drive device
	readFile, err := os.Open("/proc/partitions")
	if err != nil {
		t.Fatalf(`TestGetDiskData: Cant open /proc/partitions.`)
	}
	defer readFile.Close()

	scanner := bufio.NewScanner(readFile)
	driveToTest := "Unknown"
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		partitionDevice := strings.Fields(line)[3]
		matched, err := regexp.MatchString(`sd[a-z]+\d*`, partitionDevice)
		if err != nil {
			t.Fatalf(`TestGetDiskData: Regexp error.`)
		}
		if matched {
			driveToTest = partitionDevice
			break
		}
	}

	// No suitable drive
	if driveToTest == "Unknown" {
		t.Fatalf(`TestGetDiskData: Couldnt get a suitable drive to test.`)
	}

	diskSerialNumber, diskModel, diskIntf, diskMedium, err := GetDiskData(driveToTest)

	if err != nil {
		t.Fatalf(`TestGetDiskData: Error hapenned while executing utilsgetDiskData: %v`, err)
	}

	// When hardware disk controller is used to manage disks, almost all disk information obtained via getDiskData is Unknown, if we get at least one parameter != Unknown, we consider valid answer
	if diskSerialNumber == "Unknown" && diskModel == "Unknown" && diskIntf == "Unknown" && diskMedium == "Unknown" {
		t.Fatalf(`TestGetDiskData - getDiskData returned some Unknown value, diskSerialNumber: %v diskModel: %v diskIntf: %v diskMedium: %v.`, diskSerialNumber, diskModel, diskIntf, diskMedium)
	}
}

// Test GetDiskData incorrect drive name
func TestGetDiskDataIncorrectDriveName(t *testing.T) {
	driveToTest := "nonExistentDrive"
	diskSerialNumber, diskModel, diskIntf, diskMedium, err := GetDiskData(driveToTest)
	if err != nil {
		t.Fatalf(`TestGetDiskDataIncorrectDriveName: Error happened while executing getDiskData: %v.`, err)
	}

	if diskSerialNumber != "Unknown" || diskModel != "Unknown" || diskIntf != "Unknown" || diskMedium != "Unknown" {
		t.Fatalf(`TestGetDiskDataIncorrectDriveName - getDiskData returned some value != Unknown - diskSerialNumber: %v diskModel: %v diskIntf: %v diskMedium: %v.`, diskSerialNumber, diskModel, diskIntf, diskMedium)
	}
}

// Test GetBinaryExecutor
func TestGetBinaryExecutor(t *testing.T) {
	manufacturerToBinary := map[string]string{
		"mega":       "storcli",
		"perc":       "perccli",
		"sas2ircu":   "sas2ircu",
		"adaptec":    "arcconf",
		"zfs":        "zpool",
		"btrfs":      "btrfs",
		"lvm":        "lvm",
		"inexistent": "Unknown",
	}

	memExecSupport, _, _ := CheckMemExecKernelSupport()

	for manufacturer, binary := range manufacturerToBinary {
		raidBinaryName, exe, err := GetBinaryExecutor(manufacturer, "TestGetBinaryExecutor")

		if memExecSupport {
			// If manufacturer != "inexistent", exe executor will be returned
			// So assure that it will be closed on function exit
			if manufacturer != "inexistent" {
				defer exe.Close()
			}

			// If memExecSupport, exe cant be nil with supported manufacturers
			if exe == nil && manufacturer != "inexistent" {
				t.Fatalf(`TestGetBinaryExecutor: Error getting binary, exe == nil with supported manufacturer %v.`, manufacturer)
			}

			// If memExecSupport, exe must be nil with unsupported manufacturer
			if exe != nil && manufacturer == "inexistent" {
				t.Fatalf(`TestGetBinaryExecutor: Error getting binary, exe != nil with unsupported manufacturer %v.`, manufacturer)
			}

			// Error testing correct manufacturer
			if err != nil && manufacturer != "inexistent" {
				t.Fatalf(`TestGetBinaryExecutor: Error calling GetBinaryExecutor: %v.`, err)
			}

			// No error testing incorrect manufacturer
			if err == nil && manufacturer == "inexistent" {
				t.Fatalf(`TestGetBinaryExecutor: Nil error calling GetBinaryExecutor with inexistent manufacturer.`)
			}

			// Incorrect binary name returned
			if raidBinaryName != binary {
				t.Fatalf(`TestGetBinaryExecutor - GetBinaryExecutor: incorrect raidBinary: %v should be: %v.`, raidBinaryName, binary)
			}
		} else {
			// If no memExecSupport, exe must be nil in all cases
			if exe != nil {
				exe.Close()
				t.Fatalf(`TestGetBinaryExecutor-NoMemExecSupport - GetBinaryExecutor: Error getting binary, exe != nil, manufacturer %v.`, manufacturer)
			}

			// Error testing correct manufacturer
			if err != nil && manufacturer != "inexistent" {
				t.Fatalf(`TestGetBinaryExecutor-NoMemExecSupport: Error calling GetBinaryExecutor: %v.`, err)
			}

			// No error testing incorrect manufacturer
			if err == nil && manufacturer == "inexistent" {
				t.Fatalf(`TestGetBinaryExecutor-NoMemExecSupport: Nil error calling GetBinaryExecutor with inexistent manufacturer.`)
			}

			// Incorrect binary file returned
			if raidBinaryName != "/tmp/hardwareAnalyzerBin" {
				if manufacturer == "inexistent" && raidBinaryName != "Unknown" {
					t.Fatalf(`TestGetBinaryExecutor-NoMemExecSupport - GetBinaryExecutor: incorrect raidBinary: %v should be: Unknown.`, raidBinaryName)
				} else {
					t.Fatalf(`TestGetBinaryExecutor-NoMemExecSupport - GetBinaryExecutor: incorrect raidBinary: %v should be: /tmp/hardwareAnalyzerBin.`, raidBinaryName)
				}
			}
		}
	}
}

// Test GetCommandOutput
func TestGetCommandOutput(t *testing.T) {
	command := "version"
	outputStdout, outputStderr, err := GetCommandOutput("btrfs", "TestGetCommandOutput", command)
	if err != nil || len(outputStderr.String()) != 0 {
		t.Fatalf(`TestGetCommandOutput: Error calling getCommandOutput: %v.`, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
	for scanner.Scan() {
		line := scanner.Text()
		lineWanted := "btrfs-progs v5.16.2"
		if line != lineWanted {
			t.Fatalf(`TestGetCommandOutput - getCommandOutput: incorrect btrfs version: %v should be: %v`, line, lineWanted)
		}
	}
}

// Test CheckPreviousDiskData
func TestCheckPreviousDiskData(t *testing.T) {

	var raids []RaidStruct
	raid := RaidStruct{
		ControllerId: "adaptec-0",
		RaidLevel:    0,
		Dg:           "0",
		State:        "Okay",
		RaidType:     "raid0",
		OsDevice:     "sdf",
	}

	disk := DiskStruct{
		ControllerId: "adaptec-0",
		Dg:           "0",
		State:        "ONLINE",
		Size:         "2.0TB",
		Intf:         "SATA",
		Medium:       "SSD",
		Model:        "SamsungSSD860",
		SerialNumber: "S3YVNB0KB13582W",
		OsDevice:     "sdf",
	}
	raid.AddDisk(disk)
	raids = append(raids, raid)

	var newRaids []RaidStruct
	newRaid := RaidStruct{
		ControllerId: "btrfs-0",
		RaidLevel:    0,
		Dg:           "39c52b77-13b6-43ac-8e16-d77ba4800e39",
		State:        "ONLINE",
		RaidType:     "single",
		OsDevice:     "sdf",
	}

	newDisk := DiskStruct{
		ControllerId: "btrfs-0",
		Dg:           "39c52b77-13b6-43ac-8e16-d77ba4800e39",
		State:        "ONLINE",
		Size:         "1.77TiB",
		Intf:         "Unknown",
		Medium:       "Unknown",
		Model:        "Unknown",
		SerialNumber: "Unknown",
		OsDevice:     "sdf",
	}
	newRaid.AddDisk(newDisk)
	newRaids = append(newRaids, newRaid)

	// fmt.Println("----------- newRaids ------------")
	// spew.Dump(newRaids)

	err := CheckPreviousDiskData(newRaids, raids)
	if err != nil {
		t.Fatalf(`TestCheckPreviousDiskData: Error calling CheckPreviousDiskData: %v.`, err)
	}

	// fmt.Println("----------- newRaids ------------")
	// spew.Dump(newRaids)

	for _, newRaid := range newRaids {
		for _, newDisk := range newRaid.Disks {
			newDiskModelWanted := "SamsungSSD860"
			if newDisk.Model != newDiskModelWanted {
				t.Fatalf(`TestCheckPreviousDiskData newDisk.Model: %v != diskModelWanted: %v`, newDisk.Model, newDiskModelWanted)
			}

			newDiskIntfWanted := "SATA"
			if newDisk.Intf != newDiskIntfWanted {
				t.Fatalf(`TestCheckPreviousDiskData newDisk.Intf: %v != diskModelWanted: %v`, newDisk.Intf, newDiskIntfWanted)
			}

			newDiskMediumWanted := "SSD"
			if newDisk.Medium != newDiskMediumWanted {
				t.Fatalf(`TestCheckPreviousDiskData newDisk.Medium: %v != diskMediumWanted: %v`, newDisk.Medium, newDiskMediumWanted)
			}

			newDiskSerialNumberWanted := "S3YVNB0KB13582W"
			if newDisk.SerialNumber != newDiskSerialNumberWanted {
				t.Fatalf(`TestCheckPreviousDiskData newDisk.SerialNumber: %v != diskSerialNumberWanted: %v`, newDisk.SerialNumber, newDiskSerialNumberWanted)
			}
		}
	}
}

// Test CheckPreviousDiskData No Disk Match
func TestCheckPreviousDiskDataNoDiskMatch(t *testing.T) {
	var raids []RaidStruct
	raid := RaidStruct{
		ControllerId: "adaptec-0",
		RaidLevel:    0,
		Dg:           "0",
		State:        "Okay",
		RaidType:     "raid0",
		OsDevice:     "sdg",
	}

	disk := DiskStruct{
		ControllerId: "adaptec-0",
		Dg:           "0",
		State:        "ONLINE",
		Size:         "2.0TB",
		Intf:         "SATA",
		Medium:       "SSD",
		Model:        "SamsungSSD860",
		SerialNumber: "S3YVNB0KB13582W",
		OsDevice:     "sdg",
	}
	raid.AddDisk(disk)
	raids = append(raids, raid)

	var newRaids []RaidStruct
	newRaid := RaidStruct{
		ControllerId: "btrfs-0",
		RaidLevel:    0,
		Dg:           "39c52b77-13b6-43ac-8e16-d77ba4800e39",
		State:        "ONLINE",
		RaidType:     "single",
		OsDevice:     "sdf",
	}

	newDisk := DiskStruct{
		ControllerId: "btrfs-0",
		Dg:           "39c52b77-13b6-43ac-8e16-d77ba4800e39",
		State:        "ONLINE",
		Size:         "1.77TiB",
		Intf:         "Unknown",
		Medium:       "Unknown",
		Model:        "Unknown",
		SerialNumber: "Unknown",
		OsDevice:     "sdf",
	}
	newRaid.AddDisk(newDisk)
	newRaids = append(newRaids, newRaid)

	// fmt.Println("----------- newRaids ------------")
	// spew.Dump(newRaids)

	err := CheckPreviousDiskData(newRaids, raids)
	if err != nil {
		t.Fatalf(`TestCheckPreviousDiskData: Error calling CheckPreviousDiskData: %v.`, err)
	}

	// fmt.Println("----------- newRaids ------------")
	// spew.Dump(newRaids)

	for _, newRaid := range newRaids {
		for _, newDisk := range newRaid.Disks {
			newDiskModelWanted := "Unknown"
			if newDisk.Model != newDiskModelWanted {
				t.Fatalf(`TestCheckPreviousDiskData newDisk.Model: %v != diskModelWanted: %v`, newDisk.Model, newDiskModelWanted)
			}

			newDiskIntfWanted := "Unknown"
			if newDisk.Intf != newDiskIntfWanted {
				t.Fatalf(`TestCheckPreviousDiskData newDisk.Intf: %v != diskModelWanted: %v`, newDisk.Intf, newDiskIntfWanted)
			}

			newDiskMediumWanted := "Unknown"
			if newDisk.Medium != newDiskMediumWanted {
				t.Fatalf(`TestCheckPreviousDiskData newDisk.Medium: %v != diskMediumWanted: %v`, newDisk.Medium, newDiskMediumWanted)
			}

			newDiskSerialNumberWanted := "Unknown"
			if newDisk.SerialNumber != newDiskSerialNumberWanted {
				t.Fatalf(`TestCheckPreviousDiskData newDisk.SerialNumber: %v != diskSerialNumberWanted: %v`, newDisk.SerialNumber, newDiskSerialNumberWanted)
			}
		}
	}
}

// Test ShowGatheredData hwraid optimal state
func TestShowGatheredDataHWRaidOptimalState(t *testing.T) {
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

	var controllers = []ControllerStruct{}
	var raids = []RaidStruct{}
	var pools = []PoolStruct{}
	var volumeGroups = []VolumeGroupStruct{}
	var noRaidDisks = []NoRaidDiskStruct{}

	controller := ControllerStruct{
		Id:           "adaptec-0",
		Manufacturer: "adaptec",
		Model:        "Adaptec 6405",
		Status:       "Optimal",
	}
	controllers = append(controllers, controller)

	raid := RaidStruct{
		ControllerId: "adaptec-0",
		RaidType:     "RAID5",
		State:        "Optimal",
		Size:         "20 TB",
		OsDevice:     "sdb",
	}

	disk := DiskStruct{
		ControllerId: "adaptec-0",
		Size:         "20 TB",
		Intf:         "SATA",
		Medium:       "SSD",
		Model:        "RANDOMMODEL",
		State:        "Online",
		SerialNumber: "RANDOMSERIALNUMBER",
	}
	raid.Disks = append(raid.Disks, disk)
	raids = append(raids, raid)

	ShowGatheredData(controllers, pools, volumeGroups, raids, noRaidDisks)

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
	for scanner.Scan() {
		line := scanner.Text()
		//fmt.Println("-- LINE: ", line)
		if strings.Contains(line, "ControllerID:") {
			controllerId := strings.Fields(line)[2]
			controllerIdWanted := "adaptec-0"
			if controllerId != controllerIdWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidOptimalState controllerId: %v != controllerIdWanted: %v`, controllerId, controllerIdWanted)
			}

			controllerModelData := strings.Split(line, ":")
			controllerModel := strings.Split(controllerModelData[1], " - ")[1]
			controllerModelWanted := "Adaptec 6405"
			if controllerModel != controllerModelWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidOptimalState controllerModel: %v != controllerModelWanted: %v`, controllerModel, controllerModelWanted)
			}

			controllerState := strings.Fields(line)[6]
			controllerStateWanted := "Optimal"
			if controllerState != controllerStateWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidOptimalState controllerState: %v != controllerStateWanted: %v`, controllerState, controllerStateWanted)
			}
			continue
		}

		if strings.Contains(line, "RAID") {
			raidType := strings.Fields(line)[0]
			raidTypeWanted := "RAID5:"
			if raidType != raidTypeWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidOptimalState raidType: %v != raidTypeWanted: %v`, raidType, raidTypeWanted)
			}

			raidState := strings.Fields(line)[1]
			raidStateWanted := "Optimal"
			if raidState != raidStateWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidOptimalState raidState: %v != raidStateWanted: %v`, raidState, raidStateWanted)
			}

			raidSize := strings.Fields(line)[3] + " " + strings.Fields(line)[4]
			raidSizeWanted := "20 TB"
			if raidSize != raidSizeWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidOptimalState raidSize: %v != raidSizeWanted: %v`, raidSize, raidSizeWanted)
			}

			raidOsDevice := strings.Fields(line)[6]
			raidOsDeviceWanted := "SDB"
			if raidOsDevice != raidOsDeviceWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidOptimalState raidOsDevice: %v != raidOsDeviceWanted: %v`, raidOsDevice, raidOsDeviceWanted)
			}
			continue
		}

		if strings.Contains(line, "Model") {
			diskState := strings.Fields(line)[0]
			diskStateWanted := "Online"
			if diskState != diskStateWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidOptimalState diskState: %v != diskStateWanted: %v`, diskState, diskStateWanted)
			}

			diskSize := strings.Fields(line)[2] + " " + strings.Fields(line)[3]
			diskSizeWanted := "20 TB"
			if diskSize != diskSizeWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidOptimalState diskSize: %v != diskSizeWanted: %v`, diskSize, diskSizeWanted)
			}

			diskModel := strings.Fields(line)[5]
			diskModelWanted := "RANDOMMODEL"
			if diskModel != diskModelWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidOptimalState diskModel: %v != diskModelWanted: %v`, diskModel, diskModelWanted)
			}

			diskIntfMedium := strings.Fields(line)[7]
			diskIntfMediumWanted := "SATA/SSD"
			if diskIntfMedium != diskIntfMediumWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidOptimalState diskIntfMedium: %v != diskIntfMediumWanted: %v`, diskIntfMedium, diskIntfMediumWanted)
			}

			diskSerialNumber := strings.Fields(line)[10]
			diskSerialNumberWanted := "RANDOMSERIALNUMBER"
			if diskSerialNumber != diskSerialNumberWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidOptimalState diskSerialNumber: %v != diskSerialNumberWanted: %v`, diskSerialNumber, diskSerialNumberWanted)
			}
			continue
		}
	}
}

// Test ShowGatheredData hwraid degraded state
func TestShowGatheredDataHWRaidDegradedState(t *testing.T) {
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

	var controllers = []ControllerStruct{}
	var raids = []RaidStruct{}
	var pools = []PoolStruct{}
	var volumeGroups = []VolumeGroupStruct{}
	var noRaidDisks = []NoRaidDiskStruct{}

	controller := ControllerStruct{
		Id:           "adaptec-0",
		Manufacturer: "adaptec",
		Model:        "Adaptec 6405",
		Status:       "Degraded",
	}
	controllers = append(controllers, controller)

	raid := RaidStruct{
		ControllerId: "adaptec-0",
		RaidType:     "RAID5",
		State:        "Degraded",
		Size:         "20 TB",
		OsDevice:     "sdb",
	}

	disk := DiskStruct{
		ControllerId: "adaptec-0",
		Size:         "20 TB",
		Intf:         "SATA",
		Medium:       "SSD",
		Model:        "RANDOMMODEL",
		State:        "Online",
		SerialNumber: "RANDOMSERIALNUMBER",
	}
	raid.Disks = append(raid.Disks, disk)
	raids = append(raids, raid)

	ShowGatheredData(controllers, pools, volumeGroups, raids, noRaidDisks)

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
	for scanner.Scan() {
		line := scanner.Text()
		//fmt.Println("-- LINE: ", line)
		if strings.Contains(line, "ControllerID:") {
			controllerId := strings.Fields(line)[2]
			controllerIdWanted := "adaptec-0"
			if controllerId != controllerIdWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidDegradedState controllerId: %v != controllerIdWanted: %v`, controllerId, controllerIdWanted)
			}

			controllerModelData := strings.Split(line, ":")
			controllerModel := strings.Split(controllerModelData[1], " - ")[1]
			controllerModelWanted := "Adaptec 6405"
			if controllerModel != controllerModelWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidDegradedState controllerModel: %v != controllerModelWanted: %v`, controllerModel, controllerModelWanted)
			}

			controllerState := strings.Fields(line)[6]
			controllerStateWanted := "Degraded"
			if controllerState != controllerStateWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidDegradedState controllerState: %v != controllerStateWanted: %v`, controllerState, controllerStateWanted)
			}
			continue
		}

		if strings.Contains(line, "RAID") {
			raidType := strings.Fields(line)[0]
			raidTypeWanted := "RAID5:"
			if raidType != raidTypeWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidDegradedState raidType: %v != raidTypeWanted: %v`, raidType, raidTypeWanted)
			}

			raidState := strings.Fields(line)[1]
			raidStateWanted := "Degraded"
			if raidState != raidStateWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidDegradedState raidState: %v != raidStateWanted: %v`, raidState, raidStateWanted)
			}

			raidSize := strings.Fields(line)[3] + " " + strings.Fields(line)[4]
			raidSizeWanted := "20 TB"
			if raidSize != raidSizeWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidDegradedState raidSize: %v != raidSizeWanted: %v`, raidSize, raidSizeWanted)
			}

			raidOsDevice := strings.Fields(line)[6]
			raidOsDeviceWanted := "SDB"
			if raidOsDevice != raidOsDeviceWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidDegradedState raidOsDevice: %v != raidOsDeviceWanted: %v`, raidOsDevice, raidOsDeviceWanted)
			}
			continue
		}

		if strings.Contains(line, "Model") {
			diskState := strings.Fields(line)[0]
			diskStateWanted := "Online"
			if diskState != diskStateWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidDegradedState diskState: %v != diskStateWanted: %v`, diskState, diskStateWanted)
			}

			diskSize := strings.Fields(line)[2] + " " + strings.Fields(line)[3]
			diskSizeWanted := "20 TB"
			if diskSize != diskSizeWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidDegradedState diskSize: %v != diskSizeWanted: %v`, diskSize, diskSizeWanted)
			}

			diskModel := strings.Fields(line)[5]
			diskModelWanted := "RANDOMMODEL"
			if diskModel != diskModelWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidDegradedState diskModel: %v != diskModelWanted: %v`, diskModel, diskModelWanted)
			}

			diskIntfMedium := strings.Fields(line)[7]
			diskIntfMediumWanted := "SATA/SSD"
			if diskIntfMedium != diskIntfMediumWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidDegradedState diskIntfMedium: %v != diskIntfMediumWanted: %v`, diskIntfMedium, diskIntfMediumWanted)
			}

			diskSerialNumber := strings.Fields(line)[10]
			diskSerialNumberWanted := "RANDOMSERIALNUMBER"
			if diskSerialNumber != diskSerialNumberWanted {
				t.Fatalf(`TestShowGatheredDataHWRaidDegradedState diskSerialNumber: %v != diskSerialNumberWanted: %v`, diskSerialNumber, diskSerialNumberWanted)
			}
			continue
		}
	}
}

// Test ShowGatheredData No Raid No Disks
func TestShowGatheredDataNoRaidNoDisks(t *testing.T) {

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

	var controllers = []ControllerStruct{}
	var raids = []RaidStruct{}
	var pools = []PoolStruct{}
	var volumeGroups = []VolumeGroupStruct{}
	var noRaidDisks = []NoRaidDiskStruct{}

	ShowGatheredData(controllers, pools, volumeGroups, raids, noRaidDisks)

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
	for scanner.Scan() {
		line := scanner.Text()
		//fmt.Println("-- LINE: ", line)
		lineWanted := "> No RAIDs detected."
		if line != lineWanted {
			t.Fatalf(`TestShowGatheredDataNoRaidNoDisks output: %v must be: %v`, line, lineWanted)
		}
	}
}

// Test ShowGatheredData Raid Without Disks
func TestShowGatheredDataRaidWithoutDisks(t *testing.T) {

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

	var controllers = []ControllerStruct{}
	var raids = []RaidStruct{}
	var pools = []PoolStruct{}
	var volumeGroups = []VolumeGroupStruct{}
	var noRaidDisks = []NoRaidDiskStruct{}

	controller := ControllerStruct{
		Id:           "adaptec-0",
		Manufacturer: "adaptec",
		Model:        "Adaptec 6405",
		Status:       "Optimal",
	}
	controllers = append(controllers, controller)

	raid := RaidStruct{
		ControllerId: "adaptec-0",
		RaidType:     "RAID5",
		State:        "Optimal",
		Size:         "20 TB",
		OsDevice:     "sdb",
	}
	raids = append(raids, raid)

	ShowGatheredData(controllers, pools, volumeGroups, raids, noRaidDisks)

	// Close w pipe
	w.Close()

	// Restore Stdout/Stderr to normal output
	os.Stdout = osStdoutOri
	os.Stderr = osStderrOri
	color.Output = colorOutputOri
	color.Error = colorErrorOri

	// Read all r pipe content
	out, _ := io.ReadAll(r)
	scanner := bufio.NewScanner(bytes.NewReader(out))
	i := 1
	for scanner.Scan() {
		line := scanner.Text()
		//fmt.Printf("LINE: |%v|\n", line)
		if i == 1 && line != "" {
			t.Fatalf(`TestShowGatheredDataRaidWithoutDisks line 1: %v must be ""`, line)
		}
		if i == 2 && line != "-- ControllerID: adaptec-0 - Adaptec 6405: Optimal" {
			t.Fatalf(`TestShowGatheredDataRaidWithoutDisks line 1: %v must be "-- ControllerID: adaptec-0 - Adaptec 6405: Optimal"`, line)
		}
		if i == 3 && line != "   RAID5: Optimal   Size: 20 TB   => SDB" {
			t.Fatalf(`TestShowGatheredDataRaidWithoutDisks line 1: %v must be "   RAID5: Optimal   Size: 20 TB   => SDB"`, line)
		}
		i++
	}
}

// Test ShowGatheredData Bogus Disks
func TestShowGatheredDataBogusDisks(t *testing.T) {

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

	var controllers = []ControllerStruct{}
	var raids = []RaidStruct{}
	var pools = []PoolStruct{}
	var volumeGroups = []VolumeGroupStruct{}
	var noRaidDisks = []NoRaidDiskStruct{}

	controller := ControllerStruct{
		Id:           "adaptec-0",
		Manufacturer: "adaptec",
		Model:        "Adaptec 6405",
		Status:       "Optimal",
	}
	controllers = append(controllers, controller)

	raid := RaidStruct{
		ControllerId: "adaptec-0",
		RaidType:     "RAID5",
		State:        "Optimal",
		Size:         "Unknown",
		OsDevice:     "sdb",
	}

	disk := DiskStruct{
		ControllerId: "adaptec-0",
		Size:         "Unknown",
		Intf:         "Unknown",
		Medium:       "Unknown",
		Model:        "Unknown",
		State:        "Optimal",
		SerialNumber: "Unknown",
	}
	raid.Disks = append(raid.Disks, disk)
	raids = append(raids, raid)

	ShowGatheredData(controllers, pools, volumeGroups, raids, noRaidDisks)

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
	for scanner.Scan() {
		line := scanner.Text()
		//fmt.Println("-- LINE: ", line)
		if strings.Contains(line, "ControllerID:") {
			controllerId := strings.Fields(line)[2]
			controllerIdWanted := "adaptec-0"
			if controllerId != controllerIdWanted {
				t.Fatalf(`TestShowGatheredDataBogusDisks controllerId: %v != controllerIdWanted: %v`, controllerId, controllerIdWanted)
			}

			controllerModelData := strings.Split(line, ":")
			controllerModel := strings.Split(controllerModelData[1], " - ")[1]
			controllerModelWanted := "Adaptec 6405"
			if controllerModel != controllerModelWanted {
				t.Fatalf(`TestShowGatheredDataBogusDisks controllerModel: %v != controllerModelWanted: %v`, controllerModel, controllerModelWanted)
			}

			controllerState := strings.Fields(line)[6]
			controllerStateWanted := "Bad"
			if controllerState != controllerStateWanted {
				t.Fatalf(`TestShowGatheredDataBogusDisks controllerState: %v != controllerStateWanted: %v`, controllerState, controllerStateWanted)
			}
			continue
		}

		if strings.Contains(line, "RAID") {
			raidType := strings.Fields(line)[0]
			raidTypeWanted := "RAID5:"
			if raidType != raidTypeWanted {
				t.Fatalf(`TestShowGatheredDataBogusDisks raidType: %v != raidTypeWanted: %v`, raidType, raidTypeWanted)
			}

			raidState := strings.Fields(line)[1]
			raidStateWanted := "Bad"
			if raidState != raidStateWanted {
				t.Fatalf(`TestShowGatheredDataBogusDisks raidState: %v != raidStateWanted: %v`, raidState, raidStateWanted)
			}

			raidSize := strings.Fields(line)[3]
			raidSizeWanted := "Unknown"
			if raidSize != raidSizeWanted {
				t.Fatalf(`TestShowGatheredDataBogusDisks raidSize: %v != raidSizeWanted: %v`, raidSize, raidSizeWanted)
			}

			raidOsDevice := strings.Fields(line)[5]
			raidOsDeviceWanted := "SDB"
			if raidOsDevice != raidOsDeviceWanted {
				t.Fatalf(`TestShowGatheredDataBogusDisks raidOsDevice: %v != raidOsDeviceWanted: %v`, raidOsDevice, raidOsDeviceWanted)
			}
			continue
		}

		if strings.Contains(line, "Model") {
			diskState := strings.Fields(line)[0]
			diskStateWanted := "Optimal"
			if diskState != diskStateWanted {
				t.Fatalf(`TestShowGatheredDataBogusDisks diskState: %v != diskStateWanted: %v`, diskState, diskStateWanted)
			}

			diskSize := strings.Fields(line)[2]
			diskSizeWanted := "Unknown"
			if diskSize != diskSizeWanted {
				t.Fatalf(`TestShowGatheredDataBogusDisks diskSize: %v != diskSizeWanted: %v`, diskSize, diskSizeWanted)
			}

			diskModel := strings.Fields(line)[4]
			diskModelWanted := "Unknown"
			if diskModel != diskModelWanted {
				t.Fatalf(`TestShowGatheredDataBogusDisks diskModel: %v != diskModelWanted: %v`, diskModel, diskModelWanted)
			}

			diskIntfMedium := strings.Fields(line)[6]
			diskIntfMediumWanted := "Unknown/Unknown"
			if diskIntfMedium != diskIntfMediumWanted {
				t.Fatalf(`TestShowGatheredDataBogusDisks diskIntfMedium: %v != diskIntfMediumWanted: %v`, diskIntfMedium, diskIntfMediumWanted)
			}

			diskSerialNumber := strings.Fields(line)[9]
			diskSerialNumberWanted := "Unknown"
			if diskSerialNumber != diskSerialNumberWanted {
				t.Fatalf(`TestShowGatheredDataBogusDisks diskSerialNumber: %v != diskSerialNumberWanted: %v`, diskSerialNumber, diskSerialNumberWanted)
			}
			continue
		}
	}
}

// Test ShowGatheredData noraiddisk
func TestShowGatheredDataNoRaidDisk(t *testing.T) {

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

	var controllers = []ControllerStruct{}
	controller := ControllerStruct{
		Id:           "mega-0",
		Manufacturer: "mega",
		Model:        "LSI MegaRAID SAS 9271-4i",
		Status:       "Optimal",
	}
	controllers = append(controllers, controller)

	var raids = []RaidStruct{}
	var pools = []PoolStruct{}

	var volumeGroups = []VolumeGroupStruct{}
	noRaidDisks := []NoRaidDiskStruct{
		{
			ControllerId: "mega-0",
			EidSlot:      "21:0",
			State:        "JBOD",
			Size:         "5.458 TB",
			Intf:         "SAS",
			Medium:       "HDD",
			Model:        "HUH728060AL5200",
			SerialNumber: "2QGA0JGX",
			OsDevice:     "JBOD-sdc",
		},
	}

	ShowGatheredData(controllers, pools, volumeGroups, raids, noRaidDisks)

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
	for scanner.Scan() {
		line := scanner.Text()
		//fmt.Println("-- LINE: ", line)

		if strings.Contains(line, "-- ControllerID:") {
			controllerId := strings.Fields(line)[2]
			controllerIdWanted := "mega-0"
			if controllerId != controllerIdWanted {
				t.Fatalf(`TestShowGatheredDataNoRaidDisk controllerId: %v != controllerIdWanted: %v`, controllerId, controllerIdWanted)
			}
			controllerModelArray := strings.Fields(line)[4:8]
			controllerModel := strings.Join(controllerModelArray, " ")
			controllerModelWanted := "LSI MegaRAID SAS 9271-4i:"
			if controllerModel != controllerModelWanted {
				t.Fatalf(`TestShowGatheredDataNoRaidDisk controllerModel: %v != controllerModelWanted: %v`, controllerModel, controllerModelWanted)
			}
		}

		if strings.Contains(line, "JBOD") {
			diskSizeArray := strings.Fields(line)[2:4]
			diskSize := strings.Join(diskSizeArray, " ")
			diskSizeWanted := "5.458 TB"
			if diskSize != diskSizeWanted {
				t.Fatalf(`TestShowGatheredDataNoRaidDisk diskSize: %v != diskSizeWanted: %v`, diskSize, diskSizeWanted)
			}
			diskModel := strings.Fields(line)[5]
			diskModelWanted := "HUH728060AL5200"
			if diskModel != diskModelWanted {
				t.Fatalf(`TestShowGatheredDataNoRaidDisk diskModel: %v != diskModelWanted: %v`, diskModel, diskModelWanted)
			}
			diskIntf := strings.Fields(line)[7]
			diskIntfWanted := "SAS/HDD"
			if diskIntf != diskIntfWanted {
				t.Fatalf(`TestShowGatheredDataNoRaidDisk diskIntf: %v != diskIntfWanted: %v`, diskIntf, diskIntfWanted)
			}
			diskSerialNumber := strings.Fields(line)[10]
			diskSerialNumberWanted := "2QGA0JGX"
			if diskSerialNumber != diskSerialNumberWanted {
				t.Fatalf(`TestShowGatheredDataNoRaidDisk diskSerialNumber: %v != diskSerialNumberWanted: %v`, diskSerialNumber, diskSerialNumberWanted)
			}
		}
	}
}
