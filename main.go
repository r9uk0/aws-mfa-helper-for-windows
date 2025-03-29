package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-ini/ini"
)

type MFADevice struct {
	SerialNumber string `json:"SerialNumber"`
}

type ListMFADevicesResponse struct {
	MFADevices []MFADevice `json:"MFADevices"`
}

type SessionTokenResponse struct {
	Credentials struct {
		AccessKeyId     string `json:"AccessKeyId"`
		SecretAccessKey string `json:"SecretAccessKey"`
		SessionToken    string `json:"SessionToken"`
		Expiration      string `json:"Expiration"`
	} `json:"Credentials"`
}

func main() {
	// Step 1: シリアルナンバー取得に使うprofile名を入力させる
	fmt.Print("Enter Enter the name of the profile used to obtain the serial number: ")
	getSerialNumberProfileReader := bufio.NewReader(os.Stdin)
	serialNumberProfile, _ := getSerialNumberProfileReader.ReadString('\n')
	serialNumberProfile = strings.TrimSpace(serialNumberProfile)
	fmt.Println(serialNumberProfile)

	// Step 2: シリアルナンバーを取得
	cmd := exec.Command("aws", "iam", "list-mfa-devices", "--profile", serialNumberProfile, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing AWS CLI:", err)
		fmt.Scanln()
		return
	}

	// Step 3: JSONをパース
	var response ListMFADevicesResponse
	err = json.Unmarshal(output, &response)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		fmt.Scanln()
		return
	}

	// Step 4: デバイス確認
	if len(response.MFADevices) == 0 {
		fmt.Println("No MFA devices found.")
		fmt.Scanln()
		return
	}
	serialNumber := response.MFADevices[0].SerialNumber
	fmt.Println("MFA Serial Number:", serialNumber)

	// Step 5: MFAコードをユーザーに入力させる
	fmt.Print("Enter MFA Code: ")
	mfaReader := bufio.NewReader(os.Stdin)
	mfaCode, _ := mfaReader.ReadString('\n')
	mfaCode = strings.TrimSpace(mfaCode)

	// Step 6: STSで一時的なセッションを取得
	cmd = exec.Command("aws", "sts", "get-session-token", "--profile", serialNumberProfile,
		"--serial-number", serialNumber, "--token-code", mfaCode)
	sessionTokenOutput, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error getting session token:", err)
		fmt.Println(string(sessionTokenOutput))
		fmt.Scanln()
		return
	}

	// Step 7: 成功したら結果を表示
	fmt.Println("Session Token Response:")
	fmt.Println(string(sessionTokenOutput))

	var sessionTokenResp SessionTokenResponse
	err = json.Unmarshal(sessionTokenOutput, &sessionTokenResp)
	if err != nil {
		fmt.Println("Error parsing session token response:", err)
		fmt.Scanln()
		return
	}

	// Step 8: ユーザー名をユーザーに入力させる
	fmt.Print("Enter username: ")
	userNameReader := bufio.NewReader(os.Stdin)
	userName, _ := userNameReader.ReadString('\n')
	userName = strings.TrimSpace(userName)
	fmt.Println(userName)

	// Step 9: AWS configファイルのパス
	awsConfigFile := `C:\Users\` + userName + `\.aws\credentials`

	// Step 10: INIファイルを読み込む
	cfg, err := ini.Load(awsConfigFile)
	if err != nil {
		fmt.Println("Error read credentials:", err)
		fmt.Scanln()
	}

	// Step 11: 更新対象のProfileを入力させる
	fmt.Print("Enter update target profile: ")
	updateProfileReader := bufio.NewReader(os.Stdin)
	updateProfile, _ := updateProfileReader.ReadString('\n')
	updateProfile = strings.TrimSpace(updateProfile)
	fmt.Println(updateProfile)

	// Step 12: [updateProfile]セクションの値を更新
	sec, err := cfg.GetSection(updateProfile)
	if err != nil {
		fmt.Println("Error update credentials:", err)
		fmt.Scanln()
	}
	sec.Key("aws_access_key_id").SetValue(sessionTokenResp.Credentials.AccessKeyId)
	sec.Key("aws_secret_access_key").SetValue(sessionTokenResp.Credentials.SecretAccessKey)
	sec.Key("aws_session_token").SetValue(sessionTokenResp.Credentials.SessionToken)

	// Step 13: ファイルを保存
	err = cfg.SaveTo(awsConfigFile)
	if err != nil {
		fmt.Println("Error parsing session token response:", err)
		fmt.Scanln()
	}

	fmt.Println("AWS config updated successfully.")

	// 終了防止
	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
}
