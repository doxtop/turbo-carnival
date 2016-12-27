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

func (e *Entity) Inc(e1 *Entity) (err error) { 
  i1,err := strconv.ParseUint( e.Payload, 10, 64)
  i2,err := strconv.ParseUint(e1.Payload, 10, 64)
  e.Payload = fmt.Sprintf("%d", i1+i2)
  return
}

func Store(ctx context.Context, kind string, e *Entity) (key *datastore.Key, err error) {
  log.Infof(ctx, "Store %v %v", kind, e)

  name:= fmt.Sprintf("%d",rand.Intn(countGroup))
  key = datastore.NewKey(ctx,kind,name,0,nil)

  err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
    e1:=new(Entity) 

    if err := datastore.Get(ctx, key, e1); err!=nil && err != datastore.ErrNoSuchEntity{
      return err
    }
    
    if err := e.Inc(e1); err!=nil { return err }

    _,err = datastore.Put(ctx,key,e);

    return err

  }, nil/* once */)
  return
}

