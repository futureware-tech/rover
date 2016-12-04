package network

import (
	"log"
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
	c := &DNSClient{
		// TODO(dotdoom): read from the config file.
		project: "rover-cloud",
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
	err := c.client.ResourceRecordSets.List(c.project, c.zone).Pages(ctx,
		func(page *dns.ResourceRecordSetsListResponse) error {
			for _, v := range page.Rrsets {
				if v.Name == rrs.Name && v.Type == rrs.Type {
					// TODO(dotdoom): early return if the record is unchanged.
					log.Println("Delete:", v.Rrdatas)
					chg.Deletions = append(chg.Deletions, v)
				}
			}
			return nil
		})
	if err != nil {
		return err
	}

	log.Println("Add:", rrs.Rrdatas)
	chg, err = c.client.Changes.Create(c.project, c.zone, chg).Context(ctx).Do()
	if err != nil {
		return err
	}

	if waitPropagation {
		// TODO(dotdoom): check propagation via DNS
		// TODO(dotdoom): use old record's TTL instead
		time.Sleep(time.Duration(rrs.Ttl) * time.Second)
	}

	return c.pollCompletion(ctx, chg)
}
