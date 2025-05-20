package sapcontrol

import (
	"context"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"path"
	"time"

	"github.com/hooklift/gowsdl/soap"
)

type STATECOLOR string

const (
	STATECOLORSAPControlGRAY   STATECOLOR = "SAPControl-GRAY"
	STATECOLORSAPControlGREEN  STATECOLOR = "SAPControl-GREEN"
	STATECOLORSAPControlYELLOW STATECOLOR = "SAPControl-YELLOW"
	STATECOLORSAPControlRED    STATECOLOR = "SAPControl-RED"
)

type Start struct {
	XMLName  xml.Name `xml:"urn:SAPControl Start"`
	Runlevel string   `xml:"runlevel,omitempty" json:"runlevel,omitempty"`
}

type Stop struct {
	XMLName      xml.Name `xml:"urn:SAPControl Stop"`
	Softtimeout  int32    `xml:"softtimeout,omitempty" json:"softtimeout,omitempty"`
	IsSystemStop int32    `xml:"IsSystemStop,omitempty" json:"IsSystemStop,omitempty"`
}

type GetProcessList struct {
	XMLName xml.Name `xml:"urn:SAPControl GetProcessList"`
}

type GetProcessListResponse struct {
	XMLName   xml.Name     `xml:"urn:SAPControl GetProcessListResponse"`
	Processes []*OSProcess `xml:"process>item,omitempty" json:"process>item,omitempty"`
}

type OSProcess struct {
	Name        string      `xml:"name,omitempty" json:"name,omitempty"`
	Description string      `xml:"description,omitempty" json:"description,omitempty"`
	Dispstatus  *STATECOLOR `xml:"dispstatus,omitempty" json:"dispstatus,omitempty"`
	Textstatus  string      `xml:"textstatus,omitempty" json:"textstatus,omitempty"`
	Starttime   string      `xml:"starttime,omitempty" json:"starttime,omitempty"`
	Elapsedtime string      `xml:"elapsedtime,omitempty" json:"elapsedtime,omitempty"`
	Pid         int32       `xml:"pid,omitempty" json:"pid,omitempty"`
}

type StartResponse struct {
}

type StopResponse struct {
}

type SAPControlConnector interface {
	/* Triggers an instance start and returns immediately. */
	StartContext(ctx context.Context, request *Start) (*StartResponse, error)
	/* Triggers an instance stop and returns immediately. */
	StopContext(ctx context.Context, request *Stop) (*StopResponse, error)
	/* Returns a list of all processes directly started by the webservice according to the SAP start profile. */
	GetProcessListContext(ctx context.Context, request *GetProcessList) (*GetProcessListResponse, error)
}

type sapControlConnector struct {
	client *soap.Client
}

func NewSAPControlConnector(instNumber string) SAPControlConnector {
	socket := path.Join("/tmp", fmt.Sprintf(".sapstream5%s13", instNumber))

	udsClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "unix", socket)
			},
		},
	}

	// The url used here is just phony:
	// we need a well formed url to create the instance but the above DialContext function won't actually use it.
	client := soap.NewClient("http://unix", soap.WithHTTPClient(udsClient))

	return &sapControlConnector{
		client: client,
	}
}

func (service *sapControlConnector) StartContext(ctx context.Context, request *Start) (*StartResponse, error) {
	response := new(StartResponse)
	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *sapControlConnector) StopContext(ctx context.Context, request *Stop) (*StopResponse, error) {
	response := new(StopResponse)
	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *sapControlConnector) GetProcessListContext(ctx context.Context, request *GetProcessList) (*GetProcessListResponse, error) {
	response := new(GetProcessListResponse)
	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
