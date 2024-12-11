package kms

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash/crc32"
	"io"

	"cloud.google.com/go/auth/credentials"
	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	_ "golang.org/x/crypto/x509roots/fallback"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	Name = "projects/dev-tonym-us/locations/us-west1/keyRings/test-software-signing/cryptoKeys/cloud-lite-signing/cryptoKeyVersions/1"
)

func SignAsymmetric(w io.Writer, name string, message io.Reader) error {
	ctx := context.Background()
	creds, err := credentials.DetectDefault(
		&credentials.DetectOptions{
			Scopes: []string{"https://www.googleapis.com/auth/cloudkms"},
		},
	)
	if err != nil {
		return err
	}
	client, err := kms.NewKeyManagementClient(ctx, option.WithAuthCredentials(creds))
	if err != nil {
		return fmt.Errorf("failed to create kms client: %w", err)
	}
	defer client.Close()
	digest := sha256.New()
	_, err = io.Copy(digest, message)
	if err != nil {
		return fmt.Errorf("failed to create digest: %w", err)
	}
	// Optional but recommended: Compute digest's CRC32C.
	crc32c := func(data []byte) uint32 {
		t := crc32.MakeTable(crc32.Castagnoli)
		return crc32.Checksum(data, t)

	}
	digestCRC32C := crc32c(digest.Sum(nil))
	req := &kmspb.AsymmetricSignRequest{
		Name: name,
		Digest: &kmspb.Digest{
			Digest: &kmspb.Digest_Sha256{
				Sha256: digest.Sum(nil),
			},
		},
		DigestCrc32C: wrapperspb.Int64(int64(digestCRC32C)),
	}
	result, err := client.AsymmetricSign(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to sign digest: %w", err)
	}
	if !result.VerifiedDigestCrc32C ||
		result.Name != req.Name ||
		int64(crc32c(result.Signature)) != result.SignatureCrc32C.Value {
		return fmt.Errorf("AsymmetricSign: request corrupted in-transit")
	}
	n, err := w.Write(result.Signature)
	if n != len(result.Signature) || err != nil {
		return err
	}
	return nil
}
