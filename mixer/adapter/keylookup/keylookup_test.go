package keylookup

import (
	"io/ioutil"
	"testing"

	"strings"

	adapter_integration "istio.io/istio/mixer/pkg/adapter/test"
)

func TestReport(t *testing.T) {
	tplCrBytes, err := ioutil.ReadFile("template/template.yaml")
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}

	adptCrBytes, err := ioutil.ReadFile("config/keylookup.yaml")
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}

	operatorCfgBytes, err := ioutil.ReadFile("sample_operator_cfg.yaml")
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	operatorCfg := string(operatorCfgBytes)

	adapter_integration.RunTest(
		t,
		nil,
		adapter_integration.Scenario{
			Setup: func() (ctx interface{}, err error) {
				pServer, err := NewKeylookup("")
				if err != nil {
					return nil, err
				}
				go func() {
					pServer.Run()
				}()
				return pServer, nil
			},
			Teardown: func(ctx interface{}) {
				s := ctx.(Server)
				s.Close()
			},
			ParallelCalls: []adapter_integration.Call{
				{
					CallKind: adapter_integration.CHECK,
					Attrs:    map[string]interface{}{"request.host": "testval2"},
				},
				{
					CallKind: adapter_integration.CHECK,
					Attrs:    map[string]interface{}{"request.host": "second3"},
				},
				{
					CallKind: adapter_integration.CHECK,
					Attrs:    map[string]interface{}{"request.host": "unknown"},
				},
			},
			GetConfig: func(ctx interface{}) ([]string, error) {
				s := ctx.(Server)
				return []string{
					string(tplCrBytes),
					string(adptCrBytes),
					strings.Replace(operatorCfg, "{ADDRESS}", s.Addr(), 1),
				}, nil
			},
			Want: `
			{
				"AdapterState": null,
				"Returns": [
				 {
					"Check": {
					 "Status": {},
					 "ValidDuration": 5000000000,
					 "ValidUseCount": 0,
					 "RouteDirective": null
					},
					"Quota": null,
					"Error": {}
				 },
				 {
					"Check": {
					 "Status": {},
					 "ValidDuration": 5000000000,
					 "ValidUseCount": 0,
					 "RouteDirective": null
					},
					"Quota": null,
					"Error": {}
				 },
				 {
					"Check": {
					 "Status": {},
					 "ValidDuration": 5000000000,
					 "ValidUseCount": 0,
					 "RouteDirective": null
					},
					"Quota": null,
					"Error": {}
				 }
				]
			 }`,
		},
	)
}
