# aws-mfa-helper-for-windows

## 事前準備

AWS の config と credential を以下のように設定しておく

### credential

```
[tmp-profile]
aws_access_key_id = 自身のアクセスキー
aws_secret_access_key = 自身のシークレットアクセスキー

[target-profile]
aws_access_key_id = dummy
aws_secret_access_key = dummy
aws_session_token = dummy
```

### build（Windows 向け）

```
go build -o aws-mfa-helper-for-windows.exe main.go
```
