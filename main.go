package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type MachineInfo struct {
	Hostname     string `json:"hostname"`
	IP           string `json:"ip"`
	SerialNumber string `json:"serial_number"`
	UltraID      string `json:"ultra_id"`
	Time         string `json:"time"`
}

func getIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() &&
			ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "unknown"
}

func getSerialNumber() string {
	if runtime.GOOS != "windows" {
		return "unknown"
	}
	out, err := exec.Command("wmic", "bios", "get", "serialnumber").CombinedOutput()
	if err != nil {
		return "unknown"
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) >= 2 {
		return strings.TrimSpace(lines[1])
	}
	return "unknown"
}

func getUltraViewerID() string {
	if runtime.GOOS != "windows" {
		return "unknown"
	}

	regPaths := []string{
		`HKEY_LOCAL_MACHINE\SOFTWARE\WOW6432Node\UltraViewer`,
		`HKEY_LOCAL_MACHINE\SOFTWARE\UltraViewer`,
		`HKEY_CURRENT_USER\SOFTWARE\UltraViewer`,
	}
	for _, path := range regPaths {
		out, err := exec.Command("reg", "query", path, "/f", "PreferID").CombinedOutput()
		if err == nil {
			re := regexp.MustCompile(`PreferID\s+REG_SZ\s+(\d+)`)
			match := re.FindStringSubmatch(string(out))
			if len(match) >= 2 {
				return match[1]
			}
		}
	}

	filePaths := []string{
		`C:\ProgramData\UltraViewer\ultraviewer.ini`,
		os.Getenv("APPDATA") + `\UltraViewer\ultraviewer.ini`,
		os.Getenv("ProgramFiles") + `\UltraViewer\ultraviewer.ini`,
		os.Getenv("ProgramFiles(x86)") + `\UltraViewer\ultraviewer.ini`,
	}
	for _, path := range filePaths {
		data, err := os.ReadFile(path)
		if err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "id=") {
					return strings.TrimSpace(strings.TrimPrefix(line, "id="))
				}
			}
		}
	}

	return "unknown"
}

func writeLog(msg string) {
	f, err := os.OpenFile("reporter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Log error:", err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, msg)
	f.WriteString(logLine)
}

func collectInfo() MachineInfo {
	hostname, _ := os.Hostname()
	return MachineInfo{
		Hostname:     hostname,
		IP:           getIP(),
		SerialNumber: getSerialNumber(),
		UltraID:      getUltraViewerID(),
		Time:         time.Now().Format(time.RFC3339),
	}
}

func main() {
	var prevInfo MachineInfo

	for {
		currentInfo := collectInfo()

		// So sánh bỏ qua Time để tránh false change
		tmpPrev := prevInfo
		tmpCurr := currentInfo
		tmpPrev.Time = ""
		tmpCurr.Time = ""

		if !reflect.DeepEqual(tmpPrev, tmpCurr) {
			jsonData, _ := json.Marshal(currentInfo)
			writeLog(fmt.Sprintf("Changed info: %s", string(jsonData)))

			_, err := http.Post("http://your-api.com/api/report", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				writeLog(fmt.Sprintf("Send error: %s", err.Error()))
			} else {
				writeLog("Data sent successfully.")
				prevInfo = currentInfo
			}
		} else {
			writeLog("No change, skip sending.")
		}

		time.Sleep(1 * time.Minute)
	}
}
