package utils

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetClientIP(t *testing.T) {
	Convey("IP提取工具测试", t, func() {

		Convey("当请求为nil时", func() {
			ip := GetClientIP(nil)
			So(ip, ShouldEqual, "")
		})

		Convey("从X-Real-IP头提取IP", func() {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("X-Real-IP", "203.0.113.1")

			ip := GetClientIP(req)
			So(ip, ShouldEqual, "203.0.113.1")
		})

		Convey("从X-Forwarded-For头提取IP", func() {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.2")

			ip := GetClientIP(req)
			So(ip, ShouldEqual, "203.0.113.1")
		})

		Convey("当X-Forwarded-For包含私有IP时跳过", func() {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.1, 203.0.113.1")

			ip := GetClientIP(req)
			So(ip, ShouldEqual, "203.0.113.1")
		})

		Convey("从CF-Connecting-IP头提取IP", func() {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("CF-Connecting-IP", "203.0.113.1")

			ip := GetClientIP(req)
			So(ip, ShouldEqual, "203.0.113.1")
		})

		Convey("从RemoteAddr提取IP", func() {
			req, _ := http.NewRequest("GET", "/", nil)
			req.RemoteAddr = "203.0.113.1:12345"

			ip := GetClientIP(req)
			So(ip, ShouldEqual, "203.0.113.1")
		})

		Convey("X-Real-IP优先级最高", func() {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("X-Real-IP", "203.0.113.1")
			req.Header.Set("X-Forwarded-For", "198.51.100.2")
			req.RemoteAddr = "192.0.2.1:12345"

			ip := GetClientIP(req)
			So(ip, ShouldEqual, "203.0.113.1")
		})

		Convey("处理无效IP地址", func() {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("X-Real-IP", "invalid-ip")
			req.RemoteAddr = "203.0.113.1:12345"

			ip := GetClientIP(req)
			So(ip, ShouldEqual, "203.0.113.1")
		})

		Convey("无法获取IP时返回unknown", func() {
			req, _ := http.NewRequest("GET", "/", nil)

			ip := GetClientIP(req)
			So(ip, ShouldEqual, "unknown")
		})
	})
}

func TestIsValidIP(t *testing.T) {
	Convey("IP地址验证测试", t, func() {

		Convey("有效的IPv4地址", func() {
			So(isValidIP("192.168.1.1"), ShouldBeTrue)
			So(isValidIP("203.0.113.1"), ShouldBeTrue)
			So(isValidIP("127.0.0.1"), ShouldBeTrue)
		})

		Convey("有效的IPv6地址", func() {
			So(isValidIP("2001:db8::1"), ShouldBeTrue)
			So(isValidIP("::1"), ShouldBeTrue)
		})

		Convey("无效的IP地址", func() {
			So(isValidIP(""), ShouldBeFalse)
			So(isValidIP("invalid"), ShouldBeFalse)
			So(isValidIP("300.300.300.300"), ShouldBeFalse)
		})
	})
}

func TestIsPrivateIP(t *testing.T) {
	Convey("私有IP检测测试", t, func() {

		Convey("私有IPv4地址", func() {
			So(isPrivateIP("10.0.0.1"), ShouldBeTrue)
			So(isPrivateIP("172.16.0.1"), ShouldBeTrue)
			So(isPrivateIP("192.168.1.1"), ShouldBeTrue)
			So(isPrivateIP("127.0.0.1"), ShouldBeTrue)
		})

		Convey("公共IPv4地址", func() {
			So(isPrivateIP("203.0.113.1"), ShouldBeFalse)
			So(isPrivateIP("8.8.8.8"), ShouldBeFalse)
		})

		Convey("私有IPv6地址", func() {
			So(isPrivateIP("::1"), ShouldBeTrue)
			So(isPrivateIP("fe80::1"), ShouldBeTrue)
		})

		Convey("无效IP地址", func() {
			So(isPrivateIP("invalid"), ShouldBeFalse)
			So(isPrivateIP(""), ShouldBeFalse)
		})
	})
}

func TestValidateIPAddress(t *testing.T) {
	Convey("IP地址标准化测试", t, func() {

		Convey("有效IP地址标准化", func() {
			ip, valid := ValidateIPAddress("192.168.1.1")
			So(valid, ShouldBeTrue)
			So(ip, ShouldEqual, "192.168.1.1")
		})

		Convey("无效IP地址", func() {
			ip, valid := ValidateIPAddress("invalid")
			So(valid, ShouldBeFalse)
			So(ip, ShouldEqual, "")
		})

		Convey("空IP地址", func() {
			ip, valid := ValidateIPAddress("")
			So(valid, ShouldBeFalse)
			So(ip, ShouldEqual, "")
		})
	})
}

func TestIPVersionDetection(t *testing.T) {
	Convey("IP版本检测测试", t, func() {

		Convey("IPv4地址检测", func() {
			So(IsIPv4("192.168.1.1"), ShouldBeTrue)
			So(IsIPv4("203.0.113.1"), ShouldBeTrue)
			So(IsIPv4("2001:db8::1"), ShouldBeFalse)
		})

		Convey("IPv6地址检测", func() {
			So(IsIPv6("2001:db8::1"), ShouldBeTrue)
			So(IsIPv6("::1"), ShouldBeTrue)
			So(IsIPv6("192.168.1.1"), ShouldBeFalse)
		})

		Convey("无效IP地址", func() {
			So(IsIPv4("invalid"), ShouldBeFalse)
			So(IsIPv6("invalid"), ShouldBeFalse)
		})
	})
}
