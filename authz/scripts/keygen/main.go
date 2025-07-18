package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"log/slog"
	"os"
	"strings"

	"github.com/zalgonoise/x/cli/v2"

	"github.com/zalgonoise/x/authz/internal/keygen"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	runner := cli.NewRunner("keygen",
		cli.WithExecutors(map[string]cli.Executor{
			"new":    cli.Executable(ExecNew),
			"sign":   cli.Executable(ExecSign),
			"verify": cli.Executable(ExecVerify),
		}),
	)

	cli.Run(runner, logger)
}

func ExecNew(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("new", flag.ExitOnError)

	output := fs.String("output", "", "the input directory to write the new public and private keys.")
	filename := fs.String("name", "testkey", "the prefix to the public and private keys' filenames")

	err := fs.Parse(args)
	if err != nil {
		return 1, err
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

	stat, err := os.Stat(*output)
	if err != nil {
		return 1, err
	}

	if !stat.IsDir() {
		return 1, errors.New("output path must be a directory")
	}

	if *filename == "" {
		slog.WarnContext(ctx, "setting default filename prefix to 'testkey'")

		*filename = "testkey"
	}

	if split := strings.Split(*filename, "."); len(split) > 1 {
		slog.WarnContext(ctx, "filename contains dots ('.')",
			slog.String("final_prefix", split[0]),
		)

		*filename = split[0]
	}

	privateKey, err := keygen.New()
	if err != nil {
		return 1, err
	}

	priv, err := keygen.EncodePrivate(privateKey)
	if err != nil {
		return 1, err
	}

	if err = os.WriteFile(*output+"/"+*filename+"_private.pem", priv, 0644); err != nil {
		return 1, err
	}

	publicKey := &privateKey.PublicKey

	pub, err := keygen.EncodePublic(publicKey)
	if err != nil {
		return 1, err
	}

	if err = os.WriteFile(*output+"/"+*filename+"_public.pem", pub, 0644); err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "wrote private and public keys successfully",
		slog.String("output_path", *output),
		slog.String("public_key", *output+"/"+*filename+"_public.pem"),
		slog.String("private_key", *output+"/"+*filename+"_private.pem"),
	)

	return 0, nil
}

func ExecSign(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("enc", flag.ExitOnError)

	key := fs.String("key", "", "the private key to use when signing the data")
	data := fs.String("data", "testkey", "the content to sign using the private key")
	isBase64 := fs.Bool("b64", false, "accepts data in base-64 standard encoding")

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	rawPrivateKey, err := os.ReadFile(*key)
	if err != nil {
		return 1, err
	}

	privateKey, err := keygen.DecodePrivate(rawPrivateKey)

	s := keygen.ECDSASigner{Priv: privateKey}

	buf := []byte(*data)
	if *isBase64 {
		buf, err = base64.StdEncoding.DecodeString(*data)
		if err != nil {
			return 1, err
		}
	}

	signature, hash, err := s.Sign(buf)
	if err != nil {
		return 1, err
	}

	sig := make([]byte, base64.StdEncoding.EncodedLen(len(signature)))
	base64.StdEncoding.Encode(sig, signature)

	h := make([]byte, base64.StdEncoding.EncodedLen(len(hash)))
	base64.StdEncoding.Encode(h, hash)

	logger.InfoContext(ctx, "signed input data successfully",
		slog.String("signature", string(sig)),
		slog.String("hash", string(h)),
	)

	return 0, nil
}

func ExecVerify(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("dec", flag.ExitOnError)

	key := fs.String("key", "", "the public key to use when verifying the signed data")
	hash := fs.String("hash", "", "the hash for the signed data")
	sig := fs.String("sig", "testkey", "the signature to verify using the public key")

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	publicKeyPEM, err := os.ReadFile(*key)
	if err != nil {
		panic(err)
	}

	publicKey, err := keygen.DecodePublic(publicKeyPEM)
	if err != nil {
		return 1, err
	}

	v := keygen.ECDSAVerifier{Pub: publicKey}

	h, err := hex.DecodeString(*hash) // expect base16 input text
	if err != nil {
		return 1, err
	}

	s, err := hex.DecodeString(*sig) // expect base16 input text
	if err != nil {
		return 1, err
	}

	if err := v.Verify(h, s); err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "verified input signature successfully")

	return 0, nil
}
