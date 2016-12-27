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

func create(w http.ResponseWriter, r *http.Request){
  if(r.Method != "POST"){
    http.Error(w, "Method not supported", 405)
    return
  }
  ctx   := appengine.NewContext(r)
  key   := datastore.NewKey(ctx,"Counter","",0,nil)
  c     := Entity{key.Encode(),"0"}

  if _,err:= c.Store(ctx, "Counter");err!=nil {
    http.Error(w, fmt.Sprintf("Can't process entity: %v",err), 422)
    return
  }

  w.Header().Set("Content-Type", "application-json")
  json.NewEncoder(w).Encode(&c)
}

// get,update counter value
func counter(w http.ResponseWriter, r *http.Request){
  ctx   := appengine.NewContext(r)
  k,err := datastore.DecodeKey(strings.Split(strings.TrimSpace(r.URL.Path), "/")[2]); 

  if err!=nil{
    http.Error(w, fmt.Sprintf("Mailformed key: %v", err), 422)
    return
  }
  c := Entity{k.Encode(),"0"}

  switch r.Method {
    case "GET": 
      if err := c.Count(ctx, "Counter");err!=nil {
        http.Error(w, fmt.Sprintf("No entry: %v", err), 404)
        return
      }
    case "PUT","POST":
      if err := json.NewDecoder(r.Body).Decode(&c); err!=nil{
        http.Error(w, fmt.Sprintf("Mailformed counter: %v", err), 422)
      }
      defer r.Body.Close()
      
      if _,err := c.Store(ctx,"Counter");err!=nil{
        http.Error(w, fmt.Sprintf("Can't process entity: %s",err.Error()), 422)
        return
      }
      
      if err:= c.Count(ctx, "Counter");err!=nil{
        http.Error(w, fmt.Sprintf("Broken link: %s",err.Error()), 422)
        return
      }
    default:
      http.Error(w, fmt.Sprintf("Method not supported %v", r.Method), 405)
      return
  }

  w.Header().Set("Content-Type","application-json")
  json.NewEncoder(w).Encode(&c)  
}

// list all
func list(w http.ResponseWriter, r *http.Request){
  ctx := appengine.NewContext(r)

  switch r.Method {
    case "GET": 
      q   := datastore.NewQuery("Counter").KeysOnly()
      set := make(map[string]string)
      cs  := make([]interface{}, 0)
      t   := q.Run(ctx)

      for {
        var x datastore.Key
        k,err := t.Next(&x)
        
        if(err == datastore.Done){ break }
        if(err!=nil){ break }
        
        name :=  strings.Split(k.StringID(), "-")[1]
        _,ok := set[name]

        if !ok {
          c := Entity{name,"0"}
          if err := c.Count(ctx, "Counter");err!=nil{
            set[name] = err.Error()
          } else {
            set[name] = c.Payload
          }
        }
      }

      // map shoud somehow be serialized without this slice
      for k := range set {
        cs = append(cs, Entity{k,set[k]})
      }
      
      w.Header().Set("Content-Type","application-json")
      json.NewEncoder(w).Encode(&cs)
    case "POST", "PUT":
      h := make(http.Header)
      
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
        log.Infof(ctx, "Task added: %v", task.Name)

        k := datastore.NewKey(ctx,"Task",task.Name,0,nil)

        log.Infof(ctx,"taskKey %v",k.Encode())
        
        c := Entity{k.Encode(), "in_progress"}

        if _,err := datastore.Put(ctx,k,&c); err!=nil{
          log.Infof(ctx, "Can't save task %v", err.Error())
          http.Error(w,fmt.Sprintf("Queue can't be tracked: %s", err.Error) ,500)
        }
        w.Header().Set("Content-Type","application-json")
        json.NewEncoder(w).Encode(&c)
      }

    default:
      http.Error(w, fmt.Sprintf("Method not supported %v", r.Method), 405)
  }
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

  var c Entity
  if err = datastore.Get(ctx, k, &c); err!=nil {
    http.Error(w, fmt.Sprintf("No such entry: %v", err), 404)
    return
  }
  
  w.Header().Set("Content-Type","application-json")
  c.Key(k)
  json.NewEncoder(w).Encode(&c)
}

func persist(w http.ResponseWriter, r *http.Request){
  ctx := appengine.NewContext(r)
  log.Infof(ctx, "periodic persist")
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
