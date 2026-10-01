package powerdns

import (
	"fmt"
	"os/exec"

	"github.com/isucon/isucon13/webapp/go/usecase"
)

type dnsRecordRegistrar struct {
	// subdomainAddress はサブドメインの A レコードに登録する IP アドレス。
	subdomainAddress string
}

func NewDNSRecordRegistrar(subdomainAddress string) usecase.DNSRecordRegistrar {
	return &dnsRecordRegistrar{subdomainAddress: subdomainAddress}
}

// AddRecord は失敗した場合、移行前のレスポンスと同じ "<コマンドの出力>: <エラー>" 形式のエラーを返す。
func (r *dnsRecordRegistrar) AddRecord(name string) error {
	// 移行前と同じく context でキャンセルしない exec.Command を使う
	if out, err := exec.Command("pdnsutil", "add-record", "u.isucon.dev", name, "A", "0", r.subdomainAddress).CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w", string(out), err)
	}
	return nil
}
