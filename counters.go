package counters

import (
  "fmt"
  "strings"
  "html/template"
  "net/http"
  "encoding/json"
  "google.golang.org/appengine"
  "google.golang.org/appengine/datastore"
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
      enqueue(w,r)

    default:
      http.Error(w, fmt.Sprintf("Method not supported %v", r.Method), 405)
  }
}

func persist(w http.ResponseWriter, r *http.Request){
  //ctx := appengine.NewContext(r)
}

// entry point
func init() {
  http.HandleFunc("/counters/persist", persist)
  http.HandleFunc("/tasks/",          status)
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
      <body><dl>
        <dt>create counter</dt>
        <dt>curl -i -X POST -H "Content-length:0" https://focus-pottery-153219.appspot.com/counter</dt>
        <dd><pre><code>HTTP/1.1 200 OK
              Content-Type: application-json
              X-Cloud-Trace-Context: 5aab2330e3e9baa31ea787216eb06eee;o=1
              Date: Tue, 27 Dec 2016 15:53:24 GMT
              Server: Google Frontend
              Content-Length: 72
              Alt-Svc: quic=":443"; ma=2592000; v="35,34"

            {"id":"ahZlfmZvY3VzLXBvdHRlcnktMTUzMjE5cgsLEgdDb3VudGVyDA","count":"0"}
          </code></pre>
        </dd>

        <dt>check counter value</dt>
        <dt>curl -i -X GET https://focus-pottery-153219.appspot.com/counter/ahZlfmZvY3VzLXBvdHRlcnktMTUzMjE5cgsLEgdDb3VudGVyDA</dt>
        <dd><pre><code>
            HTTP/1.1 200 OK
            Content-Type: application-json
            X-Cloud-Trace-Context: e8e4c7256e5937bcfa344491a552ef40;o=1
            Date: Tue, 27 Dec 2016 15:56:24 GMT
            Server: Google Frontend
            Content-Length: 72
            Alt-Svc: quic=":443"; ma=2592000; v="35,34"

            {"id":"ahZlfmZvY3VzLXBvdHRlcnktMTUzMjE5cgsLEgdDb3VudGVyDA","count":"0"}
          </code></pre></dd>

          <dt>update counter</dt>
          <dt>curl -i -X POST https://focus-pottery-153219.appspot.com/counter/ahZlfmZvY3VzLXBvdHRlcnktMTUzMjE5cgsLEgdDb3VudGVyDA 
            -d '{"id":"ahZlfmZvY3VzLXBvdHRlcnktMTUzMjE5cgsLEgdDb3VudGVyDA", "count": "12"}'</dt>
          <dd><pre><code>
            HTTP/1.1 200 OK
            Content-Type: application-json
            X-Cloud-Trace-Context: 9984029b8cb1654918d1fc6d9b2a4959;o=1
            Date: Tue, 27 Dec 2016 16:01:47 GMT
            Server: Google Frontend
            Content-Length: 73
            Alt-Svc: quic=":443"; ma=2592000; v="35,34"

            {"id":"ahZlfmZvY3VzLXBvdHRlcnktMTUzMjE5cgsLEgdDb3VudGVyDA","count":"12"}
            </code></pre></dd>

          <dt>enqueue task</dt>
          <dt>curl -i -X POST http://localhost:8080/counters 
            -d '[{"id":"ahZlfmZvY3VzLXBvdHRlcnktMTUzMjE5cgsLEgdDb3VudGVyDA","count":"112"}]'
          </dt>
          <dd>
          <pre><code>
            HTTP/1.1 200 OK
            content-type: application-json
            Cache-Control: no-cache
            Expires: Fri, 01 Jan 1990 00:00:00 GMT
            Content-Length: 90
            Server: Development/2.0
            Date: Tue, 27 Dec 2016 16:06:57 GMT

            {"id":"ahhkZXZ-Zm9jdXMtcG90dGVyeS0xNTMyMTlyDwsSBFRhc2siBXRhc2sxDA","count":"in_progress"}
          </code></pre>
          </dd>

      </dl></body>
    </html>`)).Execute(w, map[string]interface{}{}); err!=nil{
    http.Error(w, err.Error(), 500)
  }
}
