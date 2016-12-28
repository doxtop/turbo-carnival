package counters

import (
  "fmt"
  "strings"
  "bytes"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "google.golang.org/appengine"
  "google.golang.org/appengine/datastore"
  "google.golang.org/appengine/taskqueue"
)

type Task struct {
  Name    string `json:"id"`
  Payload []byte `json:"value" datastore:",noindex"`
}

func (e *Task) Key(key *datastore.Key) {e.Name = key.Encode()}

func queue(w http.ResponseWriter, r *http.Request) {
  ctx   := appengine.NewContext(r)
  name  := r.Header.Get("X-AppEngine-TaskName")
  key   := datastore.NewKey(ctx,"Task",name,0,nil)

  var cs []Counter
  _ = json.NewDecoder(r.Body).Decode(&cs)
  defer r.Body.Close()

  for _,c := range cs {
    if _,err := c.Store(ctx);err!=nil{
      c.Set(err.Error())
      break
    }
    if err:= c.Collect(ctx);err!=nil{
      c.Set(err.Error())
    }
  }

  list,_ := json.Marshal(&cs)
  x := Task{key.Encode(), list}

  _,_ = datastore.Put(ctx,key,&x)
}

func enqueue(w http.ResponseWriter, r *http.Request) {
  ctx := appengine.NewContext(r)
  h   := make(http.Header)
      
  var bts []byte
  bts,err := ioutil.ReadAll(r.Body)

  if(err!=nil){
    http.Error(w, fmt.Sprintf("Error reading request:%v", err.Error()), 422)
    return
  }
  r.Body = ioutil.NopCloser(bytes.NewBuffer(bts))

  h.Set("Content-Type", "application/json")
  t := taskqueue.Task{
    Path: "/counters/queue",
    Payload: bts,
    Method: "POST",
    Header: h,
  }

  if task, err := taskqueue.Add(ctx, &t, "counters");err!=nil {
    http.Error(w, err.Error(), 500)
    return
  } else {
    k := datastore.NewKey(ctx,"Task",task.Name,0,nil)
    c := Task{k.Encode(), []byte("in_progress")}

    if _,err := datastore.Put(ctx,k,&c); err!=nil{
      http.Error(w,fmt.Sprintf("Queue can't be tracked: %s", err.Error) ,500)
    }
    w.Header().Set("Content-Type","application-json")
    json.NewEncoder(w).Encode(&c)
  }
}

func status(w http.ResponseWriter, r *http.Request){
  if(r.Method != "GET"){
    http.Error(w, "Method not supported", 405)
    return
  }
  ctx := appengine.NewContext(r)
 
  k,err := datastore.DecodeKey(strings.Split(strings.TrimSpace(r.URL.Path), "/")[2]); 

  if err!=nil{
    http.Error(w, fmt.Sprintf("Mailformed key: %v", err), 422)
    return
  }

  var c Task
  if err = datastore.Get(ctx, k, &c); err!=nil {
    http.Error(w, fmt.Sprintf("No such entry: %v", err), 404)
    return
  }
  
  w.Header().Set("Content-Type","application-json")
  c.Key(k)
  json.NewEncoder(w).Encode(&c)
}