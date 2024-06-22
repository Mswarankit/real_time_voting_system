real_time_voting_system/
|-- internal/
|   |-- auth/
|   |   |-- auth.go
|   |-- |-- auth.proto
|   |-- storage/
|   |   |-- storage.go
|   |-- websocket/
|       |-- websocket.go
|-- main.go
|-- go.mod


-- This is project structure

-- go.mod have require file which needed to run voting system
-- auth file will do authentication using JWT token not added JWT key
-- redis file have necessary connection related items
-- websocket will work with voters and taking exact identity.

