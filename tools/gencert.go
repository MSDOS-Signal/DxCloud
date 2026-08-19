// gencert.go：为 DxCloud 生产部署生成本地自签名 TLS 证书（localhost / dxcloud.local / 127.0.0.1）。
// 用法（宿主机无需安装任何工具）：
//
//	docker run --rm -v ${PWD}:/src -v ${PWD}/deploy/certs:/out -w /src golang:1.25-alpine go run tools/gencert.go
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	tmpl := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "dxcloud.local", Organization: []string{"DxCloud"}},
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost", "dxcloud.local", "*.dxcloud.local"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	_ = os.MkdirAll("/out", 0o755)

	crt, err := os.Create("/out/tls.crt")
	if err != nil {
		panic(err)
	}
	defer crt.Close()
	if err := pem.Encode(crt, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		panic(err)
	}

	keyOut, err := os.Create("/out/tls.key")
	if err != nil {
		panic(err)
	}
	defer keyOut.Close()
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		panic(err)
	}

	println("self-signed TLS cert written to /out/tls.crt /out/tls.key (CN=dxcloud.local, SAN=localhost,*.dxcloud.local,127.0.0.1,::1)")
}
