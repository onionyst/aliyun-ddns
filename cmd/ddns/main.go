package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	alidns "github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"

	"aliyun-ddns/internal/pkg/net"
)

const (
	interval      = 600 // seconds
	retry         = 3
	retryInterval = 5 // seconds
)

type Client struct {
	*alidns.Client
}

func (c *Client) init(regionID, accessKeyID, accessKeySecret string) error {
	client, err := alidns.NewClientWithAccessKey(regionID, accessKeyID, accessKeySecret)
	if err != nil {
		return err
	}

	c.Client = client
	return nil
}

func (c *Client) getSubDomainRecordIDAndIP(subDomain string) (string, string, error) {
	var err error
	for i := 0; i < retry; i++ {
		req := alidns.CreateDescribeSubDomainRecordsRequest()
		req.SubDomain = subDomain

		res, err := c.DescribeSubDomainRecords(req)
		if err == nil && res.TotalCount >= 1 {
			record := res.DomainRecords.Record[0]
			return record.RecordId, record.Value, nil
		}

		if err == nil {
			err = fmt.Errorf("no domain record for %s", subDomain)
		}
		time.Sleep(retryInterval * time.Second)
	}

	return "", "", err
}

func (c *Client) updateDomainRecord(recordID, rr, ip string) error {
	req := alidns.CreateUpdateDomainRecordRequest()
	req.RecordId = recordID
	req.RR = rr
	req.Type = "A"
	req.Value = ip

	_, err := c.UpdateDomainRecord(req)
	return err
}

func main() {
	accessKeyID := os.Getenv("ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")
	regionID := os.Getenv("REGION_ID")
	domain := os.Getenv("DOMAIN")
	rr := os.Getenv("RR")

	for {
		var client Client
		err := client.init(regionID, accessKeyID, accessKeySecret)
		if err != nil {
			panic(err)
		}

		ip, err := net.GetIP()
		if err != nil {
			panic(err)
		}

		rrs := strings.Split(rr, ",")
		for _, subRR := range rrs {
			subDomain := fmt.Sprintf("%s.%s", subRR, domain)
			recordID, domainIP, err := client.getSubDomainRecordIDAndIP(subDomain)
			if err != nil {
				panic(err)
			}

			if domainIP == ip {
				// already match
				fmt.Printf("Domain %s IP already match: %s\n", subDomain, ip)

			} else {
				err = client.updateDomainRecord(recordID, subRR, ip)
				if err != nil {
					panic(err)
				}

				fmt.Printf("Updated domain %s IP to: %s\n", subDomain, ip)
			}

		}

		time.Sleep(interval * time.Second)
	}
}
