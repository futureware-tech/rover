package network

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"reflect"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	dns "google.golang.org/api/dns/v1"
)

// DNSClient is a wrapper around Google Cloud DNS client.
type DNSClient struct {
	project string
	zone    string
	client  *dns.Service
}

// NewDNSClient connects new DNSClient and prepopulates some fields.
func NewDNSClient(ctx context.Context, zone string) (*DNSClient, error) {
	http, err := google.DefaultClient(ctx, dns.CloudPlatformScope)
	if err != nil {
		return nil, err
	}

	/* TODO(dotdoom): use this code once github.com/golang/oauth2/google f6093e3 is released
	var credentials *google.DefaultCredentials
	credentials, err = google.FindDefaultCredentials(ctx)
	*/
	var credentialsData []byte
	credentialsData, err = ioutil.ReadFile(".config/gcloud/application_default_credentials.json")
	if err != nil {
		return nil, err
	}
	var credentials struct {
		ProjectID string `json:"project_id"`
	}
	err = json.Unmarshal(credentialsData, &credentials)

	if err != nil {
		return nil, err
	}
	c := &DNSClient{
		project: credentials.ProjectID,
		zone:    zone,
	}
	c.client, err = dns.New(http)
	return c, err
}

func (c *DNSClient) pollCompletion(ctx context.Context, chg *dns.Change) error {
	startedAt, err := time.Parse(time.RFC3339Nano, chg.StartTime)
	if err != nil {
		log.Println("Failed to parse time:", err)
	}
	log.Printf("Change #%s started at %s with status: %s\n", chg.Id, startedAt.Local(), chg.Status)
	ticker := time.NewTicker(500 * time.Millisecond).C
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker:
			chg, err = c.client.Changes.Get(c.project, c.zone, chg.Id).Context(ctx).Do()
			if err != nil {
				return err
			}
			log.Printf("Change #%s: %s\n", chg.Id, chg.Status)
			if chg.Status == "done" {
				return nil
			}
		}
	}
}

// UpdateDNS adds or replaces DNS record in Google Cloud DNS according to the arguments specified.
func (c *DNSClient) UpdateDNS(ctx context.Context, rrs *dns.ResourceRecordSet,
	waitPropagation bool) error {
	chg := &dns.Change{
		Additions: []*dns.ResourceRecordSet{rrs},
	}
	maxTtl := rrs.Ttl
	err := c.client.ResourceRecordSets.List(c.project, c.zone).Pages(ctx,
		func(page *dns.ResourceRecordSetsListResponse) error {
			for _, v := range page.Rrsets {
				if v.Name == rrs.Name && v.Type == rrs.Type {
					if v.Ttl > maxTtl {
						maxTtl = v.Ttl
					}
					if len(chg.Additions) == 1 &&
						chg.Additions[0].Ttl == v.Ttl &&
						reflect.DeepEqual(chg.Additions[0].Rrdatas, v.Rrdatas) {
						log.Println("Keeping existing record:", v.Rrdatas)
						chg.Additions = chg.Additions[:0]
					} else {
						log.Println("Deleting existing record:", v.Rrdatas)
						chg.Deletions = append(chg.Deletions, v)
					}
				}
			}
			return nil
		})
	if err != nil {
		return err
	}

	if len(chg.Additions) == 0 && len(chg.Deletions) == 0 {
		return nil
	}

	if len(chg.Additions) == 1 {
		log.Println("Adding new record:", chg.Additions[0].Rrdatas)
	}

	chg, err = c.client.Changes.Create(c.project, c.zone, chg).Context(ctx).Do()
	if err != nil {
		return err
	}

	err = c.pollCompletion(ctx, chg)
	if err != nil {
		return err
	}

	if waitPropagation {
		time.Sleep(time.Duration(maxTtl) * time.Second)
	}

	return nil
}
