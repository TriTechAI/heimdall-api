package utils

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP 从HTTP请求中提取真实客户端IP地址
// 优先级：X-Real-IP > X-Forwarded-For > RemoteAddr
func GetClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	// 1. 优先检查 X-Real-IP 头
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" && isValidIP(realIP) {
		return realIP
	}

	// 2. 检查 X-Forwarded-For 头（可能包含多个IP，取第一个）
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// X-Forwarded-For 可能包含多个IP，格式：client, proxy1, proxy2
		ips := strings.Split(forwardedFor, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if isValidIP(ip) && !isPrivateIP(ip) {
				return ip
			}
		}
	}

	// 3. 检查 CF-Connecting-IP 头（Cloudflare）
	cfIP := r.Header.Get("CF-Connecting-IP")
	if cfIP != "" && isValidIP(cfIP) {
		return cfIP
	}

	// 4. 检查 X-Forwarded 头
	forwarded := r.Header.Get("X-Forwarded")
	if forwarded != "" && isValidIP(forwarded) {
		return forwarded
	}

	// 5. 最后使用 RemoteAddr
	remoteAddr := r.RemoteAddr
	if remoteAddr != "" {
		// RemoteAddr 格式通常是 "ip:port"，需要分离出IP
		if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
			if isValidIP(host) {
				return host
			}
		}
		// 如果分离失败，直接使用（可能是IPv6格式）
		if isValidIP(remoteAddr) {
			return remoteAddr
		}
	}

	// 如果都无法获取，返回默认值
	return "unknown"
}

// isValidIP 验证IP地址是否有效
func isValidIP(ip string) bool {
	if ip == "" {
		return false
	}
	return net.ParseIP(ip) != nil
}

// isPrivateIP 检查是否为私有IP地址
func isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 检查IPv4私有地址段
	if parsedIP.To4() != nil {
		// 10.0.0.0/8
		if parsedIP.To4()[0] == 10 {
			return true
		}
		// 172.16.0.0/12
		if parsedIP.To4()[0] == 172 && parsedIP.To4()[1] >= 16 && parsedIP.To4()[1] <= 31 {
			return true
		}
		// 192.168.0.0/16
		if parsedIP.To4()[0] == 192 && parsedIP.To4()[1] == 168 {
			return true
		}
		// 127.0.0.0/8 (localhost)
		if parsedIP.To4()[0] == 127 {
			return true
		}
	}

	// 检查IPv6私有地址
	if parsedIP.IsLoopback() || parsedIP.IsLinkLocalUnicast() {
		return true
	}

	return false
}

// ValidateIPAddress 验证IP地址格式并返回标准化的IP字符串
func ValidateIPAddress(ip string) (string, bool) {
	if ip == "" {
		return "", false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", false
	}

	return parsedIP.String(), true
}

// IsIPv4 检查是否为IPv4地址
func IsIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

// IsIPv6 检查是否为IPv6地址
func IsIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() == nil
}