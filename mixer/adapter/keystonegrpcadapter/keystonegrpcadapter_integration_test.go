package keystonegrpcadapter


import (
    "io/ioutil"
    "strings"
    "testing"
    "time"

    adapter_integration "istio.io/istio/mixer/pkg/adapter/test"
)

func TestReport(t *testing.T) {
  adptCrBytes, err := ioutil.ReadFile("config/keystonegrpcadapter.yaml")
  if err != nil {
     t.Fatalf("could not read file: %v", err)
  }

  operatorCfgBytes, err := ioutil.ReadFile("sample_operator_cfg.yaml")
  if err != nil {
     t.Fatalf("could not read file: %v", err)
  }
  operatorCfg := string(operatorCfgBytes)
  shutdown := make(chan error, 1)

  adapter_integration.RunTest(
     t,
     nil,
     adapter_integration.Scenario{
        Setup: func() (ctx interface{}, err error) {
           pServer, err := NewKeystoneGrpcAdapter("")
           if err != nil {
              return nil, err
           }
           go func() {
              pServer.Run(shutdown)
              _ = <-shutdown
           }()
           return pServer, nil
        },
        Teardown: func(ctx interface{}) {
           s := ctx.(Server)
           s.Close()
        },
        ParallelCalls: []adapter_integration.Call{
           {
              CallKind: adapter_integration.REPORT,
              Attrs:    map[string]interface{}{"request.size": int64(555), "request.time": time.Now()},
           },
        },
        GetState: func(ctx interface{}) (interface{}, error) {
           return nil, nil
        },
        GetConfig: func(ctx interface{}) ([]string, error) {
           s := ctx.(Server)
           return []string{
              // CRs for built-in templates (metric is what we need for this test)
              // are automatically added by the integration test framework.
              string(adptCrBytes),
              strings.Replace(operatorCfg, ":9090", s.Addr(), 1),
           }, nil
        },
        Want: `
     {
      "AdapterState": null,
      "Returns": [
       {
        "Check": {
         "Status": {},
         "ValidDuration": 0,
         "ValidUseCount": 0
        },
        "Quota": null,
        "Error": null
       }
      ]
     }`,
     },
  )
}

func normalize(s string) string {
  s = strings.TrimSpace(s)
  s = strings.Replace(s, "\t", "", -1)
  s = strings.Replace(s, "\n", "", -1)
  s = strings.Replace(s, " ", "", -1)
  return s
}
