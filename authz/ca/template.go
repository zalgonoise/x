package ca

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/url"
	"time"
	"unsafe"

	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
)

const (
	defaultExp       int64 = 130
	defaultSub       int64 = 1
	defaultDurMonths int   = 24

	typeCertificate = "CERTIFICATE"
)

type Template struct {
	Name       pkix.Name
	DurMonth   int
	PrivateKey *ecdsa.PrivateKey

	Serial    *big.Int
	SerialExp int64
	SerialSub int64
}

func NewCertificate(t Template) (ca *x509.Certificate, cert *pem.Block, err error) {
	if t.Serial == nil {
		bigInt, err := newInt(2, t.SerialExp, t.SerialSub)
		if err != nil {
			return nil, nil, err
		}

		t.Serial = bigInt
	}

	ca = &x509.Certificate{
		SerialNumber:          t.Serial,
		Subject:               t.Name,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, t.DurMonth, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	data, err := x509.CreateCertificate(rand.Reader, ca, ca, &t.PrivateKey.PublicKey, t.PrivateKey)
	if err != nil {
		return nil, nil, err
	}

	return ca, &pem.Block{Type: typeCertificate, Bytes: data}, nil
}

func newInt(base, exp, sub int64) (*big.Int, error) {
	maximum := new(big.Int)
	maximum.Exp(big.NewInt(base), big.NewInt(exp), nil).Sub(maximum, big.NewInt(sub))

	return rand.Int(rand.Reader, maximum)
}

func toCSR(req *pb.CSR) *x509.CertificateRequest {
	csr := &x509.CertificateRequest{}

	if req.Subject != nil {
		names := req.Subject.GetNames()
		pkixNames := make([]pkix.AttributeTypeAndValue, 0, len(names))

		for i := range names {
			rawType := names[i].GetType()
			typ := make([]int, 0, len(rawType))

			for idx := range rawType {
				typ = append(typ, int(rawType[idx]))
			}

			pkixNames = append(pkixNames, pkix.AttributeTypeAndValue{
				Type:  typ,
				Value: names[i].GetValue(),
			})
		}

		extraNames := req.Subject.GetExtraNames()
		pkixExtraNames := make([]pkix.AttributeTypeAndValue, 0, len(extraNames))

		for i := range extraNames {
			rawType := extraNames[i].GetType()
			typ := make([]int, 0, len(rawType))

			for idx := range rawType {
				typ = append(typ, int(rawType[idx]))
			}

			pkixExtraNames = append(pkixExtraNames, pkix.AttributeTypeAndValue{
				Type:  typ,
				Value: extraNames[i].GetValue(),
			})
		}

		csr.Subject = pkix.Name{
			Country:            req.Subject.GetCountry(),
			Organization:       req.Subject.GetOrganization(),
			OrganizationalUnit: req.Subject.GetOrganizationalUnit(),
			Locality:           req.Subject.GetLocality(),
			Province:           req.Subject.GetProvince(),
			StreetAddress:      req.Subject.GetStreetAddress(),
			PostalCode:         req.Subject.GetPostalCode(),
			SerialNumber:       req.Subject.GetSerialNumber(),
			CommonName:         req.Subject.GetCommonName(),
			Names:              pkixNames,
			ExtraNames:         pkixExtraNames,
		}
	}

	if len(req.Extensions) > 0 {
		csr.Extensions = make([]pkix.Extension, 0, len(req.Extensions))

		for i := range req.Extensions {
			if req.Extensions[i] != nil {
				rawID := req.Extensions[i].GetId()
				id := make([]int, 0, len(rawID))

				for idx := range rawID {
					id = append(id, int(rawID[idx]))
				}

				csr.Extensions = append(csr.Extensions, pkix.Extension{
					Id:       id,
					Critical: req.Extensions[i].GetCritical(),
					Value:    req.Extensions[i].GetValue(),
				})
			}
		}
	}

	if len(req.ExtraExtensions) > 0 {
		csr.ExtraExtensions = make([]pkix.Extension, 0, len(req.ExtraExtensions))

		for i := range req.ExtraExtensions {
			if req.ExtraExtensions[i] != nil {
				rawID := req.ExtraExtensions[i].GetId()
				id := make([]int, 0, len(rawID))

				for idx := range rawID {
					id = append(id, int(rawID[idx]))
				}

				csr.ExtraExtensions = append(csr.ExtraExtensions, pkix.Extension{
					Id:       id,
					Critical: req.ExtraExtensions[i].GetCritical(),
					Value:    req.ExtraExtensions[i].GetValue(),
				})
			}
		}
	}

	if len(req.DnsNames) > 0 {
		var set bool
		for i := range req.DnsNames {
			if req.DnsNames[i] == "" {
				continue
			}

			set = true

			break
		}

		if set {
			csr.DNSNames = req.DnsNames
		}
	}

	if len(req.EmailAddresses) > 0 {
		var set bool
		for i := range req.EmailAddresses {
			if req.EmailAddresses[i] == "" {
				continue
			}

			set = true

			break
		}

		if set {
			csr.EmailAddresses = req.EmailAddresses
		}
	}

	if len(req.IpAddresses) > 0 {
		csr.IPAddresses = make([]net.IP, 0, len(req.IpAddresses))

		for i := range req.IpAddresses {
			if req.IpAddresses[i] != nil && len(req.IpAddresses[i].Ip) > 0 {
				csr.IPAddresses = append(csr.IPAddresses, req.IpAddresses[i].Ip)
			}
		}
	}

	if len(req.Uris) > 0 {
		csr.URIs = make([]*url.URL, 0, len(req.Uris))

		for i := range req.Uris {
			if req.Uris[i].UserInfo != nil {
				userinfo := struct {
					username    string
					password    string
					passwordSet bool
				}{
					username:    req.Uris[i].UserInfo.GetUsername(),
					password:    req.Uris[i].UserInfo.GetPassword(),
					passwordSet: req.Uris[i].UserInfo.GetPasswordSet(),
				}

				user := *(*url.Userinfo)(unsafe.Pointer(&userinfo))

				csr.URIs = append(csr.URIs, &url.URL{
					Scheme:      req.Uris[i].GetScheme(),
					Opaque:      req.Uris[i].GetOpaque(),
					User:        &user,
					Host:        req.Uris[i].GetHost(),
					Path:        req.Uris[i].GetPath(),
					RawPath:     req.Uris[i].GetRawPath(),
					OmitHost:    req.Uris[i].GetOmitHost(),
					ForceQuery:  req.Uris[i].GetForceQuery(),
					RawQuery:    req.Uris[i].GetRawQuery(),
					Fragment:    req.Uris[i].GetFragment(),
					RawFragment: req.Uris[i].GetRawFragment(),
				})
			}
		}
	}

	return csr
}
