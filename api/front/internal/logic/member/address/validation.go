package address

import (
	"regexp"
	"strings"

	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/errorx"
)

var (
	addressPhoneRegexp      = regexp.MustCompile(`^1[3-9]\d{9}$`)
	addressPostalCodeRegexp = regexp.MustCompile(`^\d{6}$`)
)

func normalizeAndValidateAddressReq(req *types.AddressReq, requireID bool) error {
	if req == nil {
		return errorx.NewDefaultError("地址参数不能为空")
	}
	if requireID && req.Id <= 0 {
		return errorx.NewDefaultError("地址ID不能为空")
	}

	req.ReceiverName = strings.TrimSpace(req.ReceiverName)
	req.ReceiverPhone = strings.TrimSpace(req.ReceiverPhone)
	req.Province = strings.TrimSpace(req.Province)
	req.City = strings.TrimSpace(req.City)
	req.District = strings.TrimSpace(req.District)
	req.DetailAddress = strings.TrimSpace(req.DetailAddress)
	req.PostalCode = strings.TrimSpace(req.PostalCode)
	req.Tag = strings.TrimSpace(req.Tag)
	if req.IsDefault != 1 {
		req.IsDefault = 0
	}

	switch {
	case req.ReceiverName == "":
		return errorx.NewDefaultError("收件人姓名不能为空")
	case len([]rune(req.ReceiverName)) > 20:
		return errorx.NewDefaultError("收件人姓名不能超过20个字")
	case !addressPhoneRegexp.MatchString(req.ReceiverPhone):
		return errorx.NewDefaultError("请输入正确的11位手机号码")
	case req.Province == "":
		return errorx.NewDefaultError("省份不能为空")
	case req.City == "":
		return errorx.NewDefaultError("城市不能为空")
	case req.District == "":
		return errorx.NewDefaultError("区县不能为空")
	case req.DetailAddress == "":
		return errorx.NewDefaultError("详细地址不能为空")
	case len([]rune(req.DetailAddress)) < 4:
		return errorx.NewDefaultError("详细地址不能少于4个字")
	case len([]rune(req.DetailAddress)) > 100:
		return errorx.NewDefaultError("详细地址不能超过100个字")
	case req.PostalCode != "" && !addressPostalCodeRegexp.MatchString(req.PostalCode):
		return errorx.NewDefaultError("邮政编码需为6位数字")
	case len([]rune(req.Tag)) > 20:
		return errorx.NewDefaultError("地址标签不能超过20个字")
	}

	return nil
}
