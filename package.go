package counters

import(
  "google.golang.org/appengine/datastore"
)

/*
 * Entity to be stored (counters,tasks).
 * rename to counter when tasks idea will be declined
 * `datastore:-` avoid from storage
 */
type Entity struct {
  Payload string `datastore:"count"`
}

/* 
 * Prepare the structure for service response.
 * Can't check counter for nil here, 'couse of var not used in struct mapping...
 */
func Jsonf(key *datastore.Key, c *Entity)(struct{Id string `json:"id"`;Value string `json:"count"`}) {
  return struct {Id string `json:"id"`;Value string `json:"count"`}{key.Encode(),c.Payload,}
}

func Entityf() {}
