package counters

import (
  "fmt"
  "math/rand"
  "strconv"
  "golang.org/x/net/context"
  "google.golang.org/appengine/datastore"
  "google.golang.org/appengine/log"
)

const(
  countGroup = 25
)

type Entity struct {
  Name    string `json:"id" datastore:"name"`
  Payload string `json:"count" datastore:"count"`
}

func (e *Entity) Key(key *datastore.Key) {e.Name = key.Encode()}
func (e *Entity) Set(v string){e.Payload = v}
func (e *Entity) Point() {e.Payload = "0"}


/*
 * Store. check negative update effect
 *
 * 6 write ops yet
 */
func (e *Entity) Store(ctx context.Context, kind string) (key *datastore.Key, err error) {
  
  // something wrong with that key here
  // need to check how to create new shards 
  name:= fmt.Sprintf("%s-%d", e.Name, rand.Intn(countGroup))
  key = datastore.NewKey(ctx,kind,name,0,nil)

  e1 := Entity{"", "0"}

  if err := datastore.Get(ctx, key, &e1); err!=nil && err != datastore.ErrNoSuchEntity{
    return nil, err
  }

  counter,_ := strconv.ParseUint( e.Payload, 10, 64)
  inc,    _ := strconv.ParseUint(e1.Payload, 10, 64)
  counter+=inc
  
  e.Set(strconv.FormatUint(uint64(counter), 10))
    
  key,err = datastore.Put(ctx,key,e);
  log.Infof(ctx,"Actually in db: %v, counter:%v",key, counter)
  return
}

/*
 * Fold 
 */
func (e *Entity) Count(ctx context.Context, kind string) (err error) {
  q := datastore.NewQuery("Counter").Filter("name=", e.Name)
  t := q.Run(ctx)
  
  var total uint64 = 0
  for {
    e1 := Entity{"", "0"}
    _,err := t.Next(&e1)

    if err == datastore.Done { break }
    if err!=nil { return err }
    
    v,err := strconv.ParseUint(e1.Payload, 10, 64); 
    if err!=nil {
      break
    }
    total += v
  }
  e.Set(strconv.FormatUint(uint64(total), 10))
  return
}
