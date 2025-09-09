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

type StartStopOption string

const (
	StartStopOptionSAPControlALLINSTANCES      StartStopOption = "SAPControl-ALL-INSTANCES"
	StartStopOptionSAPControlSCSINSTANCES      StartStopOption = "SAPControl-SCS-INSTANCES"
	StartStopOptionSAPControlDIALOGINSTANCES   StartStopOption = "SAPControl-DIALOG-INSTANCES"
	StartStopOptionSAPControlABAPINSTANCES     StartStopOption = "SAPControl-ABAP-INSTANCES"
	StartStopOptionSAPControlJ2EEINSTANCES     StartStopOption = "SAPControl-J2EE-INSTANCES"
	StartStopOptionSAPControlPRIORITYLEVEL     StartStopOption = "SAPControl-PRIORITY-LEVEL"
	StartStopOptionSAPControlTREXINSTANCES     StartStopOption = "SAPControl-TREX-INSTANCES"
	StartStopOptionSAPControlENQREPINSTANCES   StartStopOption = "SAPControl-ENQREP-INSTANCES"
	StartStopOptionSAPControlHDBINSTANCES      StartStopOption = "SAPControl-HDB-INSTANCES"
	StartStopOptionSAPControlALLNOHDBINSTANCES StartStopOption = "SAPControl-ALLNOHDB-INSTANCES"
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

type StartSystem struct {
	XMLName       xml.Name         `xml:"urn:SAPControl StartSystem"`
	Options       *StartStopOption `xml:"options,omitempty" json:"options,omitempty"`
	Prioritylevel string           `xml:"prioritylevel,omitempty" json:"prioritylevel,omitempty"`
	Waittimeout   int32            `xml:"waittimeout,omitempty" json:"waittimeout,omitempty"`
	Runlevel      string           `xml:"runlevel,omitempty" json:"runlevel,omitempty"`
}

type StopSystem struct {
	XMLName       xml.Name         `xml:"urn:SAPControl StopSystem"`
	Options       *StartStopOption `xml:"options,omitempty" json:"options,omitempty"`
	Prioritylevel string           `xml:"prioritylevel,omitempty" json:"prioritylevel,omitempty"`
	Softtimeout   int32            `xml:"softtimeout,omitempty" json:"softtimeout,omitempty"`
	Waittimeout   int32            `xml:"waittimeout,omitempty" json:"waittimeout,omitempty"`
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

type GetSystemInstanceList struct {
	XMLName xml.Name `xml:"urn:SAPControl GetSystemInstanceList"`
	Timeout int32    `xml:"timeout,omitempty" json:"timeout,omitempty"`
}

type GetSystemInstanceListResponse struct {
	XMLName   xml.Name       `xml:"urn:SAPControl GetSystemInstanceListResponse"`
	Instances []*SAPInstance `xml:"instance>item,omitempty" json:"instance>item,omitempty"`
}

type SAPInstance struct {
	Hostname      string      `xml:"hostname,omitempty" json:"hostname,omitempty"`
	InstanceNr    int32       `xml:"instanceNr,omitempty" json:"instanceNr,omitempty"`
	HttpPort      int32       `xml:"httpPort,omitempty" json:"httpPort,omitempty"`
	HttpsPort     int32       `xml:"httpsPort,omitempty" json:"httpsPort,omitempty"`
	StartPriority string      `xml:"startPriority,omitempty" json:"startPriority,omitempty"`
	Features      string      `xml:"features,omitempty" json:"features,omitempty"`
	Dispstatus    *STATECOLOR `xml:"dispstatus,omitempty" json:"dispstatus,omitempty"`
}

type StartResponse struct {
}

type StopResponse struct {
}

type StartSystemResponse struct {
}

type StopSystemResponse struct {
}

type SAPControlConnector interface {
	/* Triggers an instance start and returns immediately. */
	StartContext(ctx context.Context, request *Start) (*StartResponse, error)
	/* Triggers an instance stop and returns immediately. */
	StopContext(ctx context.Context, request *Stop) (*StopResponse, error)
	/* Triggers start of entire system or parts of it. */
	StartSystemContext(ctx context.Context, request *StartSystem) (*StartSystemResponse, error)
	/* Triggers stop or soft shutdown of entire system or parts of it. */
	StopSystemContext(ctx context.Context, request *StopSystem) (*StopSystemResponse, error)
	/* Returns a list of all processes directly started by the webservice according to the SAP start profile. */
	GetProcessListContext(ctx context.Context, request *GetProcessList) (*GetProcessListResponse, error)
	/* Returns a list of SAP instances of the SAP system. */
	GetSystemInstanceListContext(
		ctx context.Context,
		request *GetSystemInstanceList,
	) (*GetSystemInstanceListResponse, error)
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

func (service *sapControlConnector) StartSystemContext(
	ctx context.Context,
	request *StartSystem,
) (*StartSystemResponse, error) {
	response := new(StartSystemResponse)
	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *sapControlConnector) StopSystemContext(
	ctx context.Context,
	request *StopSystem,
) (*StopSystemResponse, error) {
	response := new(StopSystemResponse)
	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *sapControlConnector) GetProcessListContext(
	ctx context.Context,
	request *GetProcessList,
) (*GetProcessListResponse, error) {
	response := new(GetProcessListResponse)
	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *sapControlConnector) GetSystemInstanceListContext(
	ctx context.Context,
	request *GetSystemInstanceList,
) (*GetSystemInstanceListResponse, error) {
	response := new(GetSystemInstanceListResponse)
	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
