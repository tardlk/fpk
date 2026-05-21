package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	socketPath := os.Getenv("FNET_SOCKET")
	if socketPath == "" {
		socketPath = "/var/apps/fnet/target/fnet.sock"
	}

	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen on %s: %v\n", socketPath, err)
		os.Exit(1)
	}
	defer os.Remove(socketPath)

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/api/bbr", handleBBR)
	mux.HandleFunc("/api/hosts", handleHosts)

	// Static files
	staticFS, _ := fs.Sub(staticFiles, "static")
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	server := &http.Server{Handler: mux}
	fmt.Printf("FNet listening on %s\n", socketPath)
	if err := server.Serve(listener); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

type BBRStatus struct {
	Enabled bool   `json:"enabled"`
	CongAlg string `json:"congAlg"`
	Qdisc   string `json:"qdisc"`
}

type HostsData struct {
	Content string `json:"content"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func handleBBR(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		congAlg, _ := sysctlGet("net.ipv4.tcp_congestion_control")
		qdisc, _ := sysctlGet("net.core.default_qdisc")
		// Check persistent config
		bbrPersist := checkSysctlConf("net.ipv4.tcp_congestion_control=bbr") &&
			checkSysctlConf("net.core.default_qdisc=fq")

		status := BBRStatus{
			Enabled: congAlg == "bbr" && bbrPersist,
			CongAlg: congAlg,
			Qdisc:   qdisc,
		}
		writeJSON(w, APIResponse{Success: true, Data: status})

	case "POST":
		var req struct {
			Enable bool `json:"enable"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, APIResponse{Success: false, Message: "无效请求"})
			return
		}

		if req.Enable {
			// Enable BBR
			if err := writeSysctlConf("net.core.default_qdisc", "fq"); err != nil {
				writeJSON(w, APIResponse{Success: false, Message: "写入 sysctl.conf 失败: " + err.Error()})
				return
			}
			if err := writeSysctlConf("net.ipv4.tcp_congestion_control", "bbr"); err != nil {
				writeJSON(w, APIResponse{Success: false, Message: "写入 sysctl.conf 失败: " + err.Error()})
				return
			}
			// Check if BBR module is available
			if out, err := exec.Command("modprobe", "tcp_bbr").CombinedOutput(); err != nil {
				writeJSON(w, APIResponse{Success: false, Message: "BBR 模块加载失败: " + string(out)})
				return
			}
			exec.Command("sysctl", "-p").Run()
			// Also apply immediately
			exec.Command("sysctl", "-w", "net.core.default_qdisc=fq").Run()
			exec.Command("sysctl", "-w", "net.ipv4.tcp_congestion_control=bbr").Run()
			writeJSON(w, APIResponse{Success: true, Message: "BBR 已开启，重启后仍然生效"})
		} else {
			// Disable BBR
			removeSysctlConf("net.core.default_qdisc=fq")
			removeSysctlConf("net.ipv4.tcp_congestion_control=bbr")
			exec.Command("sysctl", "-p").Run()
			// Switch back to cubic
			exec.Command("sysctl", "-w", "net.ipv4.tcp_congestion_control=cubic").Run()
			writeJSON(w, APIResponse{Success: true, Message: "BBR 已关闭"})
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleHosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		data, err := os.ReadFile("/etc/hosts")
		if err != nil {
			writeJSON(w, APIResponse{Success: false, Message: "读取 hosts 失败: " + err.Error()})
			return
		}
		writeJSON(w, APIResponse{Success: true, Data: HostsData{Content: string(data)}})

	case "POST":
		var req HostsData
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, APIResponse{Success: false, Message: "无效请求"})
			return
		}
		if err := os.WriteFile("/etc/hosts", []byte(req.Content), 0644); err != nil {
			writeJSON(w, APIResponse{Success: false, Message: "写入 hosts 失败: " + err.Error()})
			return
		}
		writeJSON(w, APIResponse{Success: true, Message: "hosts 已保存"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func sysctlGet(key string) (string, error) {
	out, err := exec.Command("sysctl", "-n", key).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func checkSysctlConf(line string) bool {
	data, err := os.ReadFile("/etc/sysctl.conf")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), line)
}

func writeSysctlConf(key, value string) error {
	line := key + "=" + value
	if checkSysctlConf(line) {
		return nil
	}
	f, err := os.OpenFile("/etc/sysctl.conf", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "\n%s\n", line)
	return err
}

func removeSysctlConf(line string) {
	data, _ := os.ReadFile("/etc/sysctl.conf")
	lines := strings.Split(string(data), "\n")
	var newLines []string
	for _, l := range lines {
		if strings.TrimSpace(l) == line {
			continue
		}
		newLines = append(newLines, l)
	}
	os.WriteFile("/etc/sysctl.conf", []byte(strings.Join(newLines, "\n")), 0644)
}

func writeJSON(w http.ResponseWriter, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
