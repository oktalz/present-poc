# present

# ![present](assets/go-mic.png "present")

tool for viewing presentations written in markdown like format

slides are written in text friendly format and follow all standard
markdown rules with some additions

main purpose of application is to view presentation in browsers

## installation
Use the following command to download and install this tool:
```sh
go install github.com/oktalz/present@latest
```

### binaries
  prebuilt binaries can be found on [releases](https://github.com/oktalz/present/releases) page

## example

- enter examples folder, type `present`
  - program should read all files and start web server on port 8080
    - port can be customized (see `.env` file)

## sharing presentations

- enter presentation folder
- type `present zip` - present.tar.gz will be created
- send file
  - user can start presentation with `present present.tar.gz`
  - user can unpack file, enter folder and execute `present`

## customizations && security
- `.env` file or corresponding variables can be used to customize behavior
  ```txt
  ADMIN_PWD=AdminPassword123
  USER_PWD=user
  ASK_USERNAME=false
  PORT=8080
  NEXT_PAGE=ArrowRight,ArrowDown,PageDown,Space,e
  PREVIOUS_PAGE=ArrowLeft,ArrowUp,PageUp
  TERMINAL_CAST=r,b
  TERMINAL_CLOSE=c
  MENU=m
  ```
- if `ADMIN_PWD` is set, only users authorised with that password can execute the code
  - if `ADMIN_PWD` is not provided, it will be generated and written on console
    - with `ADMIN_PWD_DISABLE=true` you can remove need for admin password
- if `USER_PWD` is set, all 'watchers' will need to enter password to see the presentation
- rest are pretty self explanatory (also in examples are defaults for all options)

## options

- please download present tool enter examples folder and run present tool,
  presentation itself explains all options (with most parts visible)
