// nolint:lll
// Generates the keylookup adapter's resource yaml. It contains the adapter's configuration, name,
// supported template names (metric in this case), and whether it is session or no-session based.
//go:generate $REPO_ROOT/bin/mixer_codegen.sh -a mixer/adapter/keylookup/config/config.proto -x "-s=false -n keylookup -t keylookup"
//go:generate $REPO_ROOT/bin/mixer_codegen.sh -t mixer/adapter/keylookup/template/template.proto -x "-n keylookup"
package keylookup

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"istio.io/api/mixer/adapter/model/v1beta1"
	"istio.io/istio/mixer/adapter/keylookup/config"
	ktpl "istio.io/istio/mixer/adapter/keylookup/template"
)

type (
	// Server is basic server interface
	Server interface {
		Addr() string
		Close() error
		Run()
		Wait() error
	}

	// Keylookup supports keylookup template
	Keylookup struct {
		shutdown chan error
		listener net.Listener
		server   *grpc.Server
	}
)

var _ ktpl.HandleKeylookupServiceServer = &Keylookup{}

// HandleKeylookup looks up the key
func (s *Keylookup) HandleKeylookup(_ context.Context, req *ktpl.HandleKeylookupRequest) (*ktpl.HandleKeylookupResponse, error) {
	config := &config.Params{}

	if err := config.Unmarshal(req.AdapterConfig.Value); err != nil {
		return nil, err
	}

	foundKey := ""

	for _, entry := range config.Map {
		for _, candidate := range entry.Values {
			if candidate == req.Instance.Entry {
				foundKey = entry.Key
			}
		}
	}

	if foundKey == "" {
		fmt.Printf("No key found for %s\n", req.Instance.Entry)
	}

	return &ktpl.HandleKeylookupResponse{
		Result: &v1beta1.CheckResult{
			ValidDuration: config.ValidDuration,
		},
		Output: &ktpl.OutputMsg{Value: foundKey},
	}, nil
}

// Addr returns the listening address of the server
func (s *Keylookup) Addr() string {
	return s.listener.Addr().String()
}

// Close gracefully shuts down the server; used for testing
func (s *Keylookup) Close() error {
	if s.shutdown != nil {
		s.server.GracefulStop()
		_ = s.Wait()
	}

	if s.listener != nil {
		_ = s.listener.Close()
	}

	return nil
}

// Wait waits for server to stop
func (s *Keylookup) Wait() error {
	if s.shutdown == nil {
		return fmt.Errorf("server not running")
	}

	err := <-s.shutdown
	s.shutdown = nil
	return err
}

// Run starts the server run
func (s *Keylookup) Run() {
	s.shutdown = make(chan error, 1)
	go func() {
		err := s.server.Serve(s.listener)

		// notify closer we're done
		s.shutdown <- err
	}()
}

// NewKeylookup creates a new mixer adapter that listens at provided port.
func NewKeylookup(addr string) (Server, error) {
	if addr == "" {
		addr = "0"
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", addr))
	if err != nil {
		return nil, fmt.Errorf("unable to listen on socket: %v", err)
	}
	s := &Keylookup{
		listener: listener,
	}
	fmt.Printf("listening on \"%v\"\n", s.Addr())
	s.server = grpc.NewServer()
	ktpl.RegisterHandleKeylookupServiceServer(s.server, s)
	return s, nil
}
