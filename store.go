package counters

import (
  "fmt"
  "math/rand"
  "strconv"
  "golang.org/x/net/context"
  "google.golang.org/appengine/datastore"
  //"google.golang.org/appengine/log"
)

const(
  countGroup = 25
)

func (e *Entity) Key(key *datastore.Key) {e.Name = key.Encode()}
func (e *Entity) Set(v string){e.Payload = v}
func (e *Entity) Point() {e.Payload = "0"}

/*
 * Store. check negative update effect
 */
func (e *Entity) Store(ctx context.Context, kind string) (key *datastore.Key, err error) {
  total,err := strconv.ParseUint(e.Payload, 10, 64)
  if(err!=nil){
    return
  }
  name:= fmt.Sprintf("%s-%d", e.Name, rand.Intn(countGroup))
  key = datastore.NewKey(ctx,kind,name,0,nil)
  
  err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
    e1 := Entity{"", "0"}

    if err := datastore.Get(ctx, key, &e1); err!=nil && err != datastore.ErrNoSuchEntity{
      return err
    }

    v,err := strconv.ParseUint(e1.Payload, 10, 64)
    if (err != nil){
      return err
    }
    total += v

    e.Set(strconv.FormatUint(uint64(total), 10))
    
    _,err = datastore.Put(ctx,key,e);

    return err

  }, nil)
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
