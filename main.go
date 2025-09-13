package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

const (
	PUBLIC_IP_FILENAME   = "public_ip"
	JIO_WAN_IP_FILENAME  = "jio_wan_ip"
	IPV4_RETRY_MAX       = 5
	IPV6_RETRY_MAX       = 5
	Jio_Wan_IP_RETRY_MAX = 5
)

var (
	botToken string
	chatID   string
	username string
	password string
)

func getPublicIP() string {
	doipv6 := os.Getenv("dpipv6")

	ipv4 := getIPWithRetry(IPV4_RETRY_MAX, 120, getIPv4)

	if toInt(doipv6) == 1 {
		ipv6 := getIPWithRetry(IPV6_RETRY_MAX, 120, getIPv6)

		return fmt.Sprintf("IPv4: %s\nIPv6: %s", ipv4, ipv6)
	}

	return fmt.Sprintf("IPv4: %s", ipv4)
}

func getIPWithRetry(maxRetry int, retryWaitSecs int, getIPFunc func() string) string {
	ip := ""
	for range maxRetry {
		ip = getIPFunc()
		if ip == "" {
			time.Sleep(time.Duration(retryWaitSecs) * time.Second)
		} else {
			break
		}
	}
	return ip
}

func getIPv4() string {
	respv4, err := http.Get("https://api.ipify.org")
	if err != nil {
		log.Println("Error: ", err)
		return ""
	}
	ipv4 := readHttpResponse(respv4)
	defer respv4.Body.Close()
	return ipv4
}

func getIPv6() string {
	respv6, err := http.Get("https://api6.ipify.org")
	if err != nil {
		log.Println("Error: ", err)
		return ""
	}
	ipv6 := readHttpResponse(respv6)
	defer respv6.Body.Close()
	return ipv6
}

func getJioWanIP() string {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	loginURL := "http://192.168.29.1/platform.cgi"
	data := url.Values{}
	// Login parameters
	data.Set("thispage", "index.html")
	data.Set("users.username", username)
	data.Set("users.password", password)
	data.Set("button.login.users.dashboard", "Login")

	req, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36")
	req.Header.Set("Origin", "http://192.168.29.1")
	req.Header.Set("Referer", "http://192.168.29.1/platform.cgi?page=index.html")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Login failed:", resp.Status)
		return ""
	}

	body, _ := io.ReadAll(resp.Body)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Println("Failed to parse HTML:", err)
	}

	// Logout
	logoutURL := "http://192.168.29.1/platform.cgi?page=index.html"
	logoutResp, err := client.Get(logoutURL)
	if err != nil {
		log.Println("Failed to logout:", err)
	}
	logoutResp.Body.Close()

	var text string
	nodes := doc.Find("div.wanLanBlock div.securityRow p").Nodes
	for _, n := range nodes {
		s := goquery.NewDocumentFromNode(n).Selection
		text = strings.TrimSpace(s.Text())

		if strings.HasPrefix(text, "10") {
			break
		}
	}

	return fmt.Sprintf("IPv4: %s", text)
}

func readHttpResponse(response *http.Response) string {
	text, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println("Error: ", err)
		return ""
	}
	return string(text)
}

func sendTelegramMessage(message string) error {
	urlEscapedMessage := url.PathEscape(message)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", botToken, chatID, urlEscapedMessage)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func createFileIfNotExistsAndOverwrite(filename string, text string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal("Failed to create or overwrite file:", err)
	}
	defer file.Close()

	_, err = file.WriteString(text)
	if err != nil {
		log.Fatal("Failed to write to file:", err)
	}
}

func checkIfFileContentMatches(filename string, text string) bool {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false
	}

	return string(data) == text
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	botToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID = os.Getenv("TELEGRAM_CHAT_ID")
	username = os.Getenv("username")
	password = os.Getenv("password")

	for {
		publicIP := getPublicIP()
		matches := checkIfFileContentMatches(PUBLIC_IP_FILENAME, publicIP)
		createFileIfNotExistsAndOverwrite(PUBLIC_IP_FILENAME, publicIP)

		if !matches {
			// log.Println("Public IP changed to:\n", publicIP)
			err = sendTelegramMessage(fmt.Sprintf("📡 Public IP Address changed:\n%s", publicIP))
			if err != nil {
				log.Println("Failed to send Telegram message:", err)
			}
		}

		jioWanIP := getIPWithRetry(Jio_Wan_IP_RETRY_MAX, 60, getJioWanIP)
		matches = checkIfFileContentMatches(JIO_WAN_IP_FILENAME, jioWanIP)
		createFileIfNotExistsAndOverwrite(JIO_WAN_IP_FILENAME, jioWanIP)

		if !matches {
			// log.Println("Jio WAN IP changed to:\n", jioWanIP)
			err = sendTelegramMessage(fmt.Sprintf("📡 Jio WAN IP Address changed:\n%s", jioWanIP))
			if err != nil {
				log.Println("Failed to send Telegram message:", err)
			}
		}

		time.Sleep(120 * time.Second)
	}
}

func toInt(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		log.Fatal("Conversion error:", err)
	}
	return i
}
