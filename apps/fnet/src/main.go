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
	"time"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	socketPath := os.Getenv("FNET_SOCKET")
	if socketPath == "" {
		socketPath = "/var/apps/fnet/target/fnet.sock"
	}

	logf("FNet starting, socket=%s", socketPath)
	logf("Kernel: %s", kernelVersion())
	logf("Current TCP congestion: %s", currentCongAlg())

	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		logf("FATAL: Failed to listen on %s: %v", socketPath, err)
		os.Exit(1)
	}
	defer os.Remove(socketPath)

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/api/bbr", handleBBR)
	mux.HandleFunc("/api/hosts", handleHosts)

	// Static files — gateway forwards /app/fnet to the socket
	staticFS, _ := fs.Sub(staticFiles, "static")
	fileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("/app/fnet/", withLogging(http.StripPrefix("/app/fnet/", fileServer)))
	mux.HandleFunc("/app/fnet", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/app/fnet/", http.StatusFound)
	})
	// Also serve at root for direct socket access
	mux.Handle("/", withLogging(fileServer))

	server := &http.Server{Handler: withLogging(mux)}
	logf("FNet listening on %s (ready)", socketPath)
	if err := server.Serve(listener); err != nil {
		logf("FATAL: Server error: %v", err)
		os.Exit(1)
	}
}

// ------- middleware -------

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logf("REQ %s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// ------- types -------

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

// ------- handlers -------

func handleBBR(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		congAlg, _ := sysctlGet("net.ipv4.tcp_congestion_control")
		qdisc, _ := sysctlGet("net.core.default_qdisc")
		bbrPersist := checkSysctlConf("net.ipv4.tcp_congestion_control=bbr") &&
			checkSysctlConf("net.core.default_qdisc=fq")

		status := BBRStatus{
			Enabled: congAlg == "bbr" && bbrPersist,
			CongAlg: congAlg,
			Qdisc:   qdisc,
		}
		logf("BBR GET -> enabled=%v cong=%s qdisc=%s", status.Enabled, status.CongAlg, status.Qdisc)
		writeJSON(w, APIResponse{Success: true, Data: status})

	case "POST":
		var req struct {
			Enable bool `json:"enable"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logf("BBR POST -> invalid request body")
			writeJSON(w, APIResponse{Success: false, Message: "无效请求"})
			return
		}

		beforeCong, _ := sysctlGet("net.ipv4.tcp_congestion_control")
		beforeQdisc, _ := sysctlGet("net.core.default_qdisc")
		logf("BBR POST enable=%v before: cong=%s qdisc=%s", req.Enable, beforeCong, beforeQdisc)

		if req.Enable {
			if err := writeSysctlConf("net.core.default_qdisc", "fq"); err != nil {
				logf("BBR POST -> write sysctl.conf fq FAILED: %v", err)
				writeJSON(w, APIResponse{Success: false, Message: "写入 sysctl.conf 失败: " + err.Error()})
				return
			}
			if err := writeSysctlConf("net.ipv4.tcp_congestion_control", "bbr"); err != nil {
				logf("BBR POST -> write sysctl.conf bbr FAILED: %v", err)
				writeJSON(w, APIResponse{Success: false, Message: "写入 sysctl.conf 失败: " + err.Error()})
				return
			}
			if out, err := exec.Command("modprobe", "tcp_bbr").CombinedOutput(); err != nil {
				logf("BBR POST -> modprobe tcp_bbr FAILED: %s", strings.TrimSpace(string(out)))
				writeJSON(w, APIResponse{Success: false, Message: "BBR 模块加载失败: " + string(out)})
				return
			}
			logf("BBR POST -> modprobe tcp_bbr OK")
			exec.Command("sysctl", "-p").Run()
			exec.Command("sysctl", "-w", "net.core.default_qdisc=fq").Run()
			exec.Command("sysctl", "-w", "net.ipv4.tcp_congestion_control=bbr").Run()

			afterCong, _ := sysctlGet("net.ipv4.tcp_congestion_control")
			afterQdisc, _ := sysctlGet("net.core.default_qdisc")
			logf("BBR POST -> DONE after: cong=%s qdisc=%s", afterCong, afterQdisc)
			writeJSON(w, APIResponse{Success: true, Message: "BBR 已开启，重启后仍然生效"})
		} else {
			removeSysctlConf("net.core.default_qdisc=fq")
			removeSysctlConf("net.ipv4.tcp_congestion_control=bbr")
			logf("BBR POST -> removed from sysctl.conf")
			exec.Command("sysctl", "-p").Run()
			exec.Command("sysctl", "-w", "net.ipv4.tcp_congestion_control=cubic").Run()

			afterCong, _ := sysctlGet("net.ipv4.tcp_congestion_control")
			logf("BBR POST -> DONE after: cong=%s", afterCong)
			writeJSON(w, APIResponse{Success: true, Message: "BBR 已关闭"})
		}

	default:
		logf("BBR -> method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleHosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		data, err := os.ReadFile("/etc/hosts")
		if err != nil {
			logf("HOSTS GET -> read FAILED: %v", err)
			writeJSON(w, APIResponse{Success: false, Message: "读取 hosts 失败: " + err.Error()})
			return
		}
		logf("HOSTS GET -> read %d bytes", len(data))
		writeJSON(w, APIResponse{Success: true, Data: HostsData{Content: string(data)}})

	case "POST":
		var req HostsData
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logf("HOSTS POST -> invalid request body")
			writeJSON(w, APIResponse{Success: false, Message: "无效请求"})
			return
		}
		logf("HOSTS POST -> writing %d bytes to /etc/hosts", len(req.Content))
		if err := os.WriteFile("/etc/hosts", []byte(req.Content), 0644); err != nil {
			logf("HOSTS POST -> write FAILED: %v", err)
			writeJSON(w, APIResponse{Success: false, Message: "写入 hosts 失败: " + err.Error()})
			return
		}
		logf("HOSTS POST -> write OK")
		writeJSON(w, APIResponse{Success: true, Message: "hosts 已保存"})

	default:
		logf("HOSTS -> method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ------- sysctl helpers -------

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
		logf("SYSCTL write: %s=%s [already present, skip]", key, value)
		return nil
	}
	f, err := os.OpenFile("/etc/sysctl.conf", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "\n%s\n", line)
	logf("SYSCTL write: %s=%s", key, value)
	return err
}

func removeSysctlConf(line string) {
	data, _ := os.ReadFile("/etc/sysctl.conf")
	lines := strings.Split(string(data), "\n")
	removed := false
	var newLines []string
	for _, l := range lines {
		if strings.TrimSpace(l) == line {
			removed = true
			continue
		}
		newLines = append(newLines, l)
	}
	os.WriteFile("/etc/sysctl.conf", []byte(strings.Join(newLines, "\n")), 0644)
	if removed {
		logf("SYSCTL remove: %s", line)
	}
}

// ------- system info -------

func kernelVersion() string {
	out, _ := os.ReadFile("/proc/version")
	return strings.TrimSpace(strings.Split(string(out), " ")[2])
}

func currentCongAlg() string {
	alg, _ := sysctlGet("net.ipv4.tcp_congestion_control")
	if alg == "" {
		alg = "unknown"
	}
	return alg
}

// ------- logging -------

func logf(format string, args ...any) {
	ts := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[%s] %s\n", ts, msg)
}

// ------- helpers -------

func writeJSON(w http.ResponseWriter, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
