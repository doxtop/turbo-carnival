package counters

import (
  "fmt"
  "testing"  
  "os"
  //"google.golang.org/appengine"
  "google.golang.org/appengine/aetest"
  "google.golang.org/appengine/datastore"
)

var(inst aetest.Instance)

func TestMain(m * testing.M){
  fmt.Println("Test Main")
  //var err error
  //inst,err := aetest.NewInstance(nil)
  //if(err!=nil){//  panic(err)//}
  c := m.Run()
  //inst.Close()
  os.Exit(c)
}

func TestStoreCounter(t *testing.T){
  ctx, done, err := aetest.NewContext()
  if err != nil { t.Fatal(err) }
  defer done()

  cases := [] struct {
    in,want Entity
  }{
    {Entity{"","0"}, Entity{"","0"}},
    {Entity{"","-1"}, Entity{"","1"}},
  }

  for _, c := range cases {
    got := c.in
    got.Store(ctx, "Counter")
    if (c.want.Payload != got.Payload) {
      t.Errorf("Store(%q) == %q, want %q", c.in, got, c.want)
    }
  }
}

func TestSomeRequest(t *testing.T){
  ctx, done, err := aetest.NewContext()
  if err != nil { t.Fatal(err) }
  defer done()
  fmt.Printf("4.280s to run: %v", datastore.NewKey(ctx, "Counter", "", 1, nil))
}
