package main

import (
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/zalgonoise/x/cli"

	"github.com/zalgonoise/x/authz/keygen"
)

var modes = []string{"new", "sign", "verify"}

func main() {
	runner := cli.NewRunner("keygen",
		cli.WithOneOf(modes...),
		cli.WithExecutors(map[string]cli.Executor{
			"new":    cli.Executable(ExecNew),
			"sign":   cli.Executable(ExecSign),
			"verify": cli.Executable(ExecVerify),
		}),
	)

	cli.Run(runner)
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

	if err := fs.Parse(args); err != nil {
		return 1, err
	}

	rawPrivateKey, err := os.ReadFile(*key)
	if err != nil {
		return 1, err
	}

	publicKey, err := keygen.DecodePrivate(rawPrivateKey)

	s := keygen.ECDSASigner{Priv: publicKey}

	plaintext := []byte(*data)
	signature, hash, err := s.Sign(plaintext)
	if err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "signed input data successfully",
		slog.String("signature", fmt.Sprintf("'%x'", signature)),
		slog.String("hash", fmt.Sprintf("'%x'", hash)),
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