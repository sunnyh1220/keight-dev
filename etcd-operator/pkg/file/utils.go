package file

import "net/url"

// ParseBackupURL 解析 backupUrl（格式：s3://my-bucket/my-dir/my-obj.db）
// Return 类型、bucket、对象名称
func ParseBackupURL(backupUrl string) (string, string, string, error) {
	u, err := url.Parse(backupUrl)
	if err != nil {
		return "", "", "", err
	}
	return u.Scheme, u.Host, u.Path[1:], nil
}
