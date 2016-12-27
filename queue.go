package counters

import (
  "net/http"
  "encoding/json"
  "google.golang.org/appengine"
  "google.golang.org/appengine/datastore"
)

func queue(w http.ResponseWriter, r *http.Request) {
  ctx   := appengine.NewContext(r)
  name  := r.Header.Get("X-AppEngine-TaskName")
  key   := datastore.NewKey(ctx,"Task",name,0,nil)

  var cs []Entity
  _ = json.NewDecoder(r.Body).Decode(&cs)
  defer r.Body.Close()

  for _,c := range cs {
    if _,err := c.Store(ctx,"Counter");err!=nil{
      c.Set(err.Error())
      break
    }
    if err:= c.Count(ctx, "Counter");err!=nil{
      c.Set(err.Error())
    }
  }

  list,_ := json.Marshal(&cs)
  x := Entity{key.Encode(),string(list)}

  _,_ = datastore.Put(ctx,key,&x)
}

