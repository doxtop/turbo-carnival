package counters

import (
  "fmt"
  "strings"
  "net/http"
  "encoding/json"
  "google.golang.org/appengine"
  "google.golang.org/appengine/datastore"
  "google.golang.org/appengine/log"
)

func create(w http.ResponseWriter, r *http.Request){
  if(r.Method != "POST"){
    http.Error(w, "Method not supported", 405)
    return
  }
  ctx := appengine.NewContext(r)
  c   := new(Counter)
  
  if _,err:= c.Store(ctx);err!=nil {
    http.Error(w, fmt.Sprintf("Can't process entity: %v",err), 422)
    return
  }

  log.Infof(ctx, "%v created", c)

  w.Header().Set("Content-Type", "application-json")
  json.NewEncoder(w).Encode(&c)
}

// get,update counter value
func counter(w http.ResponseWriter, r *http.Request){
  ctx := appengine.NewContext(r)
  name:= strings.Split(strings.TrimSpace(r.URL.Path), "/")[2]
  c   := Counter{name,"0"}

  switch r.Method {
    case "GET": 
      if err := c.Collect(ctx);err!=nil {
        http.Error(w, fmt.Sprintf("No entry: %v", err), 404)
        return
      }
    case "PUT","POST":

      if err := json.NewDecoder(r.Body).Decode(&c); err!=nil{
        http.Error(w, fmt.Sprintf("Mailformed counter: %v", err), 422)
      }
      defer r.Body.Close()
      inc := c

      if _,err := c.Store(ctx);err!=nil{
        http.Error(w, fmt.Sprintf("Can't process entity: %s",err.Error()), 422)
        return
      }
      
      if err:= c.Collect(ctx);err!=nil{
        http.Error(w, fmt.Sprintf("Broken link: %s",err.Error()), 422)
        return
      }
      // inconsistent here, the value will be updated later
      c.Inc(inc)
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
        
        name :=  strings.Split(k.StringID(), "-")[0]
        _,ok := set[name]

        if !ok {
          c := Counter{name,"0"}
          
          log.Infof(ctx,"Collect %v", c)

          if err := c.Collect(ctx);err!=nil{
            set[name] = err.Error()
          } else {
            set[name] = c.Count
          }
        }
      }

      // map shoud somehow be serialized without this slice
      for k := range set {
        cs = append(cs, Counter{k,set[k]})
      }
      
      w.Header().Set("Content-Type","application-json")
      json.NewEncoder(w).Encode(&cs)
    case "POST", "PUT":
      enqueue(w,r)

    default:
      http.Error(w, fmt.Sprintf("Method not supported %v", r.Method), 405)
  }
}

// entry point
func init() {
  http.HandleFunc("/tasks/",          status)
  http.HandleFunc("/counters/queue",  queue)
  http.HandleFunc("/counters",        list)
  http.HandleFunc("/counter/",        counter)
  http.HandleFunc("/counter",         create)
}
