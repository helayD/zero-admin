// Story 10.11: 真实 FISCO BCOS 3.x 链客户端 Config。
//
// 部署时需同步更新 etc/*.yaml 三服务（rpc/sms, consumer, job）。

package fisco

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	bcosclient "github.com/FISCO-BCOS/go-sdk/v3/client"
)

const (
	// DefaultGroupID FISCO BCOS 3.x 默认 group。
	DefaultGroupID = "group0"
	// DefaultChainID FISCO BCOS 3.x 默认 chain。
	DefaultChainID = "chain0"
	// DefaultTimeout 单次 RPC / 交易超时。
	DefaultTimeout = 30 * time.Second
	// PrivateKeyFilePrefix 私钥从文件加载的前缀，例如 "path:///opt/fisco/keys/sms-rpc.key"。
	PrivateKeyFilePrefix = "path://"
)

// Mode 客户端实现模式，供启动日志标识实际运行状态。Story 10.11 / M4。
type Mode string

const (
	ModeDisabled      Mode = "disabled"
	ModeInvalidConfig Mode = "invalid_config"
	ModeReal          Mode = "real"
)

// ChainTypeFisco3x 链类型标识，写入 sms_card_mint_task.chain_type / chain_status。
// Story 10.11 明确禁止使用旧 mock 标识。
const ChainTypeFisco3x = "fisco_bcos_3x"

// Config FISCO BCOS 3.x 客户端配置，与 etc/*.yaml 中的 Fisco 节对应。
type Config struct {
	// Enabled 总开关。false 时返回 disabledClient（所有调用 fail fast）。
	Enabled bool

	// Host / Port 节点 RPC 地址，3.x 推荐使用单节点，多节点由调用方做 LB。
	Host string
	Port int

	// DisableSsl true 时使用明文 WebSocket（仅本地测试用）。
	// 生产必须 false（默认），并配 CaCertPath/SdkCertPath/SdkKeyPath。
	DisableSsl bool

	// IsSMCrypto 国密模式开关。本仓库非国密合规场景，默认 false。
	IsSMCrypto bool

	// GroupID / ChainID 字符串形式（3.x 是字符串，2.x 是 int）。
	// 留空时使用 DefaultGroupID / DefaultChainID。
	GroupID string
	ChainID string

	// ContractAddr CardToken 合约地址，格式 "0x..."。
	ContractAddr string

	// ContractABI 内联 ABI JSON。优先使用此字段；为空时从 ContractABIPath 加载。
	ContractABI string
	// ContractABIPath ABI 文件绝对路径，例如 /opt/fisco/contracts/CardToken.abi
	// 或仓库相对路径 pkg/fisco/contracts/CardToken.abi。
	ContractABIPath string

	// PrivateKey 私钥。两种格式：
	//   1. hex 字符串（含可选 "0x" 前缀），32 字节
	//   2. "path://<absolute-path>" 形式，从 PEM 文件加载（console2 newAccount 输出）
	PrivateKey string

	// TLS 证书。DisableSsl=false 时三者必填，DisableSsl=true 时忽略。
	CaCertPath  string
	SdkCertPath string
	SdkKeyPath  string

	// TimeoutSeconds 单次 RPC / 交易整体超时。<=0 时取 DefaultTimeout。
	// 备注：FISCO go-sdk/v3 的 SendEncodedTransaction 内部已同步轮询 receipt，
	// 轮询间隔 / 最大次数由 SDK 控制，本层仅提供整体超时保护。
	TimeoutSeconds int64
}

// Validate 检查 Enabled=true 时的必填项。
// Enabled=false 返回 nil 不做检查。
func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if strings.TrimSpace(c.Host) == "" || c.Port <= 0 {
		return errors.New("Fisco.Host/Port 未配置")
	}
	if strings.TrimSpace(c.ContractAddr) == "" {
		return errors.New("Fisco.ContractAddr 未配置")
	}
	if !c.DisableSsl {
		for name, p := range map[string]string{
			"CaCertPath":  c.CaCertPath,
			"SdkCertPath": c.SdkCertPath,
			"SdkKeyPath":  c.SdkKeyPath,
		} {
			if strings.TrimSpace(p) == "" {
				return fmt.Errorf("Fisco TLS 启用时 %s 必填", name)
			}
		}
	}
	if strings.TrimSpace(c.PrivateKey) == "" {
		return errors.New("Fisco.PrivateKey 未配置")
	}
	if strings.TrimSpace(c.ContractABI) == "" && strings.TrimSpace(c.ContractABIPath) == "" {
		return errors.New("Fisco.ContractABI / ContractABIPath 必须配置其一")
	}
	return nil
}

func (c Config) resolveGroupID() string {
	if g := strings.TrimSpace(c.GroupID); g != "" {
		return g
	}
	return DefaultGroupID
}

func (c Config) resolveChainID() string {
	if ch := strings.TrimSpace(c.ChainID); ch != "" {
		return ch
	}
	return DefaultChainID
}

func (c Config) timeout() time.Duration {
	if c.TimeoutSeconds <= 0 {
		return DefaultTimeout
	}
	return time.Duration(c.TimeoutSeconds) * time.Second
}

// loadPrivateKey 解析 PrivateKey 字段为 secp256k1 32 字节裸私钥。
// 支持 hex 或 PEM 文件（console2 newAccount 输出格式）。
func (c Config) loadPrivateKey() ([]byte, error) {
	pk := strings.TrimSpace(c.PrivateKey)
	if pk == "" {
		return nil, errors.New("Fisco.PrivateKey 未配置")
	}
	if strings.HasPrefix(pk, PrivateKeyFilePrefix) {
		path := strings.TrimSpace(strings.TrimPrefix(pk, PrivateKeyFilePrefix))
		if path == "" {
			return nil, errors.New("path:// 后路径为空")
		}
		// SDK 提供的 PEM 解析（PKCS8 EC 私钥）。
		raw, curveName, err := bcosclient.LoadECPrivateKeyFromPEM(path)
		if err != nil {
			return nil, fmt.Errorf("加载 PEM 私钥失败 %s: %w", path, err)
		}
		if curveName != bcosclient.Secp256k1 {
			return nil, fmt.Errorf("私钥曲线 %s 不是 %s", curveName, bcosclient.Secp256k1)
		}
		return raw, nil
	}
	// hex literal，可选 0x 前缀
	pk = strings.TrimPrefix(pk, "0x")
	pk = strings.TrimPrefix(pk, "0X")
	b, err := hex.DecodeString(pk)
	if err != nil {
		return nil, fmt.Errorf("私钥 hex 解析失败: %w", err)
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("私钥应为 32 字节（64 hex 字符），实际 %d 字节", len(b))
	}
	return b, nil
}

// loadContractABI 优先返回 ContractABI 内联 JSON，否则从 ContractABIPath 读取文件内容。
func (c Config) loadContractABI() (string, error) {
	if abi := strings.TrimSpace(c.ContractABI); abi != "" {
		return abi, nil
	}
	p := strings.TrimSpace(c.ContractABIPath)
	if p == "" {
		return "", errors.New("ContractABI / ContractABIPath 必须配置其一")
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return "", fmt.Errorf("读取 ABI 文件失败 %s: %w", p, err)
	}
	abi := strings.TrimSpace(string(b))
	if abi == "" {
		return "", fmt.Errorf("ABI 文件 %s 内容为空", p)
	}
	return abi, nil
}
