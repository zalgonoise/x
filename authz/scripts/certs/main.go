package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"flag"
	"io"
	"log/slog"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/zalgonoise/x/authz/certs"
	"github.com/zalgonoise/x/authz/keygen"
	"github.com/zalgonoise/x/cli/v2"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	runner := cli.NewRunner("keygen",
		cli.WithExecutors(map[string]cli.Executor{
			"new":    cli.Executable(ExecNew),
			"verify": cli.Executable(ExecVerify),
		}),
	)

	cli.Run(runner, logger)
}

func ExecNew(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("new", flag.ExitOnError)

	privPath := fs.String("private-key", "", "the path to the authz service's private key, used for signing the certificate.")
	pubPath := fs.String("public-key", "", "the path to the recipient service's public key, used for identifying this service in the certificate.")
	name := fs.String("name", "authz.service", "the Common Name (CN) to identify the recipient of the certificate [default: 'authz.service'].")
	filename := fs.String("filename", "cert.pem", "the filename for the generated certificate [default: 'cert.pem'].")
	output := fs.String("output", "", "the output directory to write the new certificates.")
	parent := fs.String("parent", "", "a parent certificate to serve as issuer of a certificate. Leaving this option blank generates a self-signed certificate.")
	durMonths := fs.Int("dur-months", 24, "how long in months should the generated certificate be valid.")

	err := fs.Parse(args)
	if err != nil {
		return 1, err
	}

	if *privPath == "" {
		return 1, errors.New("private key path must be set")
	}

	if *pubPath == "" {
		return 1, errors.New("public key path must be set")
	}

	if *output == "" {
		*output, err = os.MkdirTemp(os.TempDir(), "keys")
		if err != nil {
			return 1, err
		}

		slog.WarnContext(ctx, "setting default output directory to a temporary folder",
			slog.String("output_path", *output),
		)
	}

	if *filename == "" {
		slog.WarnContext(ctx, "setting default filename prefix to 'testkey'")

		*filename = "cert.pem"
	}

	logger.InfoContext(ctx, "generating new certificate file")
	certFile, err := createCertFile(*output, *filename)
	if err != nil {
		return 1, err
	}

	defer certFile.Close()

	logger.InfoContext(ctx, "opening public and private key files")
	pub, priv, err := openKeys(*pubPath, *privPath)
	if err != nil {
		return 1, err
	}

	var parentCert *x509.Certificate
	if *parent != "" {
		logger.InfoContext(ctx, "opening parent certificate")
		parentCert, err = openCertificate(*parent)
	}

	serial, err := newInt(2, 130, 1)
	if err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "generating serial", slog.Int64("serial", serial.Int64()))

	cn := pkix.Name{
		SerialNumber: serial.String(),
		CommonName:   *name,
	}

	issuerName := pkix.Name{
		SerialNumber: cn.SerialNumber,
		CommonName:   cn.CommonName,
	}

	if parentCert != nil {
		issuerName = parentCert.Subject
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      cn,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, *durMonths, 0),
		IsCA:         parentCert == nil,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageCodeSigning,
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageOCSPSigning,
			x509.ExtKeyUsageTimeStamping,
		},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		Issuer:                issuerName,
	}

	if parentCert == nil {
		parentCert = tmpl
	}

	logger.InfoContext(ctx, "encoding certificate")
	certificate, err := certs.Encode(tmpl, parentCert, pub, priv)
	if err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "writing certificate to file")
	if _, err = certFile.Write(certificate); err != nil {
		return 1, err
	}

	if err = certFile.Close(); err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "generated certificate successfully")
	return 0, nil
}

func ExecVerify(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("new", flag.ExitOnError)

	target := fs.String("target", "", "path to the certificate's PEM file that is target of verification.")
	root := fs.String("root", "", "path to the root certificate in the chain of trust.")
	inter := fs.String("intermediates", "", "comma-separated path for any intermediate certificates to be included in the chain of trust.")

	err := fs.Parse(args)
	if err != nil {
		return 1, err
	}

	if *target == "" {
		return 1, errors.New("target certificate cannot be empty")
	}

	if *root == "" {
		return 1, errors.New("root certificate cannot be empty")
	}

	var intermediates *x509.CertPool

	if *inter != "" {
		split := strings.Split(*inter, ",")

		logger.InfoContext(ctx, "parsing intermediate certificates",
			slog.Int("num_certificates", len(split)),
			slog.Any("certificates", split))

		intermediateCerts := make([]*x509.Certificate, 0, len(split))

		for i := range split {
			pemData, err := os.ReadFile(split[i])
			if err != nil {
				return 1, err
			}

			cert, err := certs.Decode(pemData)
			if err != nil {
				return 1, err
			}

			intermediateCerts = append(intermediateCerts, cert)
		}

		if len(intermediateCerts) > 0 {
			intermediates = x509.NewCertPool()

			for i := range intermediateCerts {
				intermediates.AddCert(intermediateCerts[i])
			}

			logger.InfoContext(ctx, "added certificates to the pool",
				slog.Int("num_certificates", len(intermediateCerts)))
		}
	}

	logger.InfoContext(ctx, "reading target PEM")
	targetPEM, err := os.ReadFile(*target)
	if err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "reading root PEM")
	rootPEM, err := os.ReadFile(*root)
	if err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "decoding root PEM")
	rootCert, err := certs.Decode(rootPEM)
	if err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "verifying target PEM against root PEM (and intermediates)")
	if err = certs.Verify(targetPEM, rootCert, intermediates); err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "verified target certificate successfully")

	return 0, nil
}

func createCertFile(path, name string) (io.WriteCloser, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		return nil, errors.New("output path must be a directory")
	}

	if path[len(path)-1] != '/' {
		path += "/"
	}

	return os.Create(path + name)
}

func openKeys(pub, priv string) (*ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	pubPEM, err := os.ReadFile(pub)
	if err != nil {
		return nil, nil, err
	}

	public, err := keygen.DecodePublic(pubPEM)
	if err != nil {
		return nil, nil, err
	}

	privPEM, err := os.ReadFile(priv)
	if err != nil {
		return nil, nil, err
	}

	private, err := keygen.DecodePrivate(privPEM)
	if err != nil {
		return nil, nil, err
	}

	return public, private, nil
}

func newInt(base, exp, sub int64) (*big.Int, error) {
	maximum := new(big.Int)
	maximum.Exp(big.NewInt(base), big.NewInt(exp), nil).Sub(maximum, big.NewInt(sub))

	return rand.Int(rand.Reader, maximum)
}

func openCertificate(path string) (*x509.Certificate, error) {
	certPEM, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return certs.Decode(certPEM)
}
