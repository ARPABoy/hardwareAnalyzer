package lvm

import (
	"bufio"
	"fmt"
	"hardwareAnalyzer/utils"
	"regexp"
	"strconv"
	"strings"

	human "github.com/dustin/go-humanize"

	"github.com/fatih/color"
)

var CheckLVMRaid = func() (bool, error) {
	fmt.Println("> Checking LVM RAIDs.")
	command := "lvs"
	outputStdout, outputStderr, err := utils.GetCommandOutput("lvm", "checkLVMRaid", command)
	//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())
	//fmt.Println("err: ", err)
	if err != nil {
		//color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
		return false, fmt.Errorf("Something went wrong executing command %s: %v", command, err)
	}
	if len(outputStderr.String()) != 0 {
		// Check if its a missing device error or a real error
		scanner := bufio.NewScanner(strings.NewReader(outputStderr.String()))
		missingDeviceError := false
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			//fmt.Println("line: ", line)
			if strings.Contains(line, "is missing PV") {
				missingDeviceError = true
				break
			}
			// VG using old PV header
			if strings.Contains(line, "is using an old PV header, modify the VG to update") {
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
	// We get empty output when no LVMs exist, we get inside loop only when LVMs exist.
	for scanner.Scan() {
		//line := scanner.Text()
		//fmt.Println("line: ", line)
		color.Magenta("> LVM volume detected.")
		return true, nil
	}

	fmt.Println("> No LVM volumes detected.")
	return false, nil
}

// LVM: Drives are in VG, yet LVs determine the RAID level.
var ProcessLVMRaid = func(manufacturer string) ([]utils.ControllerStruct, []utils.VolumeGroupStruct, []utils.RaidStruct, error) {
	//fmt.Println("-- processLVMRaid --")
	var controllers = []utils.ControllerStruct{}
	var volumeGroups = []utils.VolumeGroupStruct{}
	var raids = []utils.RaidStruct{}

	controller := utils.ControllerStruct{
		Id:           "lvm-0",
		Manufacturer: "lvm",
		Model:        "LVM",
		Status:       "Good",
	}
	controllers = append(controllers, controller)

	fmt.Println("> Getting current LVM configuration.")

	// VGs:
	command := "vgs --noheadings --units b -o vg_name,vg_size,vg_missing_pv_count"
	outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "processLVMRaid", command)
	//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())
	if err != nil {
		color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
		return controllers, volumeGroups, raids, fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
	}
	if len(outputStderr.String()) != 0 {
		// Check if its a missing device error or a real error
		scanner := bufio.NewScanner(strings.NewReader(outputStderr.String()))
		missingDeviceError := false
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			//fmt.Println("line: ", line)
			if strings.Contains(line, "is missing PV") {
				missingDeviceError = true
				break
			}
			// VG using old PV header
			if strings.Contains(line, "is using an old PV header, modify the VG to update") {
				missingDeviceError = true
				break
			}
		}
		if !missingDeviceError {
			color.Red("++ ERROR: Something went wrong executing command: %s.", command)
			return controllers, volumeGroups, raids, fmt.Errorf("Error: Something went wrong executing command: %s.", command)
		}
	}

	fmt.Println("> Parsing LVM data.")
	vgsScanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
	vgName := "Unknown"
	vgSize := "Unknown"
	vgMissingPvCount := 0
	vgHealth := "Unknown"
	for vgsScanner.Scan() {
		line := vgsScanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		vgName = strings.Fields(line)[0]

		// Size in human
		vgSize = strings.Fields(line)[1]
		// Remove alphas
		re := regexp.MustCompile(`\D`)
		vgSizeValue := re.ReplaceAllString(vgSize, "")
		vgSizeInt, _ := strconv.Atoi(vgSizeValue)
		vgSize = human.Bytes(uint64(vgSizeInt))

		vgMissingPvCount, _ = strconv.Atoi(strings.Fields(line)[2])
		if vgMissingPvCount == 0 {
			vgHealth = "ONLINE"
		} else {
			vgHealth = "Bad: " + strconv.Itoa(vgMissingPvCount) + " missing device."
		}
		volumeGroup := utils.VolumeGroupStruct{
			ControllerId: "lvm-0",
			Name:         vgName,
			State:        vgHealth,
			Size:         vgSize,
		}
		volumeGroups = append(volumeGroups, volumeGroup)
	}

	// LVs:
	command = "vgs --noheadings --units b -o lv_size,segtype,vg_name,lv_path"
	outputStdout, outputStderr, err = utils.GetCommandOutput(manufacturer, "processLVMRaid", command)
	//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())
	if err != nil {
		color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
		return controllers, volumeGroups, raids, fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
	}
	if len(outputStderr.String()) != 0 {
		// Check if its a missing device error or a real error
		scanner := bufio.NewScanner(strings.NewReader(outputStderr.String()))
		missingDeviceError := false
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			//fmt.Println("line: ", line)
			if strings.Contains(line, "is missing PV") {
				missingDeviceError = true
				break
			}
			// VG using old PV header
			if strings.Contains(line, "is using an old PV header, modify the VG to update") {
				missingDeviceError = true
				break
			}
		}
		if !missingDeviceError {
			color.Red("++ ERROR: Something went wrong executing command: %s.", command)
			return controllers, volumeGroups, raids, fmt.Errorf("Error: Something went wrong executing command: %s.", command)
		}
	}

	lvScanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
	lvSize := "Unknown"
	lvType := "Unknown"
	lvVg := "Unknown"
	lvPath := "Unknown"
	lvStatus := "Unknown"
	for lvScanner.Scan() {
		line := lvScanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Size in human
		lvSize = strings.Fields(line)[0]
		// Remove alphas
		re := regexp.MustCompile(`\D`)
		lvSizeValue := re.ReplaceAllString(lvSize, "")
		lvSizeValueInt, _ := strconv.Atoi(lvSizeValue)
		lvSize = human.Bytes(uint64(lvSizeValueInt))

		lvType = strings.Fields(line)[1]
		lvVg = strings.Fields(line)[2]
		if len(strings.Fields(line)) == 4 {
			lvPath = strings.Fields(line)[3]
			lvPath = strings.ReplaceAll(lvPath, "/dev/", "")
		} else {
			lvPath = "NONE"
		}
		// Search VG health to determine LV health
		for _, volumeGroup := range volumeGroups {
			if lvVg == volumeGroup.Name {
				if volumeGroup.State == "ONLINE" {
					lvStatus = volumeGroup.State
				} else {
					lvStatus = "Bad"
				}
				break
			}
		}

		raid := utils.RaidStruct{
			ControllerId: "lvm-0",
			RaidLevel:    0,
			Dg:           lvVg,
			RaidType:     lvType,
			State:        lvStatus,
			Size:         lvSize,
			OsDevice:     lvPath,
		}

		// Add PVs to raid(LV)
		command := "pvs --noheadings --units b -o pv_name,vg_name,pv_size"
		outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "processLVMRaid", command)
		//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())
		if err != nil {
			color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
			return controllers, volumeGroups, raids, fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
		}
		if len(outputStderr.String()) != 0 {
			// Check if its a missing device error or a real error
			scanner := bufio.NewScanner(strings.NewReader(outputStderr.String()))
			missingDeviceError := false
			for scanner.Scan() {
				line := scanner.Text()
				line = strings.TrimSpace(line)
				if len(line) == 0 {
					continue
				}
				//fmt.Println("line: ", line)
				if strings.Contains(line, "is missing PV") {
					missingDeviceError = true
					break
				}
				// VG using old PV header
				if strings.Contains(line, "is using an old PV header, modify the VG to update") {
					missingDeviceError = true
					break
				}
			}
			if !missingDeviceError {
				color.Red("++ ERROR: Something went wrong executing command: %s.", command)
				return controllers, volumeGroups, raids, fmt.Errorf("Error: Something went wrong executing command: %s.", command)
			}
		}

		//fmt.Println("> Parsing PVS list.")
		pvScanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
		diskSerialNumber := "Unknown"
		diskMedium := "Unknown"
		diskModel := "Unknown"
		diskIntf := "Unknown"
		for pvScanner.Scan() {
			line := pvScanner.Text()
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			//fmt.Printf("--- LINE: |%v|\n", line)
			//fmt.Println("len(line): ", len(line))

			diskPv := strings.Fields(line)[0]
			diskPv = strings.ReplaceAll(diskPv, "/dev/", "")
			//fmt.Println("diskPv: ", diskPv)
			diskVg := strings.Fields(line)[1]
			//fmt.Println("diskVg: ", diskVg)

			if diskVg == lvVg {
				// Size in human
				diskSize := strings.Fields(line)[2]
				// Remove alphas
				re := regexp.MustCompile(`\D`)
				diskSizeValue := re.ReplaceAllString(diskSize, "")
				diskSizeValueInt, _ := strconv.Atoi(diskSizeValue)
				diskSize = human.Bytes(uint64(diskSizeValueInt))

				//fmt.Println("Disk size: ", diskSize)
				diskSerialNumber, diskModel, diskIntf, diskMedium, err = utils.GetDiskData(diskPv)
				if err != nil {
					color.Red("++ ERROR: utils.GetDiskData: %s", err)
				}

				// We assume that if disk appears listed, its online
				physicalDisk := utils.DiskStruct{
					ControllerId: "lvm-0",
					Dg:           diskVg,
					State:        "ONLINE",
					Size:         diskSize,
					Intf:         diskIntf,
					Medium:       diskMedium,
					Model:        diskModel,
					SerialNumber: diskSerialNumber,
					OsDevice:     diskPv,
				}
				raid.AddDisk(physicalDisk)
				continue
			}
		}
		//spew.Dump(raid)
		raids = append(raids, raid)
	}
	return controllers, volumeGroups, raids, nil
}
