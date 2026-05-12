// Story 3.1.1: 把会员相关接口共享的正则等校验工具集中在本文件，
// 此前散落在已删除的 register_logic.go 中。

package member

import "regexp"

// mobileRegexp 中国大陆手机号正则: 1[3-9] 开头 + 9 位数字
var mobileRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)
