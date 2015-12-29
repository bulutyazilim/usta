package main

import (
  "fmt"
  "net/http"
  "log"
  "github.com/jmervine/exec"
  "os"
  "encoding/json"
  "bytes"
)
type SlackResponse struct {
  ResponseType string `json:"response_type"`
  Text string `json:"text"`
}
const slack_token="-"
const working_dir=""
const project_name="-"
const assetic_dir=""
var err error
var response SlackResponse
var out string
func main(){
  http.HandleFunc("/",func(w http.ResponseWriter, req *http.Request){
      token := req.FormValue("token")
      //team_id := req.FormValue("team_id")
      //team_domain := req.FormValue("team_domain")
      //channel_id := req.FormValue("channel_id")
      //channel_name := req.FormValue("channel_name")
      //user_id := req.FormValue("user_id")
     // user_name := req.FormValue("user_name")
     // text := req.FormValue("text")
     response_url := req.FormValue("response_url")

      //fmt.Fprintf(w, "%s %s %s %s %s %s %s %s %s",token,team_id,team_domain,channel_id,channel_name,user_id,user_name,text,response_url)


      //fmt.Fprintf(w,token)

      if token!=slack_token {
        fmt.Fprintf(w,"Access denied!")
	      return
      }
      go deploy(response_url)
      fmt.Fprintf(w,"Deploy started");
  })

  fmt.Println("Listening on 2523");
  err = http.ListenAndServe(":2523",nil)
   if err != nil {
   fmt.Println(err)
   }
}
func deploy(response_url string){

  out,err = git(response_url)
  if err!=nil {
    log.Fatal(err)
    return
  }

  out,err = composer(response_url)
  if err!=nil {
    log.Fatal(err)
    return
  }



  out,err = schemaupdate(response_url)
  if err!=nil {
    log.Fatal(err)
    return
  }

  out,err=assetic(response_url)
  if err!=nil {
    log.Fatal(err)
    return
  }


  out,err=cacheclear(response_url)
  if err!=nil {
    log.Fatal(err)
    return
  }
  out,err=permission(response_url)
  if err!=nil {
    log.Fatal(err)
    return
  }
 
}
func composer(response_url string) (out string, err error) {
  return work(response_url,"composer","--working-dir="+working_dir,"update")
}
func git(response_url string) (out string, err error){
  return work(response_url,"git","--git-dir="+working_dir+"/.git", "--work-tree="+working_dir,"pull","origin","master")
}
func assetic(response_url string)(out string, err error){
  return work(response_url,"php",working_dir+"/app/console", "a:d","--env=prod",assetic_dir)
}
func schemaupdate(response_url string)(out string, err error){
  return work(response_url,"php",working_dir+"/app/console", "d:s:u", "--force","--env=prod")
}
func cacheclear(response_url string)(out string, err error){
  return work(response_url,"php",working_dir+"/app/console","cache:clear","--env=prod")
}
func permission(response_url string)(out string, err error){
  return work(response_url,"chmod","-R","777",working_dir+"/app/cache")
}
func work(response_url string,command string, args ... string)(out string,err error){
  var byteout []byte
  byteout,err = exec.ExecTee(os.Stdout,command, args...)
  out = string(byteout)
  if err!=nil {
    fmt.Println(err)
    return
  }
  response.ResponseType="in_channel"
  response.Text = out
  result , errs := json.Marshal(response)
  if errs!=nil {
    log.Fatal(errs)
  }
  sendBack(response_url, result)
  fmt.Println(result)
  return out,err
}
func sendBack(url string, message []byte){
   fmt.Println("Sending status")
   req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))
   req.Header.Set("Content-Type", "application/json")
   client := &http.Client{}
   resp, err := client.Do(req)
   if err != nil {
       panic(err)
   }
   defer resp.Body.Close()
}
