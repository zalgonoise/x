package keygen

import (
	"crypto/ecdsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"net/url"
	"time"
	"unsafe"

	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
)

func ToCSR(name string, pub *ecdsa.PublicKey, req *pb.CSR) *x509.CertificateRequest {
	csr := &x509.CertificateRequest{
		PublicKey: pub,
	}

	if req == nil {
		req = new(pb.CSR)
	}

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

		if csr.Subject.CommonName == "" {
			csr.Subject.CommonName = name
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

func NewCertFromCSR(version, durMonth int, csr *x509.CertificateRequest) (*x509.Certificate, error) {
	i, err := newInt(2, defaultExp, defaultSub)
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		Version:         version,
		SerialNumber:    i,
		Subject:         csr.Subject,
		Extensions:      csr.Extensions,
		ExtraExtensions: csr.ExtraExtensions,
		DNSNames:        csr.DNSNames,
		EmailAddresses:  csr.EmailAddresses,
		IPAddresses:     csr.IPAddresses,
		URIs:            csr.URIs,
		NotBefore:       time.Now(),
		NotAfter:        time.Now().AddDate(0, durMonth, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageCodeSigning,
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageOCSPSigning,
			x509.ExtKeyUsageCodeSigning,
		},
		KeyUsage: x509.KeyUsageCertSign,
	}, nil
}
