package counters

import (
  "fmt"
  "testing"  
  "google.golang.org/appengine"
  "google.golang.org/appengine/aetest"
  "google.golang.org/appengine/datastore"
)

func SlowAsHellTest(t *testing.T){
  inst, err := aetest.NewInstance(nil)
  if err != nil {
    t.Fatalf("Failed to create instance: %v", err)
  }
  defer inst.Close()

  req, err := inst.NewRequest("GET", "/counters", nil)
  if err != nil {
    t.Fatalf("Failed to create req: %v", err)
  }
  ctx := appengine.NewContext(req)

  fmt.Printf("4.280s to run: %v", datastore.NewKey(ctx, "Counter", "", 1, nil))
}

