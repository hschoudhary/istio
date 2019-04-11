// Generates the keystoneAdapter adapter's resource yaml. It contains the adapter's configuration, name,
// supported template names (metric in this case), and whether it is session or no-session based.
//go:generate $GOPATH/src/istio.io/istio/bin/mixer_codegen.sh -a mixer/adapter/keystonegrpcadapter/config/config.proto -x "-s=false -n keystonegrpcadapter -t logentry"

package keystonegrpcadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/pborman/uuid"
	"google.golang.org/grpc"
	"istio.io/api/mixer/adapter/model/v1beta1"
	"istio.io/istio/mixer/adapter/keystonegrpcadapter/config"
	"istio.io/istio/mixer/template/logentry"
	"istio.io/istio/pkg/log"
)

// Log Entry Source type
const logEntrySourceType = "istio.logEntry"

// Server is basic server interface
type Server interface {
	Addr() string
	Close() error
	Run(shutdown chan error)
}

// Keystone adapter
type KeystoneGrpcAdapter struct {
	listener net.Listener
	server   *grpc.Server
}

var _ logentry.HandleLogEntryServiceServer = &KeystoneGrpcAdapter{}

// HandleMetric records metric entries
func (s *KeystoneGrpcAdapter) HandleLogEntry(ctx context.Context, r *logentry.HandleLogEntryRequest) (*v1beta1.ReportResult, error) {
	// TODO - Handle logic here
	log.Infof("Handling request %v", *r)
	cfg := &config.Params{}
	if r.AdapterConfig != nil {
		if err := cfg.Unmarshal(r.AdapterConfig.Value); err != nil {
			log.Errorf("Error unmarshalling adapter config:%v", err)
			return nil, err
		}
	}
	// Get log data
	log.Infof("Handler %d instances data", len(r.Instances))
	data, err := s.transformsInstanceData(r.Instances)
	if err != nil {
		log.Errorf("Could not transform any data, error=%v", err)
		return nil, err
	}
	if err := s.sendToKeystone(data, cfg.ApiKey); err != nil {
		log.Errorf("Failed to send data to keystone, error=%v", err)
		return nil, err
	}
	log.Infof("Successful sent data to keystone using config=%s", cfg.String())
	return &v1beta1.ReportResult{}, nil
}

func (s *KeystoneGrpcAdapter) transformsInstanceData(instances []*logentry.InstanceMsg) ([]byte, error) {
	if instances == nil {
		return nil, errors.New("Failed to transform instances data")
	}
	// TODO - convert data to json format
	var buffer bytes.Buffer
	instanceData := make(map[string]interface{})
	for _, instance := range instances {
		if len(instance.Variables) > 0 {
			instanceData["variables"] = instance.Variables
		}
		if len(instance.MonitoredResourceDimensions) > 0 {
			instanceData["monitorResourceDimensions"] = instance.MonitoredResourceDimensions
		}
		if instance.Severity != "" {
			instanceData["severity"] = instance.Severity
		}
		if instance.MonitoredResourceType != "" {
			instanceData["monitorResourceType"] = instance.MonitoredResourceType
		}

		// Add sourcetype and timestamp
		instanceData["_time"] = instance.Timestamp.GetValue().GetSeconds()
		instanceData["sourceType"] = logEntrySourceType
		data, err := json.Marshal(instanceData)
		if err != nil {
			log.Errorf("Failed to marshal instance data, instanceData=%v", instanceData)
			continue
		} else {
			buffer.Write(data)
		}
	}
	// Send bytes data
	return buffer.Bytes(), nil
}

func (s *KeystoneGrpcAdapter) sendToKeystone(data []byte, cfg *config.Params) error {
	if data == nil || len(data) == 0 || cfg.ApiKey == "" {
		return errors.New(fmt.Sprintf("invalid apikey=%s, or data=%s", cfg.ApiKey, data))
	}
	client := http.Client{
		Transport: http.DefaultTransport,
	}
	// TODO - Add config request time out here
	client.Timeout = time.Duration(10*time.Second)
	// TODO - Add actions and error details
	resp, err := client.Post(fmt.Sprintf("https://%s.api.keystone.splunkbeta.com/1.0/%s/%s/0/1", cfg.ApiKey, cfg.ApiKey, uuid.NewUUID().String()),
		"application/json", bytes.NewReader(data))
	if err != nil {
		log.Errora(err)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Infof("%s", body)
	return nil
}

// Addr returns the listening address of the server
func (s *KeystoneGrpcAdapter) Addr() string {
	return s.listener.Addr().String()
}

// Run starts the server run
func (s *KeystoneGrpcAdapter) Run(shutdown chan error) {
	shutdown <- s.server.Serve(s.listener)
}

// Close gracefully shuts down the server; used for testing
func (s *KeystoneGrpcAdapter) Close() error {
	if s.server != nil {
		s.server.GracefulStop()
	}

	if s.listener != nil {
		_ = s.listener.Close()
	}

	return nil
}

// NewKeystoneGrpcAdapter creates a new Keystone adapter that listens at provided port.
func NewKeystoneGrpcAdapter(addr string) (Server, error) {
	if addr == "" {
		addr = "0"
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", addr))
	if err != nil {
		return nil, fmt.Errorf("unable to listen on socket: %v", err)
	}
	s := &KeystoneGrpcAdapter{
		listener: listener,
	}
	fmt.Printf("listening on \"%v\"\n", s.Addr())
	s.server = grpc.NewServer()
	logentry.RegisterHandleLogEntryServiceServer(s.server, s)
	return s, nil
}
