package counters

import (
  "fmt"
  "math/big"
  "crypto/rand"
  "strconv"
  "golang.org/x/net/context"
  "google.golang.org/appengine/datastore"
  "google.golang.org/appengine/log"
)

const(cshards = 25)

type Counter struct {
  Name    string `json:"id"`
  Count   string `json:"count" datastore:",noindex"`
}

//because we can
func Shard() (shard uint64, err error) {
  b := make([]byte,10)

  if _,err  = rand.Read(b);err!=nil { return }

  sh1,err := rand.Int(rand.Reader,big.NewInt(cshards)) 
  shard = sh1.Uint64()

  return
}

// Create a name for the counter
func (c *Counter) MkName() {
  b := make([]byte,16)

  if _,err := rand.Read(b);err!=nil { return }

  c.Name = fmt.Sprintf("%x%x%x%x%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
} 

// Unsafe increment string ints
func (c *Counter) Inc(i Counter) {
  counter,_ := strconv.ParseUint(c.Count, 10, 64)
  inc,    _ := strconv.ParseUint(i.Count, 10, 64)
  counter+=inc

  c.Count = strconv.FormatUint(uint64(counter), 10)
}

func (e *Counter) Set(v string){e.Count = v}

/*
 * Store.
 *
 * 4 write ops
 */
func (e *Counter) Store(ctx context.Context) (key *datastore.Key, err error) {
  if len(e.Name) == 0 { e.MkName() }

  e1      := Counter{Count:"0"}
  sh,err  := Shard()
  keyname := fmt.Sprintf("%s-%d", e.Name, sh)
  key      = datastore.NewKey(ctx, "Counter", keyname, 0, nil)
  
  if err  := datastore.Get(ctx, key, &e1);err!=nil && err!= datastore.ErrNoSuchEntity {return nil, err}

  e.Inc(e1)

  key,err  = datastore.Put(ctx,key,e);

  log.Infof(ctx, "Chech the name '%s' and store under key %v", keyname, key, )
  return
}

/*
 * Fold 
 */
func (e *Counter) Collect(ctx context.Context) (err error) {
  q := datastore.NewQuery("Counter").Filter("Name=", e.Name)
  t := q.Run(ctx)
  
  var total uint64 = 0
  for {
    e1 := Counter{Count:"0"}
    _,err := t.Next(&e1)

    if err == datastore.Done { break }
    if err!=nil { return err }
    
    v,err := strconv.ParseUint(e1.Count, 10, 64); 
    if err!=nil { break }

    total += v
  }
  e.Set(strconv.FormatUint(uint64(total), 10))
  return
}
