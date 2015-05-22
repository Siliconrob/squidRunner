package main

import "fmt"
import "os"
import "os/exec"
import "time"
import "path"
import "strings"
import "io/ioutil"

const greenText = "\x1b[32;1m%v\x1b[0m"
const redText = "\x1b[31;1m%v\x1b[0m"

const OK = " [OK]"
const FAILURE = " [FAILURE]"
const RUNNING = " [RUNNING]"
const STOPPED = " [STOPPED]"
const DEFAULT_TIMEOUT = 120

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func PidFileLocation(targetSquid string) string {
	return "/var/run/" + targetSquid + ".pid"
}

func FileCheck(targetFile string, endCondition bool, maxTime int) bool {
	waitSeconds := 0
	for endCondition == Exists(targetFile) {
		fmt.Print(".")
		time.Sleep(time.Second)
		waitSeconds = waitSeconds + 1
		if waitSeconds > maxTime {
			fmt.Print("Timeout exceeded (" +  string(maxTime) + " seconds), verify rights/permissions/sudo")
			return false
		}
	}
	return true
}

func SquidConfLocation(targetSquid string) string {
	return "/etc/squid/" + targetSquid + ".conf"
}

func SquidConfCheck(targetSquid string) bool {
	squidConf := SquidConfLocation(targetSquid)
	if !Exists(squidConf) {
		fmt.Print("\nSquid configuration file missing: " + squidConf)
		return false
	}
	return true
}

func Start(targetSquid string) bool {
	fmt.Print("Starting: " + targetSquid)
	if !SquidConfCheck(targetSquid) {
		return false;
	}
	startCmd := exec.Command("squid", "-f", SquidConfLocation(targetSquid))
	err := startCmd.Run()
	if err != nil {
		fmt.Println(err)
	}	
	return FileCheck(PidFileLocation(targetSquid), false, DEFAULT_TIMEOUT)
}

func GetPidFromFile(targetSquid string) string {
	pidFile := PidFileLocation(targetSquid)
	if Exists(pidFile) {
		content, err := ioutil.ReadFile(pidFile)
		if err != nil {
			fmt.Println(err)
		}
		return strings.Split(string(content), "\n")[0]
	}
	return ""
}

func GetPidStatusFile(targetSquid string) string {
	pid := GetPidFromFile(targetSquid)
	if pid != "" {
		return "/proc/" + pid + "/status"
	}
	return ""
}

func Stop(targetSquid string) bool {
	fmt.Print("Stopping: " + targetSquid)
	if !SquidConfCheck(targetSquid) {
		return false;
	}
	statusFile := GetPidStatusFile(targetSquid)
	if statusFile == "" {
		return true
	}
	startCmd := exec.Command("squid", "-k", "shutdown", "-f", SquidConfLocation(targetSquid))
	err := startCmd.Run()
	if err != nil {
		fmt.Println(err)
	}	
	return FileCheck(statusFile, true, DEFAULT_TIMEOUT)
}

func IsRunning(targetSquid string) bool {
	return Exists(GetPidStatusFile(targetSquid))
}

func main() {

	argsWithProg := os.Args
	if len(argsWithProg) != 3 {
		fmt.Println("Expected format: squidRunner [squid instance] [start|stop|status]")
		return;
	}

	targetSquid := path.Base(argsWithProg[1])	
	targetCmd := strings.ToLower(argsWithProg[2])
	
	switch targetCmd {
		case "start":
			if Start(targetSquid) {
				fmt.Printf(greenText, OK)
			} else {
				fmt.Printf(redText, FAILURE)
			}
			break
		case "stop":
			if Stop(targetSquid) {
				fmt.Printf(greenText, OK)
			} else {
				fmt.Printf(redText, FAILURE)
			}
		case "status":			
			fmt.Print("Status: " + targetSquid)
			if IsRunning(targetSquid) {				
				fmt.Printf(greenText, RUNNING)
			} else {
				fmt.Printf(redText, STOPPED)
			}			
			break
		default:
			fmt.Println("Unsupported command: " + targetCmd)
	}
	fmt.Println()
}
