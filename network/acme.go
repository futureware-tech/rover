package network

// TODO(dotdoom): consider using golang.org/x/crypto/acme/autocert when it supports DNS challenge.

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	dns "google.golang.org/api/dns/v1"

	"golang.org/x/crypto/acme"
	"golang.org/x/net/context"
)

const (
	keyType         = "EC PRIVATE KEY"
	accountFilename = "account.json"
	keyFilename     = "account.key"
)

// ACMEClient is incapsulating needed information to access ACME and Google Cloud DNS
type ACMEClient struct {
	DNS           *DNSClient
	WorkDirectory string
	client        *acme.Client
	account       *acme.Account
}

func readKey(filename string) (crypto.Signer, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	d, _ := pem.Decode(b)
	if d == nil {
		return nil, fmt.Errorf("Key block not found in %q", filename)
	}
	if d.Type == keyType {
		return x509.ParseECPrivateKey(d.Bytes)
	}
	return nil, fmt.Errorf("Key block type %q is not supported", d.Type)
}

func writeKey(filename string, k *ecdsa.PrivateKey) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	b := &pem.Block{Type: keyType}
	b.Bytes, err = x509.MarshalECPrivateKey(k)
	if err == nil {
		err = pem.Encode(f, b)
	}
	if closeErr := f.Close(); closeErr != nil {
		if err == nil {
			err = closeErr
		} else {
			log.Println(closeErr)
		}
	}
	return err
}

func readOrCreateKey(filename string) (crypto.Signer, error) {
	key, err := readKey(filename)
	if err != nil {
		log.Println(err)
		ecKey, e := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if e != nil {
			return nil, e
		}
		err, key = writeKey(filename, ecKey), ecKey
	}
	return key, err
}

func readAccount(filename string) (*acme.Account, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	account := &acme.Account{}
	return account, json.Unmarshal(b, account)
}

func writeAccount(filename string, account *acme.Account) error {
	b, err := json.MarshalIndent(account, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, b, 0600)
}

// NewACMEClient creates an ACME account, if necessary, and writes it's key and config to disk
func NewACMEClient(ctx context.Context, directory string) (*ACMEClient, error) {
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	key, err := readOrCreateKey(filepath.Join(directory, keyFilename))
	if err != nil {
		return nil, err
	}
	c := ACMEClient{
		client:        &acme.Client{Key: key},
		WorkDirectory: directory,
	}
	accountFullFilename := filepath.Join(directory, accountFilename)
	c.account, err = readAccount(accountFullFilename)
	if err != nil {
		c.account, err = c.client.Register(ctx, c.account, acme.AcceptTOS)
		if err != nil {
			return nil, err
		}
		err = writeAccount(accountFullFilename, c.account)
	}
	return &c, err
}

// GetDomainsCertpairPath returns paths to a certificate and key for domains, based on WorkDirectory
func (c *ACMEClient) GetDomainsCertpairPath(domains ...string) (string, string) {
	commonPrefix := filepath.Join(c.WorkDirectory, domains[0])
	return commonPrefix + ".crt", commonPrefix + ".key"
}

func (c *ACMEClient) requestAndWriteCertificate(ctx context.Context, domains []string) error {
	for _, domain := range domains {
		if err := c.authorizeDomain(ctx, domain); err != nil {
			return err
		}
	}

	certPath, keyPath := c.GetDomainsCertpairPath(domains...)
	key, err := readOrCreateKey(keyPath)
	if err != nil {
		return err
	}

	var csr []byte
	csr, err = x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: domains[0]},
		DNSNames: domains,
	}, key)
	if err != nil {
		return err
	}

	var (
		der  [][]byte
		cert []byte
	)
	der, _, err = c.client.CreateCert(ctx, csr, 90*24*time.Hour, true)
	if err != nil {
		return err
	}
	for _, b := range der {
		b = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: b})
		cert = append(cert, b...)
	}
	return ioutil.WriteFile(certPath, cert, 0644)
}

func (c *ACMEClient) authorizeDomain(ctx context.Context, domain string) error {
	auth, err := c.client.Authorize(ctx, domain)
	if err != nil {
		return err
	}
	if auth.Status == acme.StatusValid {
		return nil
	}
	for _, challenge := range auth.Challenges {
		if challenge.Type == "dns-01" {
			var txt string
			txt, err = c.client.DNS01ChallengeRecord(challenge.Token)
			if err != nil {
				return err
			}
			err = c.DNS.UpdateDNS(ctx, &dns.ResourceRecordSet{
				Name:    "_acme-challenge." + domain + ".",
				Type:    "TXT",
				Rrdatas: []string{txt},
				Ttl:     60,
			}, true)
			if err != nil {
				return err
			}
			if _, err = c.client.Accept(ctx, challenge); err != nil {
				return err
			}
			_, err = c.client.WaitAuthorization(ctx, auth.URI)
			return err
		}
	}
	return fmt.Errorf("No DNS challenge found for domain %s", domain)
}

// CheckOrRefreshCertificate checks expiration date of the certificate (if it exists),
// and renews it if necessary
func (c *ACMEClient) CheckOrRefreshCertificate(ctx context.Context, domains ...string) error {
	// TODO(dotdoom): check certificate expiration date
	// TODO(dotdoom): set up a timer to refresh the certs?
	certPath, _ := c.GetDomainsCertpairPath(domains...)
	if _, err := os.Stat(certPath); err != nil {
		log.Println(err)
		return c.requestAndWriteCertificate(ctx, domains)
	}
	return nil
}
