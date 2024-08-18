# (W)eb (D)evelopment-(40)
[![Go Report Card](https://goreportcard.com/badge/github.com/baalimago/wd-40)](https://goreportcard.com/report/github.com/baalimago/wd-40)
[![wakatime](https://wakatime.com/badge/user/018cc8d2-3fd9-47ef-81dc-e4ad645d5f34/project/3bc921ec-dc23-4222-bf00-578f2eda0cbd.svg)](https://wakatime.com/badge/user/018cc8d2-3fd9-47ef-81dc-e4ad645d5f34/project/3bc921ec-dc23-4222-bf00-578f2eda0cbd)

Test coverage:

This is a static webserver which hot-reloads your web-browser on any local filechanges.

## Usage
`wd-40 s|serve <relative directory>` or `wd-40 s|serve`

## Architecture
1. First the content of the website is copied to a temporary directory
1. At every file, the MIME type is inspected, if it's text/html, a `delta-streamer.js` script is injected
1. The web server is started, hosting the _mirrored_ content
1. The `delta-streamer.js` in turn sets up a websocket connection to wd-40
1. The original file system is monitored, on any file changes:
  1. the new file is copied to the mirror
  1. the file name is propagated to the browser via the websocket
  1. if the browser's origin matches the recently updated file, the browser is told to reload via javascript

```
       ┌───────────────┐                                                 
       │ Web Developer │                                                 
       └───────┬───────┘                                                 
               │                                                         
       [writes <content>]                                                
               │                                                         
               ▼                                                         
 ┌─────────────────────────────┐        ┌─────────────────────┐          
 │ website-directory/<content> │        │ file system notify  │          
 └─────────────┬───────────────┘        └─────────┬───────────┘          
               │                                  │                      
               │                      [update mirrored content]          
               ▼                                  │                      
     ┌────────────────────┐                       │                      
     │ ws-script injector │◄──────────────────────┘                      
     └─────────┬──────────┘                                              
               │                                                         
               │                                                         
               ▼                                                         
   ┌────────────────────────┐                                            
   │ tmp-abcd1234/<content> │                                            
   └───────────┬────────────┘                                            
               │                                                         
       [serves <content>]                                                
               │                               ┌────────────────────────┐
               ▼                               │         Browser        │
┌──────────────────────────────┐               │                        │
│          Web Server          │               │  ┌────┐  ┌───────────┐ │
│ [localhost:<port>/<content>] │◄───[reload────┼─►│ ws │  │ <content> │ │
└──────────────────────────────┘     page]     │  └────┘  └───────────┘ │
                                               │                        │
                                               └────────────────────────┘
```
