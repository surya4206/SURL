package api

import(
  "encoding/json"
  "net/http"
)

type user struct{
  user_id int
  user_name string
}

type url struct{
  original_url string
  new_url string
  malicious bool
  creation time
}

type error struct{
  code int
  message string
}

func()
