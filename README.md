# gshub-server-api

## Cycle System

### Requirements

_Must Have_

1. The main API must store the initial amount of cycles when a user creates an instance.
2. The instance API must request all available cycles from the main API when it starts.
3. The instance API must burn cycles while running, with 1 cycle equating to 1 second.
4. The instance API must inform the main API of the number of burned cycles upon shutdown.
5. If the instance cannot request cycles from the main API on startup, it must shut down.
6. If the instance cannot inform the main API of burned cycles on shutdown, it must save the information in a local SQLite database and send it on the next startup.
7. The system must be fault-tolerant in handling cycle requests and usage reports.

_Should Have_

1. Pricing for cycles must be based on the plan selected for the instance.

_Could Have_

1. Future provision for adding alerting systems when instances are running low on cycles.

### Algorithms

==== Requesting Cycles on Startup

1. Instance API sends a request to the Main API for available cycles.
2. Main API checks the `remaining_cycles` for the instance.
3. If cycles are available, Main API responds with the number of available cycles.
4. If cycles are not available or the Main API is unreachable, the instance shuts down.

==== Burning Cycles

1. The Instance API decreases the available cycles every second.
2. If the available cycles reach zero, the Instance API shuts down the instance and reports the burned cycles to the Main API.

==== Reporting Burned Cycles on Shutdown

1. Instance API sends a report to the Main API with the number of burned cycles.
2. If the Main API is unreachable, the Instance API saves the burned cycles information in the local SQLite database.
3. On the next startup, the Instance API attempts to resend the saved burned cycles information from the SQLite database to the Main API.

=== Fault Tolerance

1. Instance API must handle network failures gracefully by retrying requests to the Main API with exponential backoff.
2. Local SQLite database on the instance ensures that burned cycles are not lost if the instance shuts down unexpectedly or cannot reach the Main API.
