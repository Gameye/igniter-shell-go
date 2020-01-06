# Igniter Integration Guide

## What is the Igniter?
Igniter is the tool we use to parse .yaml files through to docker. This allows us to spin up game servers within Docker while only needing 2 .yaml config files with the game server specification inside.

## How is it used?
When we start up a docker container on a server, we pull a file onto the machine which contains the information that the container needs to start and run. Within this file is a command which tells the igniter tool to launch and also points it to the location of the config.yaml file that it should start pulling info from in order to set up the server. 

## How does it work?
In order to understand how the igniter tool works, it’s important to know how the config and argument files are laid out and the different parts of these files. 

## Entrypoint.sh 
For every game hosted on our system, an Entrypont.sh file is needed. This file is what the Docker container first reads in order to know what directory to launch the game server executables from. Below is an example of the contents of the Entrypoint.sh file and how it tells the igniter tool when to launch and what directory to launch from. In this case, your file should be a copy of the one below, with the directory of your game executable added in.  

```#!/bin/sh

set -e

CONFIG="$1"
shift

set +e # continue on error

/usr/local/bin/igniter-shell launch --config-file /home/steam/config/$CONFIG.yaml --emulate-tty "$@" -- /home/steam/GAMEDIRECTORY/GAMEDIRECTORY/GAMEEXECUTABLE
```

## Config files
All config files are laid out in the same way.

1. "Cmd", this is used to specify the arguments the server parses when it starts up. These are parsed on the command line when the server’s .exe is being run. 
2. The second section is “script”, which is where specific things that happen when the server is being used is parsed back to our API. This section is also used for timing and regex checks. Finally, this section has a long list of various config settings that the server will use. Things like player limits, how many bots are enabled and what map to load are all listed here. 
3. The final section to the config file is “defaults”. This is where specific arguments that require the end-user to enter will have their defaults set. For example, if you allow the end-user to choose how many players are allowed on the server, but they do not fill it in, the server will then use the value listed in the defaults section. 

## Arg files
Arg files are very similar to config files in terms of their layout and what they are used for, however, there are some very important differences. 

The “port” section. This is used by the Igniter tool to set port types on the container. As ports on docker are closed by default, we have to specify which ones need to be in use by the game server. 
Finally, we have the “arg” section. This is used to specify variables that the server uses which can be set via the API.
Opening ports
One of the first things that is done in the Arg file is the opening of ports. Below are 3 examples on how to open UDP, TCP or both ports. All ports are unique and are assigned automatically by our system using the Ephemeral port range. For more examples of layout, see the arg file example at the bottom of this guide. 

```UDP
port:
  - name: game
    protocol: udp

TCP
port:
  - name: game
    protocol: tcp

UDP & TCP
port:
  - name: game
    protocol: any
```

## Server states
When a server is running, it will have different states which are sent back to the API. These states are important as they allow us to know what the server is doing. Examples of these states are “Idle”, “Configure” and “playing”., however, you can create and execute your own states depending on what game you’re running and what functionality it needs. We are using an example game states below as they give a good idea of how the system works and what you can use it for. 

The states need to be put into the config.yaml file in the order that they would be executed by the server. The formatting needs to be correct so that the igniter tool is able to read and send it over to the API.
Idle

The example below is important as it will allow the server to be in an idle state until something happens. As you can see below, the idle state has 3 different outcomes or “nextState” - Depending on what happens on the server, defined by the idle type, the system will move onto the next corresponding state. 
1. Type = This is what the system would be looking for in order to move over to the next state. Regex and timer are examples shown below. 
2. Pattern = This is the regex string that needs to be accepted by the system in order to move into the next state
3. ignoreCase = Is the regex case sensitive or not
4. nextState = What state should be executed next
5. Interval = Time-based in milliseconds. 900,000 milliseconds = 15 minutes

```idle:
  events:
   - type: regex
   pattern: '^exec:\s+couldn''t\s+exec\s+gamemode_\w+_server\.cfg$'
   ignoreCase: true
   nextState: configure
   - type: regex
pattern:>-^L\s+\d{2}\/\d{2}\/\d{4}\s+\-\s+\d{2}:\d{2}:\d{2}:\s+World\s+triggered\s+"Match_Start"
   ignoreCase: true
   nextState: playing
   - type: timer
   interval: 900000 # 15 minutes
   nextState: quit
```   
   
### Configure
The “configure” state is used by the system in order to initiate the loading of the settings the server is going to use. When looking at the config.yaml file, the section that comes under “script” and also under the server states is the configuration section. 
The configure state is really simple. Once the server has transitioned to it, it will load the configuration settings, then move straight back to the idle state. 

```configure:
  events:
   - type: literal
   value: Configure ready...
   nextState: idle
```

### Playing
The playing state is pretty self-explanatory. This state is used to tell that the server currently has a game that is being played. As you can see, there are two events in the example below. One for the game going into intermission (Paused) or one where the game loops back into the idle state due to there not being enough players on each team. 

```playing:
  events:
   - type: literal
    value: Going to intermission...
    ignoreCase: true
    nextState: end
    - type: literal
     value: Game will not start until both teams have players.
     nextState: idle
```
     
### End
The end state is also self-explanatory. It is used when the game comes to an end in order to tell the ignitor that the session is over and should be shut down. There is a 10 second timer enabled so that players aren’t instantly kicked the moment the game ends. 

```end:
   events:
   - type: timer
   interval: 10000 # 10 seconds
   nextState: quit
```   
   
### Transitions
The transitions section is used when the server states section does something. For example, when the server state is changed to the configure state, the system will “transition” to configure, which in turn will then execute the commands listed below it. 
Idle

As listed above in the idle server stats section, this is a state that the server is in when it’s waiting for something to happen. The example below shows the transition for idle. All that is done in this case is the system writes “Idle…” to the console. 

```- to: idle
  command: |
  echo "Idle..."
```  
  
### Configure
As mentioned previously in the server states part of this document, the configure section is used to execute commands on the server as if it was a stand-alone configuration file.  

The section below is activated when the server state is set to configure. It displays in the console “Configure...”, then starts executing line by line, the server config lines like “log on” and “bot_quota”. Finally, once completed, it then prints the message “Bots only config loaded”  in-game a well as displaying “configure ready” in the console. 

```- to: configure
  command: >
  echo "Configure..."

  log on                // Enable server logging.
bot_quota "10"        // Determines the total number of bots in the game.
  say ">Bots Only Config Loaded"
  echo "Configure ready..."
```

### Playing
Like the idle state, the below transition writes to the console “Playing....”

```- to: playing
  command: |
  echo "Playing..."
```  
  
### End
The end transition is used to write to the console that the session is ending as well as display in-game that the match has ended and that it will shut down in 10 seconds time. 

```- to: end
  command: |
  echo "End..."
  say "The match has ended, thank you for playing!"
  say "This server will shut down in 10 seconds"
```  
  
### Quit
The quit state is meant to terminate the game server. In this example it's write the quit command to the game server to tell it to stop.

```- to: quit
  command: |
  echo "Quit..."
  Quit
```  
  
### Error
The error state is meant to catch cases were the game server didn't start up correctly. An example for which we use it, is to pick up invalid login tokens and terminate the inaccessible server.

```- to: error
  command: |
  echo "Error..."
  quit
```  
  
## Extra Igniter tool features

### Writing config files to disk
Your game may need to pass information to a config file stored on disk. With the igniter tool, this is quick and easy.

Within the config.yaml file, you’ll need to include a “files” section. This will allow you to specify a file path location as well as what information you wish to write to it. Below are some examples from a CSGO server setup.

```files:
  - path: csgo/motd.txt
    content: >
      ${arg.motd}
  - path: csgo/match.cfg
    content: >
      "Match"
      {
        "matchid"     "example_match"
        "num_maps"    "1" // Must be an odd number or 2. 1->Bo1, 2->Bo2, 3->Bo3, etc.

        "spectators" // players allowed in spectator (e.g., admins) should go here
        {
          "players"
          {
            "STEAM_1:1:....."   ""
            "STEAM_1:1:....."   ""
            "STEAM_1:1:....."   ""
          }
        }
```

As you can see, there is a “path” variable and a “content” variable. The path is used to specify the config location that you wish to write to and the content is what is being written to that config. As you can see, you can use a mixture of $ arguments as well as “hard coded” variables (Variables that can’t be changed other than by what is written in the content section) 

### Environment Variables
Normally, the variables in the config and arg files are limited and can only be used in them by the game. If you need to use a variable that does something outside of the main config or arg files, then an environmental variable can be used. In the example below, we are able to set a file name for a CSGO demo recording using an environment variable. The reason for this is that the game will create a demo file which is outside of the games normal files. 
When setting environment variables, it is important that the variable is set at the very top of the config or arg file that uses it. 

```env:
  DEM_FILENAME: '${arg.demfilename}'
```

### Timer type
The igniter tool has the ability to use a timer. This is so that we can keep a specific state running for set times, or wait until the timer is over before transitioning to a new state. Using the timer is easy as it just requires its own “event”, a specified amount of time and then the next state that it should transition to. 

```- type: timer
  interval: 10000 # 10 seconds
  nextState: 
```
  
### Regex
As you have seen in some of the examples above, the igniter tool can use regex. The system understands literal characters as well as special characters. It is always advised that you use a regex checker with some example strings as this limits the chances of errors.
Example of complete Config and Arg files
Below is an example of a complete working Config file as well as an Arg file. They show the flow and structure that these files need to be written in.

### Config
```cmd:
  - -Port=${port.game}
  - -QueryPort=${port.query}
    
script:
  initialState: idle
  states:
    idle:
      events:
        - type: regex
          pattern: ''
          ignoreCase: true
          nextState: configure
        - type: timer
          interval: 900000 # 15 minutes
          nextState: end
    configure:
      events:
        - type: literal
          value: Configure ready...
          nextState: idle
    playing:
      events:
        - type: literal
          value: Game will not start until both teams have players.
          nextState: idle
    end:
      events:
        - type: timer
          interval: 60000 # 60 seconds
          nextState: quit
  transitions:
    - to: configure
      command: >
        echo "Configure..."
        echo "Configure ready..."
    - to: idle
      command: |
        echo "Idle..."
    - to: playing
      command: |
        echo "Playing..."
    - to: end
      command: |
        echo "End..."
    - to: quit
      command: |
        echo "Quit..."
defaults:
  port.game: ''
  port.query: ''

Arg file
port:
  - name: game
    protocol: udp
  - name: query
    protocol: udp
arg:
  - name: platform
    type: string
    defaultValue: 'steam'
 
