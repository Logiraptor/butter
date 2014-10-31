

# Butter Manifesto

The following outlines a plan for the design and implementation of Butter, the cohesive set of tools designed by Toast Mobile to facilitate the sort of work we do.


##Target Use Cases
###Server Implementation
- Goals
  - Minimize runtime reflection
  - No Google App Engine Dependencies in butter core code.
    - I don't think this is feasible for datastore.
  - Butter Core Must run on appengine (no unsafe, etc).
  - Support External Routers like Gorilla/Mux
- User Authentication
  - Methods:
    - Local Username / Password Pair
    - Recovery Method (Email)
    - Google Oauth
    - Facebook Oauth
    - Twitter Oauth
  - Requirements:
    - Easy addition of fields to the user model
    - Minimal Setup to start authenticating
    - Difficult / Impossible to write insecure code from outside butter.
    - User can “Own” an entity in the datastore.
- Data Storage
  - Essential Methods:
    - Get entity by key
    - Put entity by key
    - Delete entity by key
    - Batch Put
    - Batch Get
    - Batch Delete
    - Query Builder
      - Filter
      - Order
      - “Owner”
      - Query Fetcher
      - GetN (w/ paging)
      - GetAll
      - Iterator
  - Requirements:
    - Opaque Keys
    - Post-Get / Pre-Save / Pre Delete Hooks
    - Datatype can be inferred from query and fetcher can return it directly
- Data Synchronization
  - Push Notification Support
    - Registration
    - Deregistration
  - Analytics
    - App Usage Statistics
    - Easy Dashboards
    - Stats Over Time
  - Administration
    - User Management
  - Documentation
    - Automatic REST Documentation


Database:

  An ideal implementation for paging might let me write code like this:

```Go
http.Handle("/posts", db.Paged(func(req *http.Request) (*datastore.Query, error){
  user, err := GetUser(req)
  if err != nil {
    return nil, err
  }

  return datastore.NewQuery("Post").Ancestor(db.Key(ancestor)).Order("CreatedAt"), nil
}, Post{})
```
