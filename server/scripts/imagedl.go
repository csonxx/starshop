package scripts

import (
	"crypto/md5"
	"fmt"
	"os"
)

// imagePoolSize 本地图池大小 (public/img-pool/case_01.jpg ... case_NN.jpg)
const imagePoolSize = 33

// img 选一张本地图片 (确定性, 根据 prompt 的 hash) 并返回 /img-pool/case_NN.jpg
// 优点: 100% 不裂图, 永久有效, 体积可控
func img(prompt string) string {
	return pickFromPool(prompt, 0)
}

// imgW 大尺寸（banner）也用同一池
func imgW(prompt string) string {
	return pickFromPool(prompt, 0)
}

// pickFromPool 内部: 根据 prompt+salt 散列到 33 张本地真实 Unsplash 图
func pickFromPool(prompt string, salt int) string {
	if prompt == "" {
		return ""
	}
	sum := md5.Sum([]byte(fmt.Sprintf("%s|%d", prompt, salt)))
	idx := (int(sum[0])<<24 | int(sum[1])<<16 | int(sum[2])<<8 | int(sum[3])) % imagePoolSize
	if idx < 0 {
		idx = -idx
	}
	return fmt.Sprintf("/img-pool/case_%02d.jpg", idx+1)
}

// removeAll / mkdirAll 给 seed.go 用于清理旧的 uploads 目录
func removeAll(p string) error      { return os.RemoveAll(p) }
func mkdirAll(p string) error       { return os.MkdirAll(p, 0755) }