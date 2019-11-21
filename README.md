# Gameye Igniter
The igniter was developed as an internal tool at Gameye to help managing the game server process running inside a container.

Currently, it can help you doing the following tasks:

- Start the game server executable with dynamic parameters.
- Keep track of the game server state based on log output and literal or regex line matching, e.g. idle, warmup, playing, end.
- Perform certain actions when a change of state occurs.
- Write configuration files with dynamic parameters to disk before the game servers is started.

## Manual
You can find the manual and how to get started over [here](manual.md).
