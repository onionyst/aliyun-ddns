package main

import (
	"aliyun-ddns/internal/pkg/net"
	"fmt"
	"os"
	"time"

	alidns "github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

const (
	interval = 600 // seconds
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
	req := alidns.CreateDescribeSubDomainRecordsRequest()
	req.SubDomain = subDomain

	res, err := c.DescribeSubDomainRecords(req)
	if err != nil {
		return "", "", err
	}
	if res.TotalCount < 1 {
		return "", "", fmt.Errorf("no domain record for %s", subDomain)
	}

	record := res.DomainRecords.Record[0]

	return record.RecordId, record.Value, nil
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

		subDomain := fmt.Sprintf("%s.%s", rr, domain)
		recordID, domainIP, err := client.getSubDomainRecordIDAndIP(subDomain)
		if err != nil {
			panic(err)
		}

		ip, err := net.GetIP()
		if err != nil {
			panic(err)
		}

		if ip == domainIP {
			// already match
			fmt.Printf("Current domain IP already match: %s\n", ip)

		} else {
			err = client.updateDomainRecord(recordID, rr, ip)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Updated domain IP to: %s\n", ip)
		}

		time.Sleep(interval * time.Second)
	}
}
