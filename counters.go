package counters

import (
  "fmt"
  "strings"
  "html/template"
  "net/http"
  "io/ioutil"
  "bytes"
  "encoding/json"
  "google.golang.org/appengine"
  "google.golang.org/appengine/datastore"
  "google.golang.org/appengine/taskqueue"
  "google.golang.org/appengine/log"
  //"google.golang.org/appengine/memcache" - no memory caching for the moment
)

/* 
 * Carries the entities (Counters,Tasks) from client to datastore. 
 * Values stored as string, service should check the type with ParseUint()
 * Id - duplicated key.
 */
type Counter struct{
  Id string     `json:"id"`
  Value string  `json:"value"`
}

func (c *Counter) Point(id *datastore.Key) {c.Id = id.Encode()}

func create(w http.ResponseWriter, r *http.Request){
  if(r.Method != "POST"){
    http.Error(w, "Method not supported", 405)
    return
  }
  ctx := appengine.NewContext(r)
  k := datastore.NewKey(ctx,"Counter","",0,nil)
  c := Counter{k.Encode(),"0"}

  if rk,err := datastore.Put(ctx,k,&c); err!=nil {
    http.Error(w, fmt.Sprintf("Can't process entity: %v",err), 422)
    return
  } else {
    c.Point(rk)
  }

  w.Header().Set("Content-Type", "application-json")
  json.NewEncoder(w).Encode(&c)
}

// get,update counter value
func counter(w http.ResponseWriter, r *http.Request){
  ctx := appengine.NewContext(r)

  switch r.Method {
    case "GET": 
      k,err := datastore.DecodeKey(strings.Split(strings.TrimSpace(r.URL.Path), "/")[2]); 

      if err!=nil{
        http.Error(w, fmt.Sprintf("Mailformed key: %v", err), 422)
        return
      }
      var c = Counter{"","0"}
      if err = datastore.Get(ctx, k, &c); err!=nil {
        http.Error(w, fmt.Sprintf("No such entry: %v", err), 404)
        return
      }
      c.Point(k)
  
      w.Header().Set("Content-Type","application-json")
      json.NewEncoder(w).Encode(&c)

    case "PUT","POST":
      k,err := datastore.DecodeKey(strings.Split(strings.TrimSpace(r.URL.Path), "/")[2]);

      if err!=nil{
        http.Error(w, fmt.Sprintf("Mailformed key: %v", err), 422)
        return
      }

      var c Counter
      if err := json.NewDecoder(r.Body).Decode(&c); err!=nil{
        log.Infof(ctx, "Fail %v", err)
        http.Error(w, fmt.Sprintf("Mailformed counter: %v", err), 422)
      }
      defer r.Body.Close()

      c.Point(k)

      if rk,err := datastore.Put(ctx,k,&c); err!=nil {
        http.Error(w, fmt.Sprintf("Can't process entity: %s",err.Error()), 422)
        return
      } else {
        c.Point(rk)
      }

      w.Header().Set("Content-Type","application-json")
      json.NewEncoder(w).Encode(&c)

    default:
      http.Error(w, fmt.Sprintf("Method not supported %v", r.Method), 405)
  }
}

// list all
func list(w http.ResponseWriter, r *http.Request){
  ctx := appengine.NewContext(r)
  log.Infof(ctx, "Request to fucking list %v ", r.Method)

  switch r.Method {
    case "GET": 
      q :=datastore.NewQuery("Counter").Limit(10)
      cs := make([]Counter, 0, 10)

      //_,err := q.GetAll(ctx, &cs)
      t := q.Run(ctx)
      for{
        var x Counter
        k,err := t.Next(&x)
        
        if(err == datastore.Done){ break }
        if(err!=nil){
          x.Point(k)
          x.Value = err.Error()
          cs = append(cs, x)
          break
        }
        x.Point(k)
        cs = append(cs,x)
      }
      
      w.Header().Set("Content-Type","application-json")
      json.NewEncoder(w).Encode(&cs)
    case "POST", "PUT":
      // queue
      h := make(http.Header)
      
      var bts []byte
      //defer r.Body.Close()
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
        log.Infof(ctx, "Task added: %v", task.Name)

        k := datastore.NewKey(ctx,"Task",task.Name,0,nil)

        log.Infof(ctx,"taskKey %v",k)
        
        c := Counter{k.Encode(), "in-progress"}

        if _,err := datastore.Put(ctx,k,&c); err!=nil{
          log.Infof(ctx, "Can't save task %v", err.Error())
        }

        w.Header().Set("Content-Type","application-json")
        json.NewEncoder(w).Encode(&c)
      }

    default:
      http.Error(w, fmt.Sprintf("Method not supported %v", r.Method), 405)
  }
}

// handle queue task
func queue(w http.ResponseWriter, r *http.Request) {
  ctx := appengine.NewContext(r)
  name := r.Header.Get("X-AppEngine-TaskName")

  log.Infof(ctx, "this is queue handler for %v", name)

  var cs []Counter
  if err := json.NewDecoder(r.Body).Decode(&cs); err!=nil {
    log.Infof(ctx, "Decode failed: %s", err.Error())
    // put the err into db
    return
  }
  defer r.Body.Close()

  for i,c := range cs {
    if k,err := datastore.DecodeKey(c.Id); err!=nil{
      cs[i].Value = fmt.Sprintf("Invalid key: %s", err.Error())
    } else {
      if _,err := datastore.Put(ctx,k,&c); err!=nil {
        log.Infof(ctx, "some shit happened %v", err)
        cs[i].Value = fmt.Sprintf("", err.Error())
      }
    }
  }

  log.Infof(ctx, "Will encode this: %v", &cs)

  k := datastore.NewKey(ctx,"Task",name,0,nil)
  x := Counter{k.Encode(), "done"}

  if _,err := datastore.Put(ctx,k,&x); err!=nil{
    log.Infof(ctx, "Tak update error", err.Error())
  }

  w.Header().Set("Content-Type", "application-json")
  json.NewEncoder(w).Encode(&cs)

}

func task(w http.ResponseWriter, r *http.Request){
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

  var c Counter
  if err = datastore.Get(ctx, k, &c); err!=nil {
    http.Error(w, fmt.Sprintf("No such entry: %v", err), 404)
    return
  }
  c.Point(k)
  
  w.Header().Set("Content-Type","application-json")
  json.NewEncoder(w).Encode(&c)
}

func persist(w http.ResponseWriter, r *http.Request){
  // take the memory
  ctx := appengine.NewContext(r)
  log.Infof(ctx, "periodic persist")
  // store to disk
}

// entry point
func init() {
  http.HandleFunc("/counters/persist", persist)
  http.HandleFunc("/tasks/",          task)
  http.HandleFunc("/counters/queue",  queue)
  http.HandleFunc("/counters",        list)
  http.HandleFunc("/counter/",        counter)
  http.HandleFunc("/counter",         create)
  http.HandleFunc("/",                index)
}

func index(w http.ResponseWriter, r *http.Request){
  if err := template.Must(template.New("conters").Parse(`
    <!doctype html>
    <html>
      <head><title>counters</title></head>
      <body><p>counters</p></body>
    </html>`)).Execute(w, map[string]interface{}{}); err!=nil{
    http.Error(w, err.Error(), 500)
  }
}
